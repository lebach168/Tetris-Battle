package main

import (
	"log/slog"
	"net/http"
	"tetris-be/internal/game"
)

func addRoutes(
	mux *http.ServeMux,
	logger *slog.Logger,
	config *Config,
	roomManager game.RoomManager,
) http.Handler {

	mux.HandleFunc("GET /healthcheck", healthcheck)

	//rooms?roomid=... to join an existing room. If the parameter is missing, a new room will be created
	mux.Handle("GET /rooms", getAllRoomsHandler(roomManager))
	mux.Handle("POST /rooms", joinRoomHandler(config, roomManager))

	mux.Handle("GET /ws/match", serveWs(roomManager))

	return mux
}
