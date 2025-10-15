package game

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
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

const (
	//temp value, refactor it later
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = 50 * time.Second //must smaller than pong wait
	maxMessageSize = 1024 * 4         //4 kbyte

)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

type BoardState struct {
	Board [][]int `json:"board,omitempty"`
	Block [][]int `json:"block,omitempty"`
	CRow  int     `json:"cRow,omitempty"`
	CCol  int     `json:"cCol,omitempty"`
}
type Input struct {
	Key   string
	Frame int
}
type Message struct {
	Type     string `json:"type"`
	PlayerId string `json:"playerid,omitempty"`
	Payload  struct {
		Frame      int        `json:"frame,omitempty"`
		ListBlock  []int      `json:"listBlock,omitempty"`
		BoardState BoardState `json:"state,omitempty"`
		Inputs     []Input    `json:"inputs,omitempty"`
		StartAt    int64      `json:"startAt,omitempty"`
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
		fmt.Errorf("marshal error ", err)
	}
	return res
}
func NewPlayerConn(ID string, room *Room, conn *websocket.Conn) *PlayerConn {
	return &PlayerConn{
		ID:   ID,
		r:    room,
		conn: conn,
		send: make(chan []byte, 128),
	}
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

/*
Message Flow:

   P.conn.Read()
      ↓
   p.conn.ReadMessage()
      ↓
   handleMessage() ==> r.Game.receiveInput(frame, input)
  ==> apply input in gameloop.onUpdate()
      ↓
   room.broadcast <- (all or exclude sender)
      ↓
   oP.send<-
	  ↓
   other player.conn.Write()

*/

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
				fmt.Errorf("player %s disconnected: %v", p.ID, err)
			}
			return
		}
		//TODO
		p.handleGameMessage(msg)

	}
}

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
			w, err := p.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)
			//TODO refactor it follow the game business
			//Drain the channel queue and flush all messages in one write.
			n := len(p.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-p.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			p.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := p.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
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
		p.r.game.state.RecordInputs(p.ID, msg.Payload.Inputs)
	case "":
	case "start":
		p.r.game.Init(p.ID, msg, p.r.broadcast)
	default:
		fmt.Printf("%v \n", msg)
		break
	}

}
