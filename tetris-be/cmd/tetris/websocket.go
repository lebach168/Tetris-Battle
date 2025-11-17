package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"tetris-be/internal/game"
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
	Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {
		log.Printf("WebSocket upgrade error: %v, Status: %d", reason, status)
		http.Error(w, reason.Error(), status)
	},
}

func serveWs(roomManager game.RoomManager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		roomID := readString(r.URL.Query(), "roomid", "")
		playerID := readString(r.URL.Query(), "playerid", "")
		//TODO validate params
		err := roomManager.CreateMockRoom(roomID)
		if err != nil {
			return
		} //test only should delete this room id = ABC12
		room, err := roomManager.Get(roomID)
		if err != nil {
			notFoundResponse(w, r)
			return
		}
		//upgrade sau error để trả http response, sau khi upgrade thành công ws conn ko sửa header dc nữa.
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			serverErrorResponse(w, r, err)
			return
		}
		playerConn := game.NewPlayerConn(playerID, room, conn)
		roomManager.AddPlayer(playerConn)

		go playerConn.Read()
		go playerConn.Write()
	})
}
