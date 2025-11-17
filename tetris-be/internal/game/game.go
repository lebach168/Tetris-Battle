package game

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

type Game struct {
	players     map[string]*FrameExecutor
	isPlaying   bool
	delayBuffer int //fixed value, refactor later
}
type FrameExecutor struct {
	playerId string
	gl       *GameLoop
	frames   *FrameQueue
	/*
		If the gap between curFrame and gl.tickFrame becomes large because the client hasn't sent periodic
		updates—  the server will auto simulate the frames state for that duration.
	*/
	netFrame  int //last frame received from client
	listBlock []int
	opponentC chan Attack
	mu        sync.Mutex
}

func NewFrameExecutor(playerId string) *FrameExecutor {
	return &FrameExecutor{
		playerId:  playerId,
		frames:    NewQueue(QUEUE_SIZE),
		netFrame:  0,
		listBlock: make([]int, 0),
	}
}

var ErrGameOver = errors.New("game over")
var ErrOutOfRange = errors.New("out of range")

func NewGame() *Game {
	return &Game{
		players:     map[string]*FrameExecutor{},
		isPlaying:   false,
		delayBuffer: 4, //2 frames

	}
}
func (g *Game) Rematch() {
	//TODO gọi lại Init -> start lại match mới
}
func (g *Game) Init(broadcast chan Packet, conns map[string]*PlayerConn, sender string) {
	playerCount := 0
	for playerId, conn := range conns {
		g.players[playerId] = NewFrameExecutor(playerId)
		if conn != nil {
			playerCount++
		}
	}
	if playerCount < 2 {
		var packet Packet
		body := NewMessage("start")
		body.Error = "cannot start"
		packet.body = MarshalMessage(body)
		packet.directId = sender
		broadcast <- packet

		return
	}

	//init data for game state: list block for player
	for pId, exec := range g.players {
		exec.listBlock = GenerateList_7bag(exec.listBlock, 1000)
		exec.gl = NewGameLoop(exec.onUpdate, exec.recordInputs, exec.receiveGarbage)
		list := exec.listBlock
		body := NewMessage("start")
		body.Payload.ListBlock = list
		var packet Packet
		packet.directId = pId

		packet.body = MarshalMessage(body)

		broadcast <- packet
	}
outer:
	for p1, exec := range g.players {
		for p2, exec2 := range g.players {
			if p1 != p2 {
				exec.opponentC = exec2.gl.attacked
				exec2.opponentC = exec.gl.attacked
				break outer
			}
		}
	}
	//startTime := time.Now().Add(time.Second * 2).UnixMilli()
	//body.Payload.StartAt = startTime
	//init first state at tickFrame 0

}
func (g *Game) StartGame(broadcast chan Packet) {
	for _, exec := range g.players {
		list := exec.listBlock
		firstState := NewBoardState(CreateEmptyBoard(), 0, Tetromino[list[0]], 0, 0, 4, true, DROPSPEED,
			make(InputBuffer), 0, false)
		exec.frames.data[0] = firstState
		exec.netFrame = 1
		go exec.gl.Run(broadcast)
	}

}
func (g *Game) computeDelayBuffer(msg Message, broadcast chan Packet) {
	msg.Type = "ping"
	now := time.Now().UnixMilli()
	var packet Packet
	msg.Timestamp = now
	broadcast <- packet
}
func (exec *FrameExecutor) onUpdate(broadcast chan Packet) {
	frameQueue := exec.frames
	//update current tickFrame if client inactive in sending messages

	if exec.netFrame+4 < exec.gl.tickFrame {
		exec.netFrame++
	}
	//update tickFrame depends on state.curFrame
	err := exec.computeBatchFrames(frameQueue.simFrame+1, exec.netFrame, broadcast) //
	if errors.Is(err, ErrGameOver) {
		//todo stop cho cả 2 players
		exec.Stop()
		var packet Packet
		msg := NewMessage("gameover")
		msg.PlayerId = "winner id"
		packet.body = MarshalMessage(msg)
		broadcast <- packet
	}
	if exec.gl.tickFrame%3 == 0 {
		msg := NewMessage("opponent") //temp type
		var packet Packet
		ps, err := frameQueue.Get(frameQueue.simFrame)
		if err != nil || ps == nil {
			log.Printf("something wrong with BoardState %s\n", err.Error())
		}
		msg.Payload.BoardState = BoardStateDTO{ps.board, ps.block.shape, ps.cRow, ps.cCol}
		msg.Payload.LatestFrame = frameQueue.simFrame
		packet.excludeId = exec.playerId
		packet.body = MarshalMessage(msg)
		broadcast <- packet
	}

}

