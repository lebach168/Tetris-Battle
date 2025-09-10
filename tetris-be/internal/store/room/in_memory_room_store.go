package store

import (
	"fmt"
	"sync"
	"tetris-be/internal/model"
)

type InMemoryRoomStore struct {
	Rooms map[string]*model.Room
	mu    *sync.Mutex
}

func (s InMemoryRoomStore) IsFull(roomID string) bool {
	return len(s.Rooms[roomID].P) >= 2
}
func (s InMemoryRoomStore) Exists(roomID string) bool {
	_, ok := s.Rooms[roomID]
	return ok
}
func (s InMemoryRoomStore) GetAll() ([]*model.Room, error) {

	res := make([]*model.Room, 0)
	for _, room := range s.Rooms {
		res = append(res, room)
	}
	return res, nil
}

func (s InMemoryRoomStore) CreateRoom(roomID string, playerID string) (*model.Room, error) {
	if s.Exists(roomID) {
		return nil, fmt.Errorf("existed")
	}
	s.Rooms[roomID] = &model.Room{
		ID: roomID,
		P:  make([]model.Player, 0),
	}
	s.Rooms[roomID].AddPlayer(playerID)
	return s.Rooms[roomID], nil
}
func NewInmemRoomStore() *InMemoryRoomStore {
	return &InMemoryRoomStore{
		Rooms: make(map[string]*model.Room),
		mu:    &sync.Mutex{},
	}
}
