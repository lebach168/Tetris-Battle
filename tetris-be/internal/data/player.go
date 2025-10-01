package data

import (
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
	writeWait      = 6 * time.Second
	pongWait       = 30 * time.Second
	pingPeriod     = 20 * time.Second //must smaller than pong wait
	maxMessageSize = 1024 * 4         //4 kbyte

)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

type Message struct {
	Action  string `json:"action,omitempty"`
	Payload struct {
		Tetrominos []int   `json:"tetrominos,omitempty"`
		Board      [][]int `json:"board,omitempty"`
	} `json:"payload"`
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
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Errorf("player %s disconnected: %v", p.ID, err)
			}
			return
		}
		var send Packet
		//TODO
		//handle message here
		
		//exclude or include message

		p.r.broadcast <- send
	}
}
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
