package game

import (
	"crypto/rand"
	"log"
	"math/big"
)

const letters = "abcdefghijklmnopqrstxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type Room struct {
	ID        string
	Key       string
	Players   map[string]*PlayerConn // map[playerId] *PlayerConn
	join      chan *PlayerConn
	leave     chan *PlayerConn
	broadcast chan Packet
	game      *Game

	stop          chan struct{}
	callbackClose func()
}

type Packet struct {
	excludeID string
	body      []byte
}

func (r *Room) listenAndServe() {
	defer r.callbackClose()

	for {
		select {
		case pConn := <-r.join:
			r.Players[pConn.ID] = pConn
			log.Printf("[ws][room:%s] num players: %v ", r.ID, len(r.Players))
			//fmt.Println(player.ID)//for debug
		case player := <-r.leave:
			if conn, ok := r.Players[player.ID]; ok && conn != nil {
				delete(r.Players, player.ID)
				close(player.send)
			}
			if len(r.Players) == 0 {
				close(r.stop)
			}
		case msg := <-r.broadcast:
			for _, pConn := range r.Players {
				if pConn.ID == msg.excludeID {
					continue
				}
				select {
				case pConn.send <- msg.body:
				default: //send channel is blocked
					log.Printf("Drop message for %s: outbound full", pConn.ID)
					close(pConn.send)
					delete(r.Players, pConn.ID)
				}

			}
		case <-r.stop:
			return
		}
	}
}

func GenerateID(n int) (string, error) {
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
func NewRoom(roomID, key string, close func()) *Room {
	return &Room{
		ID:            roomID,
		Key:           key,
		Players:       make(map[string]*PlayerConn),
		join:          make(chan *PlayerConn),
		leave:         make(chan *PlayerConn),
		broadcast:     make(chan Packet, 32),
		stop:          make(chan struct{}),
		callbackClose: close,
		game:          NewGame(),
	}
}
