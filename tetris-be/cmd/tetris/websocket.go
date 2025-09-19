package main

import (
	"github.com/gorilla/websocket"
	"net/http"
	"tetris-be/internal/data"
)

const (
	// Maximum message size allowed from peer.
	readBufferSize  = 1024 * 5
	writeBufferSize = 1024 * 5
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
		//origin := r.Header.Get("Origin")
		//return origin == "https://frontend.com"

	},
}

func serveWs(roomManager data.RoomManager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			serverErrorResponse(w, r, err)
			return
		}

		roomID := readString(r.URL.Query(), "roomid", "")
		playerID := readString(r.URL.Query(), "playerid", "")
		room, err := roomManager.Get(roomID)
		if err != nil {
			notFoundResponse(w, r)
			return
		}
		player := data.NewPlayerConn(playerID, room, conn)
		roomManager.AddPlayer(player)
		go player.Read()
		go player.Write()
	})
}
