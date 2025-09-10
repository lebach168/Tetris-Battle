package store

import (
	"fmt"
	"sync"
	"tetris-be/internal/model"
)

type InMemoryRoomStore struct {
	Rooms map[string]*model.Room
	mu    *sync.RWMutex
}

func (s *InMemoryRoomStore) JoinRoom(roomID string, playerID string) (*model.Room, error) {
	if !s.Exists(roomID) {
		return nil, fmt.Errorf("room does not exists")
	}
	if s.IsFull(roomID) {
		return nil, fmt.Errorf("room is full")
	}
	for _, p := range s.Rooms[roomID].P {
		if p.ID == playerID {
			return nil, fmt.Errorf("player already joined")
		}
	}
	s.AddPlayer(roomID, playerID)

	return s.Rooms[roomID], nil
}

func (s *InMemoryRoomStore) IsFull(roomID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.Rooms[roomID].P) >= 2
}
func (s *InMemoryRoomStore) Exists(roomID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.Rooms[roomID]
	return ok
}
func (s *InMemoryRoomStore) GetAll() ([]*model.Room, error) {

	res := make([]*model.Room, 0)
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, room := range s.Rooms {
		res = append(res, room)
	}

	return res, nil
}

func (s *InMemoryRoomStore) CreateRoom(roomID string, playerID string) (*model.Room, error) {
	if s.Exists(roomID) {
		return nil, fmt.Errorf("existed")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	key, _ := model.GenerateRoomKey(5)
	s.Rooms[roomID] = &model.Room{
		ID:  roomID,
		P:   make([]model.Player, 0),
		Key: key,
	}
	s.AddPlayer(roomID, playerID)
	return s.Rooms[roomID], nil
}
func NewInmemRoomStore() *InMemoryRoomStore {
	return &InMemoryRoomStore{
		Rooms: make(map[string]*model.Room),
		mu:    &sync.RWMutex{},
	}
}
func (s *InMemoryRoomStore) AddRoom(room model.Room) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Rooms[room.ID] = &room
}
func (s *InMemoryRoomStore) AddPlayer(roomID, playerID string) {
	s.Rooms[roomID].P = append(s.Rooms[roomID].P, model.Player{
		ID: playerID,
	})
}