func (exec *FrameExecutor) computeBatchFrames(fromFrame, toFrame int, broadcast chan Packet) error {

	fq := exec.frames // frame queue

	if fromFrame > toFrame {
		return nil
	}

	for frame := fromFrame; frame <= toFrame; frame++ {

		//TODO log all frames inputBuffer here for deterministic replay
		bs, _ := fq.Get(frame)
		previous, _ := fq.Get(frame - 1)
		PropagateState(previous, bs)
		//apply garbage
		//ps.cRow cũng sẽ bị đẩy lên
		garbage := fq.GetGarbage(frame)
		bs.cancel = min(fq.GetCancel(frame)+bs.cancel, 0)
		if garbage > 0 {
			garbage, bs.cancel = max(garbage+bs.cancel, 0), min(bs.cancel+garbage, 0)
			fmt.Printf("[%s] receive %d garbage lines at frame: %d \n", exec.playerId, garbage, frame)
			TakeGarbage(garbage, bs.board)
			bs.cRow = max(1, bs.cRow-garbage)
			if garbage > 0 {
				msg := NewMessage("garbage-sync")
				var packet Packet
				msg.Payload.BoardState = BoardStateDTO{bs.board, bs.block.shape, bs.cRow, bs.cCol}
				msg.Payload.LatestFrame = frame - 1
				packet.directId = exec.playerId
				packet.body = MarshalMessage(msg)
				broadcast <- packet

			}

			if CheckGameOver(bs.board, bs.block.shape, bs.cRow, bs.cCol) {
				return errors.New("game over")
			}
		}
		//apply input
		input := bs.inputBuffer
		hasSpin := input[rotate] || input[rrotate]
		if len(input) > 0 {
			ApplyInputBuffer(exec.listBlock, bs, input)
		}
		//clean Input buffer
		bs.inputBuffer = InputBuffer{}
		landingRow := FindLandingPosition(bs.board, bs.block.shape, bs.cRow, bs.cCol)
		//gravity drop

		if !bs.onGround {
			bs.gravityTimer += INTERVAL
			if bs.gravityTimer >= float64(bs.dropSpeed) {
				if bs.cRow < landingRow {
					bs.cRow++
					bs.gravityTimer -= float64(bs.dropSpeed)
				}
				if bs.cRow >= landingRow {
					bs.onGround = true
				}
			}
		} else {
			bs.lockTimer += INTERVAL
		}
		//commit phase:
		if bs.onGround {
			if bs.cRow < landingRow {
				bs.onGround = false
				bs.lockTimer = 0

			}
			if bs.lockTimer >= LOCKDELAY {
				PlaceBlock(bs.board, bs.block.shape, bs.cRow, bs.cCol)
				//clear lines then reset timer spawn new piece(block)
				lines := ClearLines(bs.board)
				b2bType := "none"
				if hasSpin {
					b2bType = fmt.Sprintf("spin-%t:%d", hasSpin, lines)
				}
				if lines == 4 {
					b2bType = "clear4"
				}

				if bs.b2b != b2bType {
					bs.b2b = b2bType
				}
				if lines > 0 {
					perfect := isPerfect(bs.board)
					b2bFlag := (bs.b2b == b2bType) && (bs.b2b != "none")
					bs.b2b = b2bType
					garbageSent := CalculateGarbageRows(lines, hasSpin, bs.combo, b2bFlag, perfect)
					bs.send += garbageSent
					fq.CancelGarbage(frame, garbageSent)
					bs.combo++
				} else {
					bs.b2b = "none"
					bs.combo = 0
					//TODO combo-end here starting send garbage (delay in 45 frame from this frame )
					// send message to client
					if bs.send > 0 {
						fmt.Printf("[%s] send garbage at frame: %d \n", exec.playerId, frame)
						exec.opponentC <- Attack{lines: bs.send, atFrame: frame}
					}
					bs.send = 0
				}

				SpawnNewPiece(exec.listBlock, bs)
				//check game over
				if CheckGameOver(bs.board, bs.block.shape, bs.cRow, bs.cCol) {
					return errors.New("game over")
				}
			}
		}

		fq.Forward()
	}

	return nil
}

// record inputs store inputBuffer event correspond tickFrame # and  server
func (exec *FrameExecutor) recordInputs(inputs []Input, latestFrame int, broadcast chan Packet) {
	//ghi nhận lại inputBuffer và kể cả tickFrame ko có inputBuffer của client
	//=> để server biết được cần phải update state tới tickFrame nào
	var packet Packet
	msg := NewMessage("input-server")
	fqueue := exec.frames

	for _, input := range inputs {

		serverConfirmedKeys := []string{}
		frame := input.Frame
		keys := input.Keys
		ps, err := fqueue.Get(frame)
		if err != nil {
			if frame < exec.netFrame {
				//log.Printf("Input at frame %d arrived late — skip. Current server frame: %d", frame, exec.netFrame)
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

	exec.mu.Lock()
	exec.netFrame = latestFrame
	exec.mu.Unlock()
	if len(msg.Payload.Inputs) > 0 {
		packet.body = MarshalMessage(msg)
		broadcast <- packet
	}

}
func (exec *FrameExecutor) receiveGarbage(atk Attack) {
	exec.frames.GarbageUpcoming(atk.atFrame, atk.lines)
}
func (g *Game) Pause() {
	for _, exec := range g.players {
		exec.gl.pause <- struct{}{}
	}

}
func (g *Game) Unpause() {
	for _, exec := range g.players {
		exec.gl.resume <- struct{}{}
	}
}
func (exec *FrameExecutor) Stop() {
	exec.gl.quit <- struct{}{}
}
