package game

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net"
	"time"
)

type PlayerConn struct {
	ID string `json:"ID,omitempty"`
	r  *Room
	// The websocket connection.
	conn *websocket.Conn
	// Buffered channel of outbound messages
	send chan []byte
}

func NewPlayerConn(ID string, room *Room, conn *websocket.Conn) *PlayerConn {
	return &PlayerConn{
		ID:   ID,
		r:    room,
		conn: conn,
		send: make(chan []byte, 256),
	}
}

// classifyErr trả về mô tả lỗi thân thiện
func classifyErr(prefix string, err error) string {
	if err == nil {
		return prefix + ": <nil>"
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return fmt.Sprintf("%s: net timeout: %v", prefix, err)
		}
		return fmt.Sprintf("%s: net error: %v", prefix, err)
	}
	if closeErr, ok := err.(*websocket.CloseError); ok {
		return fmt.Sprintf("%s: ws close code=%d text=%s", prefix, closeErr.Code, closeErr.Text)
	}
	return fmt.Sprintf("%s: %T %v", prefix, err, err)
}

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

/*
Message Flow:

	   P.conn.Read()
	      ↓
	   p.conn.ReadMessage()
	      ↓
	   handleMessage() ==> r.Game.receiveInput(serverFrame, input)
	  ==> apply input in gameloop.onUpdate()
	      ↓
	   room.broadcast <- (all or exclude sender)
	      ↓
	   oP.send<-
		  ↓
	   other player.conn.Write()
*/
const (
	//temp value, refactor it later
	writeWait      = 3 * time.Second // receive and send ~15msg/sec
	pongWait       = 30 * time.Second
	pingPeriod     = 25 * time.Second //must smaller than pong wait
	maxMessageSize = 1024 * 10        //10 kbyte
	outboundSize   = 256
)

func (p *PlayerConn) Read() {
	defer func() {
		p.r.leave <- p
		p.conn.Close()
	}()
	p.conn.SetReadLimit(maxMessageSize)
	p.conn.SetReadDeadline(time.Now().Add(pongWait))
	p.conn.SetPongHandler(func(string) error {
		p.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		_, msg, err := p.conn.ReadMessage()
		//fmt.Printf("Server received: %s\n", msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[ws][%s] %s", p.ID, classifyErr("ReadMessage error", err))
			}
			return
		}
		//TODO
		p.handleGameMessage(msg)

	}
}

// // Enqueue gửi message an toàn (không ghi trực tiếp vào conn từ nhiều goroutine)
//
//	func (p *PlayerConnLog) Enqueue(msg []byte) {
//		select {
//		case p.outbound <- msg:
//		default:
//			// backpressure: queue đầy -> log và drop (hoặc kick)
//			log.Printf("[ws][%s] outbound queue full -> dropping message", p.id)
//			return
//		}
//
// Write() k0 nhận message trực tiếp từ p.send, tất cả đều đi qua chan broadcast , broadcast sẽ gửi vào chan p.send
func (p *PlayerConn) Write() {
	ticker := time.NewTicker(pingPeriod) //send ping pong every period duration to simulate heartbeat of connection
	defer func() {
		ticker.Stop()
		p.conn.Close()
	}()
	for {
		select {
		case message, alive := <-p.send:
			if !alive {
				//send channel has closed
				p.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			p.conn.SetWriteDeadline(time.Now().Add(writeWait))
			w, err := p.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Printf("[ws][%s] %s", p.ID, classifyErr("NextWriter error", err))
				return
			}

			if _, err = w.Write(message); err != nil {
				log.Printf("[ws][%s] %s", p.ID, classifyErr("Write error", err))
				w.Close()
				return
			}

			if err := w.Close(); err != nil {
				log.Printf("[ws][%s] %s", p.ID, classifyErr("Close writer error", err))
				return
			}

		case <-ticker.C:
			p.conn.SetWriteDeadline(time.Now().Add(writeWait))

			if err := p.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("[ws][%s] %s", p.ID, classifyErr("Ping write error", err))
				return
			}
		}
	}

}

func (p *PlayerConn) handleGameMessage(raw []byte) {

	var msg Message
	err := json.Unmarshal(raw, &msg)
	if err != nil {
		//
	}

	//exclude message
	switch msg.Type {

	case "input":
		p.r.game.state.recordInputs(p.ID, msg.Payload.Inputs, msg.Payload.LatestFrame)

	case "ping":
		p.r.game.computeDelayBuffer(msg, p.r.broadcast)
	case "start":
		p.r.game.StartGame(p.ID, p.r.broadcast)
	case "ready":
		p.r.game.Init(p.ID, msg, p.r.broadcast)
	case "pause":
		p.r.game.Pause()
	case "unpause":
		p.r.game.Unpause()
	default:
		log.Printf("unkown type message: %v \n", msg)
		break
	}

}

type BoardState struct {
	Board [][]int `json:"board,omitempty"`
	Block [][]int `json:"block,omitempty"`
	CRow  int     `json:"cRow"`
	CCol  int     `json:"cCol"`
}
type Input struct {
	Key   string
	Frame int
}
type Message struct {
	Type     string `json:"type"`
	PlayerId string `json:"playerid,omitempty"`
	Payload  struct {
		LatestFrame int        `json:"latestFrame,omitempty"`
		ListBlock   []int      `json:"listBlock,omitempty"`
		BoardState  BoardState `json:"state,omitempty"`
		Inputs      []Input    `json:"inputs,omitempty"`
		StartAt     int64      `json:"startAt,omitempty"`
	} `json:"payload"`
	Timestamp int64 `json:"timestamp"`
}

func NewMessage(t string) Message {
	return Message{
		Type: t,
		//Timestamp: time.Now().UnixMilli(),
	}
}
func MarshalMessage(msg Message) []byte {
	res, err := json.Marshal(msg)
	if err != nil {
		fmt.Errorf("marshal error: %v  messge: {%+v}", err, msg)
	}
	return res
}

func (p *PlayerConn) GetConn() *websocket.Conn {
	return p.conn
}
func (p *PlayerConn) SetConn(conn *websocket.Conn) {
	p.conn = conn
}
func (p *PlayerConn) GetRoomID() string {
	return p.r.ID
}
