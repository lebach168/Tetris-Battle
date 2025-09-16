package data

import (
	"fmt"
	"sync"
)

type RoomManager interface {
	GetAll() ([]RoomDTO, error)
	CreateRoom(key string) (RoomDTO, error)
	JoinRoom(roomID string, key string) (RoomDTO, error)
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

	for p := range r.Players {
		dto.Players = append(dto.Players, PlayerDTO{ID: p.ID})
	}

	return dto
}
func (i *InMemoryRoomManager) GetAll() ([]RoomDTO, error) {
	rooms := make([]RoomDTO, 0, len(i.Rooms))
	for _, r := range i.Rooms {
		rooms = append(rooms, r.ToDTO())
	}
	return rooms, nil
}
func (i *InMemoryRoomManager) exists(roomID string) bool {
	//mutex lock ở nơi gọi. ko nên lock ở đây nữa
	_, ok := i.Rooms[roomID]
	return ok
}
func (i *InMemoryRoomManager) CreateRoom(key string) (RoomDTO, error) {
	i.mu.Lock()
	defer i.mu.Unlock()

	for tries := 0; tries <= 5; tries++ {
		roomID, err := generateID(5)
		if err != nil {
			return RoomDTO{}, err
		}
		if !i.exists(roomID) {
			room := NewRoom(roomID, key)
			i.Rooms[roomID] = room
			return room.ToDTO(), nil
		}

	}
	return RoomDTO{}, fmt.Errorf("server is busy")

}

func (i *InMemoryRoomManager) JoinRoom(roomID string, key string) (RoomDTO, error) {
	i.mu.Lock()
	defer i.mu.Unlock()
	if !i.exists(roomID) {
		return RoomDTO{}, fmt.Errorf("not found")
	}
	fmt.Println(len(i.Rooms[roomID].Players))
	if len(i.Rooms[roomID].Players) >= 2 {
		return RoomDTO{}, fmt.Errorf("room is full")
	}
	return i.Rooms[roomID].ToDTO(), nil
}

func NewInMemoryRoomManager() *InMemoryRoomManager {
	return &InMemoryRoomManager{
		Rooms: make(map[string]*Room),
	}
}
