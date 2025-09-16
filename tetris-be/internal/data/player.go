package data

import "github.com/gorilla/websocket"

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
		send: make(chan []byte, 128),
	}
}

func (p PlayerConn) receive() {

}
func (p PlayerConn) broadcast() {

}
