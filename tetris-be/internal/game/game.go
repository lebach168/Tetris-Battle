package game

import (
	"errors"
	"fmt"
	"time"
)

type PlayerState struct {
	board [][]int

	blockIndex   int
	block        Block
	holdBlock    int //1->7 convert to Block : Tetromino[int]
	cRow         int
	cCol         int
	canHold      bool
	dropSpeed    int
	input        InputBuffer
	accumulator  float64
	isCommitting bool
	/*
	 combo int
	 backToBack string
	 shield int
	*/
}

func NewPlayerState(board [][]int, blockIndex int, block Block, holdBlock int, cRow, cCol int, canHold bool,
	dropSpeed int, input InputBuffer, accumulator float64, isCommiting bool) *PlayerState {
	return &PlayerState{
		board:        board,
		blockIndex:   blockIndex,
		block:        block,
		holdBlock:    holdBlock, //o is empty value
		cRow:         cRow,
		cCol:         cCol,
		canHold:      canHold,
		dropSpeed:    dropSpeed,
		input:        input,
		accumulator:  accumulator,
		isCommitting: isCommiting,
	}
}
func NewDefaultPlayerState() *PlayerState {
	return &PlayerState{
		board:     CreateEmptyBoard(),
		canHold:   true,
		dropSpeed: DROPSPEED,
		input:     make(InputBuffer),
	}
}

type key string

const (
	down     key = "down"
	downOff  key = "downOff"
	left     key = "left"
	right    key = "right"
	rotate   key = "rotate" //arrow up
	rrotate  key = "rrotate"
	spacebar key = "space"
	hold     key = "hold"
)

type InputBuffer map[key]bool
type Game struct {
	gl        *GameLoop
	state     *GameState
	isPlaying bool

	delayBuffer int //fixed value, refactor it later
}

var ErrGameOver = errors.New("game over")

func NewGame() *Game {
	return &Game{
		state:       NewGameState(),
		isPlaying:   false,
		delayBuffer: 10, //2 frames

	}
}

func (g *Game) Init(playerId string, msg Message, broadcast chan Packet) {
	//init data for game state: list block for player
	g.state.listBlock = GenerateList_7bag(g.state.listBlock, 1000)
	list := g.state.listBlock
	body := NewMessage("start")
	//startTime := time.Now().Add(time.Second * 2).UnixMilli()
	//body.Payload.StartAt = startTime
	//init first state at frame 0
	firstState := NewPlayerState(CreateEmptyBoard(), 0, Tetromino[list[0]], 0, 0, 4, true, DROPSPEED, make(InputBuffer), 0, false)

	g.state.frames[playerId].AddNew(firstState)
	body.Payload.ListBlock = list
	var packet Packet
	g.gl = NewGameLoop(playerId, g.onUpdate)

	packet.body = MarshalMessage(body)
	go g.gl.Run(broadcast)
	broadcast <- packet

}
func (g *Game) computeDelayBuffer() {

}
func (g *Game) onUpdate(playerId string, broadcast chan Packet) {
	f := g.state.frames[playerId]

	//add empty state for latest frame
	if f.size < f.cap {
		f.AddNew(NewDefaultPlayerState())
	} else {
		f.Add()
	}
	//apply input from old current frame to (latest - number frames on delay buffer) aka new current frame
	fromFrame := g.state.curF + 1
	toFrame := f.latestFrame - g.delayBuffer
	if toFrame < 0 {
		return
	}

	msg := NewMessage("opponent")
	err := g.state.computeBatchFrames(fromFrame, toFrame, playerId, g.gl.tick)
	if err != nil {
		if errors.Is(err, ErrGameOver) {
			msg.Type = "gameover"
		}
	}

	//send back to p.send
	g.state.curF = toFrame
	if toFrame%3 == 0 { // 3 frames = 100ms mới gửi 1 lần
		ps, err := f.PlayerStateOfFrame(toFrame)
		if err != nil {
			fmt.Println("outdated frame")
			return
		}
		msg.Payload.BoardState = BoardState{Board: ps.board, Block: ps.block.shape, CRow: ps.cRow, CCol: ps.cCol}
		msg.Payload.Frame = toFrame

		broadcast <- Packet{body: MarshalMessage(msg)}

	}

}

func (state *GameState) computeBatchFrames(fromFrame, toFrame int, playerId string, tick time.Duration) error {

	f := state.frames[playerId]
	if fromFrame > toFrame {
		return nil
	}

	for i := fromFrame; i <= toFrame; i++ {
		diff := f.latestFrame - i
		if diff < 0 || diff >= f.cap {
			fmt.Println("Error: Frame", i, "out of queue range for player", playerId)
			continue
		}
		idx := (f.head - diff + f.cap) % f.cap
		ps := f.data[idx]

		if ps == nil {
			fmt.Println("Error: Nil PlayerState at frame", i)
			continue
		}
		previous := f.data[(idx-1+f.cap)%f.cap]
		propagateState(previous, ps)
		input := ps.input
		if len(input) > 0 {
			applyInputBuffer(state, ps, input)
		}
		if input[spacebar] {
			landingRow := FindLandingPosition(ps.board, ps.block.shape, ps.cRow, ps.cCol)
			placeBlock(state, ps, ps.block.shape, landingRow, ps.cCol)
			ps.accumulator = 0
			if CheckGameOver(ps.board, ps.block.shape) {
				return errors.New("game over")
			}

		} else { //apply gravity drop
			interval := float64(time.Second.Milliseconds()) / float64(tick)
			ps.accumulator += interval
			if ps.accumulator >= float64(ps.dropSpeed) {
				if ps.isCommitting {
					placeBlock(state, ps, ps.block.shape, ps.cRow, ps.cCol)
					ps.accumulator = 0 //reset accumulator after commit block
					if CheckGameOver(ps.board, ps.block.shape) {
						return errors.New("game over")
					}
					break
				}
				ps.accumulator -= float64(ps.dropSpeed)
				landingRow := FindLandingPosition(ps.board, ps.block.shape, ps.cRow, ps.cCol)

				if ps.cRow >= landingRow {
					ps.isCommitting = true
					ps.dropSpeed = LOCKDELAY
					break
				}
				//if no conditions are satisfied -> drop piece\
				ps.cRow++
			}
		}
		//clean input
		for k := range input {
			delete(input, k)
		}
	}
	return nil
}
func (state *GameState) RecordInputs(playerId string, inputs []Input) {
	fqueue := state.frames[playerId]
	for _, input := range inputs {
		frame := input.Frame
		key_event := input.Key
		if frame <= state.curF {
			//fmt.Printf("missed %s input in frame:%v \n", key_event, frame)
			continue
		}
		ps, err := fqueue.PlayerStateOfFrame(frame)
		if err != nil {
			fmt.Println("fail to get frame")
		}
		if ps == nil {
			fmt.Printf("frame %d : nil player state, something went wrong!\n", frame)
			return
		}
		ps.input[key(key_event)] = true
	}

}
func (state *GameState) InitFrameQueue(playerId string) {
	state.frames[playerId] = NewQueue(QUEUE_SIZE)
}
