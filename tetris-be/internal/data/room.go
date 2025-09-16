package data

import (
	"crypto/rand"
	"math/big"
)

const letters = "abcdefghijklmnopqrstxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type Room struct {
	ID        string
	Key       string
	Players   map[*PlayerConn]bool
	join      chan *PlayerConn
	leave     chan *PlayerConn
	broadcast chan Message
	stop      chan struct{}
}

type Message struct {
	exclude string
	body    []byte
}

func (r *Room) start() {
	for {
		select {
		case player := <-r.join:
			r.Players[player] = true
		case player := <-r.leave:
			if _, ok := r.Players[player]; ok {
				delete(r.Players, player)
				close(player.send)
			}
			if len(r.Players) == 0 {
				close(r.stop)
			}
		case msg := <-r.broadcast:
			for p := range r.Players {
				if p.ID == msg.exclude {
					continue
				}
				select {
				case p.send <- msg.body:
				default: //send channel is blocked
					close(p.send)
					delete(r.Players, p)
				}

			}
		case <-r.stop:
			return
		}
	}
}

func generateID(n int) (string, error) {
	b := make([]byte, n)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		b[i] = letters[num.Int64()]
	}
	return string(b), nil
}
func NewRoom(roomID, key string) *Room {
	return &Room{
		ID:        roomID,
		Key:       key,
		Players:   make(map[*PlayerConn]bool),
		join:      make(chan *PlayerConn),
		leave:     make(chan *PlayerConn),
		broadcast: make(chan Message, 32),
		stop:      make(chan struct{}),
	}
}
