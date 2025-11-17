package game

import (
	"fmt"
	"sync"
)

type RoomManager interface {
	Get(roomID string) (*Room, error)
	GetAllDTO() ([]RoomDTO, error)
	CreateRoom(key string) (RoomDTO, error)
	CreateMockRoom(id string) error
	JoinRoom(roomID string, key string) (RoomDTO, error)
	AddPlayer(pConn *PlayerConn)
}
type InMemoryRoomManager struct {
	Rooms map[string]*Room
	mu    sync.RWMutex
}

type RoomDTO struct {
	ID      string      `json:"ID"`
	Players []PlayerDTO `json:"players,omitempty"`
	key     string
}
type PlayerDTO struct {
	ID string `json:"ID"`
}

func (r Room) ToDTO() RoomDTO {
	dto := RoomDTO{
		ID: r.ID,
	}

	for _, pConn := range r.PlayerConns {
		dto.Players = append(dto.Players, PlayerDTO{ID: pConn.ID})
	}

	return dto
}
func (i *InMemoryRoomManager) GetAllDTO() ([]RoomDTO, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	rooms := make([]RoomDTO, 0, len(i.Rooms))
	for _, r := range i.Rooms {
		rooms = append(rooms, r.ToDTO())
	}
	return rooms, nil
}
func (i *InMemoryRoomManager) Get(roomID string) (*Room, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	if !i.exists(roomID) {
		return nil, fmt.Errorf("not found")
	}
	return i.Rooms[roomID], nil
}
func (i *InMemoryRoomManager) exists(roomID string) bool {
	//mutex lock ở nơi gọi. ko nên lock ở đây nữa
	_, ok := i.Rooms[roomID]
	return ok
}
func (i *InMemoryRoomManager) CreateRoom(key string) (RoomDTO, error) {
	i.mu.Lock()

	for tries := 0; tries <= 5; tries++ {
		roomID, err := GenerateID(5)
		if err != nil {
			return RoomDTO{}, err
		}
		if !i.exists(roomID) {
			closeRoom := func() {
				i.mu.Lock()
				defer i.mu.Unlock()
				delete(i.Rooms, roomID)
			}
			room := NewRoom(roomID, key, closeRoom)
			i.Rooms[roomID] = room

			i.mu.Unlock()
			//unlock before listenAndServe listener goroutine
			go room.listenAndServe()
			return room.ToDTO(), nil

		}

	}
	return RoomDTO{}, fmt.Errorf("server is busy")

}

func (i *InMemoryRoomManager) CreateMockRoom(id string) error {
	if _, ok := i.Rooms[id]; ok {
		return nil
	}
	i.mu.Lock()
	closeRoom := func() {
		i.mu.Lock()
		defer i.mu.Unlock()
		delete(i.Rooms, id)
	}
	room := NewRoom(id, "", closeRoom)
	i.Rooms[id] = room

	i.mu.Unlock()
	//unlock before listenAndServe listener goroutine
	go room.listenAndServe()
	return nil
	//return RoomDTO{}, fmt.Errorf("server is busy")

}
func (i *InMemoryRoomManager) AddPlayer(pConn *PlayerConn) {
	//join thông qua send vào goroutine room.listenAndServe()
	room := pConn.r
	room.join <- pConn

}

func (i *InMemoryRoomManager) JoinRoom(roomID string, key string) (RoomDTO, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	if !i.exists(roomID) {
		return RoomDTO{}, fmt.Errorf("not found")
	}
	room := i.Rooms[roomID]
	if len(room.PlayerConns) >= 2 {
		return RoomDTO{}, fmt.Errorf("room is full")
	}
	//check key
	//if key!=room.Key{
	//	...
	//}
	return i.Rooms[roomID].ToDTO(), nil
}

func NewInMemoryRoomManager() *InMemoryRoomManager {
	return &InMemoryRoomManager{
		Rooms: make(map[string]*Room),
	}
}
