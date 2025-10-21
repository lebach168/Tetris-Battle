package game

import (
	"errors"
	"fmt"
	"log"
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
	input        InputBuffer // map fat pointer
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
	firstState := NewPlayerState(CreateEmptyBoard(), 0, Tetromino[list[0]], 0, 0, 4, true, DROPSPEED, make(InputBuffer), 0, false)
	g.state.frames[playerId].data[0] = firstState

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
	if g.state.curFrame+3 < g.gl.serverFrame {
		g.state.curFrame = g.gl.serverFrame
	}

	g.state.mu.Unlock()

	//update serverFrame depends on state.currentFrame
	err := g.state.computeBatchFrames(frameQueue.headFrame+1, g.state.curFrame, playerId, g.gl.tick) //
	if err == ErrGameOver {
		g.Stop()
		//TODO send game over message
	}
	if g.gl.serverFrame%3 == 0 {
		msg := NewMessage("opponent") //temp type
		var packet Packet
		ps, err := frameQueue.PlayerStateOfFrame(frameQueue.headFrame)
		if err != nil || ps == nil {
			log.Printf("something wrong with PlayerState %s\n", err.Error())
		}
		msg.Payload.BoardState = BoardState{ps.board, ps.block.shape, ps.cRow, ps.cCol}
		msg.Payload.LatestFrame = frameQueue.headFrame

		packet.body = MarshalMessage(msg)
		broadcast <- packet
	}

}

func (state *GameState) computeBatchFrames(fromFrame, toFrame int, playerId string, tick time.Duration) error {

	f := state.frames[playerId]
	if fromFrame > toFrame {
		return nil
	}

	for i := fromFrame; i <= toFrame; i++ {
		if i == 0 {
			continue
		}
		//TODO log all frames input here for deterministic replay
		ps, _ := f.PlayerStateOfFrame(i)
		previous, _ := f.PlayerStateOfFrame(i - 1)
		propagateState(previous, ps)
		//clean computed frame input
		for k := range previous.input {
			delete(previous.input, k)
		}
		//
		f.head++
		f.headFrame++
		//check game over
		if CheckGameOver(ps.board, ps.block.shape) {
			return errors.New("game over")
		}
		//apply input then check gravity drop
		input := ps.input

		if len(input) > 0 {
			if input[spacebar] {
				landingRow := FindLandingPosition(ps.board, ps.block.shape, ps.cRow, ps.cCol)
				placeBlock(state, ps, ps.block.shape, landingRow, ps.cCol)
				ps.accumulator = 0
				continue

			}
			if input[left] {
				ps.cCol--
				if hasCollision(ps.board, ps.block.shape, ps.cRow, ps.cCol) {
					ps.cCol++ // Revert
				}
			}
			if input[right] {
				ps.cCol++
				if hasCollision(ps.board, ps.block.shape, ps.cRow, ps.cCol) {
					ps.cCol-- // Revert
				}
			}

			// Rotate (rotate clockwise, rrotate counterclockwise)
			if input[rotate] || input[rrotate] {

				rotatedShape := copySlice(ps.block.shape)
				x := 1
				if ok := input[rotate]; ok {
					rotatedShape = RotateRight(rotatedShape)
				} else {
					rotatedShape = RotateLeft(rotatedShape)
					x = 3 //3 = -1 in modula
				}
				form := ps.block.form
				newForm := (form + x) % 4
				if !hasCollision(ps.board, rotatedShape, ps.cRow, ps.cCol) {
					ps.block.shape = rotatedShape
					ps.block.form = newForm

				} else {
					wallkickOffset := GetWallKickData(len(ps.block.shape), form, newForm)
					for _, d := range wallkickOffset {
						newRow := ps.cRow + d[1]
						newCol := ps.cCol + d[0]
						if !hasCollision(ps.board, rotatedShape, newRow, newCol) {
							fmt.Printf("Applying wall kick: dx=%v, dy=%v\n", d[0], d[1])
							ps.cRow = newRow
							ps.cCol = newCol
							ps.block.shape = rotatedShape
							ps.block.form = newForm
							break
						}
					}
				}

			}

			// Vertical moves (down soft drop key down set dropspeed -> 100ms, space hard drop)
			if input[down] {
				landingRow := FindLandingPosition(ps.board, ps.block.shape, ps.cRow, ps.cCol)
				if ps.cRow < landingRow {
					ps.dropSpeed = SOFT_DROP
				}

			}
			if input[downOff] {
				if ps.dropSpeed == SOFT_DROP {
					ps.dropSpeed = DROPSPEED
				}

			}

			// Hold
			if input[hold] && ps.canHold {
				if ps.holdBlock == 0 {
					ps.holdBlock = state.listBlock[ps.blockIndex]
					ps.blockIndex++
					ps.block = Tetromino[state.listBlock[ps.blockIndex]]

				} else {
					ps.block = Tetromino[ps.holdBlock]
					ps.holdBlock = state.listBlock[ps.blockIndex]
				}
				ps.cRow = 0
				ps.cCol = 4
				ps.canHold = false
			}
		}

		//apply gravity drop
		interval := float64(time.Second.Milliseconds()) / float64(tick)
		ps.accumulator += interval

		if ps.accumulator >= float64(ps.dropSpeed) {
			if ps.isCommitting {
				//check
				landingRow := FindLandingPosition(ps.board, ps.block.shape, ps.cRow, ps.cCol)
				if ps.cRow < landingRow {
					ps.isCommitting = false
					ps.dropSpeed = DROPSPEED
				} else {
					placeBlock(state, ps, ps.block.shape, ps.cRow, ps.cCol)
					ps.accumulator = 0 //reset accumulator after commit block

					continue
				}

			} else {
				ps.accumulator -= float64(ps.dropSpeed)
				landingRow := FindLandingPosition(ps.board, ps.block.shape, ps.cRow, ps.cCol)

				if ps.cRow >= landingRow {
					ps.isCommitting = true
					ps.dropSpeed = LOCKDELAY

				} else { //if no conditions are satisfied -> drop piece\
					ps.cRow++
				}
			}
		}
	}

	return nil
}

// record inputs store input event correspond serverFrame # and  server
func (state *GameState) recordInputs(playerId string, inputs []Input, latestFrame int) {
	//ghi nhận lại input và kể cả serverFrame ko có input của client => để server biết được cần phải update state tới serverFrame
	//nào
	state.mu.Lock()
	fqueue := state.frames[playerId]
	for _, input := range inputs {
		state.inputCounter++
		fmt.Println(state.inputCounter)
		frame := input.Frame
		key_event := input.Key
		ps, err := fqueue.PlayerStateOfFrame(frame)
		if err != nil {
			log.Printf("fail to get serverFrame:%d \n", frame)
			return
		}
		if ps == nil {
			log.Printf("serverFrame %d : nil player state, something went wrong!\n", frame)
			return
		}
		if frame < state.curFrame {
			log.Printf("Input at frame %d arrived late — skippe. Current server frame: %d", frame, state.curFrame)
			continue
		}
		ps.input[key(key_event)] = true

	}
	state.curFrame = latestFrame

	state.mu.Unlock()

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
