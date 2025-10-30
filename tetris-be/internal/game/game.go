package game

import (
	"errors"
	"log"
	"sync"
	"time"
)

type GameState struct {
	frames map[string]*FrameQueue // map key : player id
	/*
		If the gap between curFrame and gl.serverFrame becomes large because the client hasn't sent periodic
		updates—  the server will auto simulate the frames state for that duration.
	*/
	curFrame  int //lastest confirmed serverFrame, if diff from curFrame and gl.serverFrame
	listBlock []int
	mu        sync.Mutex
}

func NewGameState() *GameState {
	return &GameState{
		frames:    make(map[string]*FrameQueue),
		curFrame:  0,
		listBlock: make([]int, 0),
	}
}

type PlayerState struct {
	board [][]int

	blockIndex   int
	block        Block
	holdBlock    int //1->7 convert to Block : Tetromino[int]
	cRow         int
	cCol         int
	canHold      bool
	dropSpeed    float64
	inputBuffer  InputBuffer // map fat pointer
	gravityTimer float64
	lockTimer    float64
	onGround     bool
	/*
	 combo int
	 backToBack string
	 shield int
	*/
}

func NewPlayerState(board [][]int, blockIndex int, block Block, holdBlock int, cRow, cCol int, canHold bool,
	dropSpeed float64, input InputBuffer, accumulator float64, onGround bool) *PlayerState {
	return &PlayerState{
		board:        board,
		blockIndex:   blockIndex,
		block:        block,
		holdBlock:    holdBlock, //0 is empty value
		cRow:         cRow,
		cCol:         cCol,
		canHold:      canHold,
		dropSpeed:    dropSpeed,
		inputBuffer:  input,
		gravityTimer: accumulator,
		onGround:     onGround,
	}
}
func NewDefaultPlayerState() *PlayerState {
	return &PlayerState{
		board:       CreateEmptyBoard(),
		canHold:     true,
		dropSpeed:   DROPSPEED,
		inputBuffer: make(InputBuffer),
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
var ErrOutOfRange = errors.New("out of range")

func NewGame() *Game {
	return &Game{
		state:       NewGameState(),
		isPlaying:   false,
		delayBuffer: 4, //2 frames

	}
}

func (g *Game) Init(playerId string, msg Message, broadcast chan Packet) {
	//init data for game state: list block for player
	g.state.listBlock = GenerateList_7bag(g.state.listBlock, 1000)
	list := g.state.listBlock
	body := NewMessage("start")
	//startTime := time.Now().Add(time.Second * 2).UnixMilli()
	//body.Payload.StartAt = startTime
	//init first state at serverFrame 0
	g.state.InitFrameQueue(playerId)
	body.Payload.ListBlock = list
	var packet Packet
	g.gl = NewGameLoop(playerId, g.onUpdate)

	packet.body = MarshalMessage(body)

	broadcast <- packet

}
func (g *Game) StartGame(playerId string, broadcast chan Packet) {
	list := g.state.listBlock
	firstState := NewPlayerState(CreateEmptyBoard(), 0, Tetromino[list[0]], 0, 0, 4, true, DROPSPEED,
		make(InputBuffer), 0, false)
	g.state.frames[playerId].data[0] = firstState
	g.state.curFrame = 1

	go g.gl.Run(broadcast)
}
func (g *Game) computeDelayBuffer(msg Message, broadcast chan Packet) {
	msg.Type = "ping"
	now := time.Now().UnixMilli()
	var packet Packet
	msg.Timestamp = now
	broadcast <- packet
}
func (g *Game) onUpdate(playerId string, broadcast chan Packet) {
	frameQueue := g.state.frames[playerId]
	//update current serverFrame if client inactive in sending messages
	g.state.mu.Lock()
	if g.state.curFrame+4 < g.gl.serverFrame {
		g.state.curFrame++
	}

	g.state.mu.Unlock()

	//update serverFrame depends on state.curFrame
	err := g.state.computeBatchFrames(frameQueue.pFrame+1, g.state.curFrame, playerId) //
	if err == ErrGameOver {
		g.Stop()
		//TODO send game over message
	}
	if g.gl.serverFrame%3 == 0 {
		msg := NewMessage("opponent") //temp type
		var packet Packet
		ps, err := frameQueue.Get(frameQueue.pFrame)
		if err != nil || ps == nil {
			log.Printf("something wrong with PlayerState %s\n", err.Error())
		}
		msg.Payload.BoardState = BoardState{ps.board, ps.block.shape, ps.cRow, ps.cCol}
		msg.Payload.LatestFrame = frameQueue.pFrame

		packet.body = MarshalMessage(msg)
		broadcast <- packet

	}

}

func (state *GameState) computeBatchFrames(fromFrame, toFrame int, playerId string) error {

	f := state.frames[playerId] // frame queue

	if fromFrame > toFrame {
		return nil
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	for i := fromFrame; i <= toFrame; i++ {

		//TODO log all frames inputBuffer here for deterministic replay
		ps, _ := f.Get(i)
		previous, _ := f.Get(i - 1)
		propagateState(previous, ps)

		//TODO check logic queue head
		//apply inputBuffer
		input := ps.inputBuffer
		if len(input) > 0 {
			applyInputBuffer(state, ps, input)
		}
		landingRow := FindLandingPosition(ps.board, ps.block.shape, ps.cRow, ps.cCol)
		//gravity drop

		if !ps.onGround {
			ps.gravityTimer += INTERVAL
			if ps.gravityTimer >= float64(ps.dropSpeed) {
				if ps.cRow < landingRow {
					ps.cRow++
					ps.gravityTimer -= float64(ps.dropSpeed)
				}
				if ps.cRow >= landingRow {
					ps.onGround = true
				}
			}
		} else {
			ps.lockTimer += INTERVAL
		}
		//commit phase:
		if ps.onGround {
			if ps.cRow < landingRow {
				ps.onGround = false
				ps.lockTimer = 0

			}
			if ps.lockTimer >= LOCKDELAY {
				placeBlock(state, ps, ps.block.shape, landingRow, ps.cCol)
				//clear lines then reset timer spawn new piece(block)
				lines := clearLines(ps.board)
				if lines > 0 {
					//TODO
				}
				spawnNewPiece(state, ps)
				//check game over
				if CheckGameOver(ps.board, ps.block.shape) {
					return errors.New("game over")
				}
			}
		}
		f.Forward()
	}

	return nil
}

// record inputs store inputBuffer event correspond serverFrame # and  server
func (state *GameState) recordInputs(playerId string, inputs []Input, latestFrame int, broadcast chan Packet) {
	//ghi nhận lại inputBuffer và kể cả serverFrame ko có inputBuffer của client => để server biết được cần phải update state tới serverFrame
	//nào
	var packet Packet
	msg := NewMessage("input-server")
	fqueue := state.frames[playerId]
	fqueue.mu.Lock()
	defer fqueue.mu.Unlock()
	for _, input := range inputs {

		serverConfirmedKeys := []string{}
		frame := input.Frame
		keys := input.Keys
		ps, err := fqueue.Get(frame)
		if err != nil {
			if frame < state.curFrame {
				log.Printf("Input at frame %d arrived late — skip. Current server frame: %d", frame, state.curFrame)
			}

			if errors.Is(err, ErrOutOfRange) {
				log.Printf("invalid frame counter:%d something wrong \n", frame)
			}
			//TODO sync clock message here
			return
		}
		if ps == nil {
			log.Printf("frame %d : nil player state, something went wrong!\n", frame)
			return
		}

		ps.inputBuffer = InputBuffer{}
		for _, key_event := range keys {
			ps.inputBuffer[key(key_event)] = true
		}
		for k, v := range ps.inputBuffer {
			if v == true {
				serverConfirmedKeys = append(serverConfirmedKeys, string(k))
			}
		}

		if len(serverConfirmedKeys) > 0 {
			msg.Payload.Inputs = append(msg.Payload.Inputs, Input{Frame: frame, Keys: serverConfirmedKeys})
		}

	}

	state.mu.Lock()
	state.curFrame = latestFrame
	state.mu.Unlock()
	if len(msg.Payload.Inputs) > 0 {
		packet.body = MarshalMessage(msg)
		broadcast <- packet
	}

}
func (state *GameState) InitFrameQueue(playerId string) {
	state.frames[playerId] = NewQueue(QUEUE_SIZE)
}
func (g *Game) Pause() {
	g.gl.pause <- struct{}{}
}
func (g *Game) Unpause() {
	g.gl.resume <- struct{}{}
}
func (g *Game) Stop() {
	g.gl.quit <- struct{}{}
}
