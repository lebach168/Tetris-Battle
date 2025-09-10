package main

import (
	"log/slog"
	"net/http"
	"tetris-be/internal/store/room"
)

func addRoutes(
	mux *http.ServeMux,
	logger *slog.Logger,
	config *Config,
	roomsStore store.RoomsStore,
) http.Handler {

	mux.HandleFunc("GET /healthcheck", healthcheck)

	mux.Handle("GET /rooms", getAllRoomsHandler(roomsStore))
	mux.Handle("POST /rooms", createRoomHandler(roomsStore))
	mux.Handle("POST /rooms/{roomID}", joinRoomHandler(roomsStore))

	return mux
}
