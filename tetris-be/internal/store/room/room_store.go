package store

import "tetris-be/internal/model"

type RoomsStore interface {
	GetAll() ([]*model.Room, error)
	CreateRoom(roomID string, playerID string) (*model.Room, error)
}
