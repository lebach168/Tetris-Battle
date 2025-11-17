package game

import (
	"crypto/rand"
	"log"
	"math/big"
)

const letters = "abcdefghijklmnopqrstxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type Room struct {
	ID          string
	Key         string
	PlayerConns map[string]*PlayerConn // map[playerId] *PlayerConn
	join        chan *PlayerConn
	leave       chan *PlayerConn
	broadcast   chan Packet
	game        *Game

	stop          chan struct{}
	callbackClose func()
}

type Packet struct {
	directId  string //direct message to playerid
	excludeId string // broadcast all exclude playerid // else this field and directid empty then broadcast all
	body      []byte
}

func (r *Room) listenAndServe() {
	defer r.callbackClose()

	for {
		select {
		case pConn := <-r.join:
			r.PlayerConns[pConn.ID] = pConn
			log.Printf("[ws][room:%s] %s joined, num players: %v ", r.ID, pConn.ID, len(r.PlayerConns))

			//fmt.Println(player.ID)//for debug
		case playerConn := <-r.leave:
			if conn, ok := r.PlayerConns[playerConn.ID]; ok && playerConn == conn && conn != nil {
				delete(r.PlayerConns, playerConn.ID)
				if len(r.PlayerConns) == 0 {
					close(r.stop)
				}
			}
			if playerConn != nil {
				close(playerConn.send)
			}

		case msg := <-r.broadcast:
			for _, pConn := range r.PlayerConns {
				if msg.directId != "" {
					if pConn.ID != msg.directId {
						continue
					}
				} else if pConn.ID == msg.excludeId {
					continue
				}
				select {
				case pConn.send <- msg.body:
				default: //send channel is blocked
					log.Printf("Drop message for %s: outbound full", pConn.ID)
					close(pConn.send)
					delete(r.PlayerConns, pConn.ID)
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
		PlayerConns:   make(map[string]*PlayerConn),
		join:          make(chan *PlayerConn),
		leave:         make(chan *PlayerConn),
		broadcast:     make(chan Packet, 32),
		stop:          make(chan struct{}),
		callbackClose: close,
		game:          NewGame(),
	}
}
