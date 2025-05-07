package api

import (
	"tetris-be/handler"

	"github.com/gorilla/mux"
)

func RegisterRoutes(r *mux.Router) {

	r.HandleFunc("/rooms", handler.ListRooms).Methods("Get")
	r.HandleFunc("/rooms/create", handler.CreateRoom).Methods("POST")
	r.HandleFunc("/rooms/join", handler.JoinRoom).Methods("PUT")
	// r.HandleFunc("/start-game", StartGameHandler).Methods("POST")
	// r.HandleFunc("/rooms/{id}/clients", GetClientsHandler).Methods("GET")
	r.HandleFunc("/ws", handler.WebSocketHandler) 
}
