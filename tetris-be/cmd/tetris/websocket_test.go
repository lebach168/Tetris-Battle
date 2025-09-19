package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"tetris-be/internal/data"
)

func TestWebsocketConnection(t *testing.T) {
	stubRoomManager := newStubRoomManager()
	server := NewServerHandler(nil, nil, stubRoomManager)
	t.Run("successful connection", func(t *testing.T) {
		roomID := "ABC12"
		playerID := "anon123"
		req := newWsRequest(roomID, playerID)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, req)

		totalPlayers := len(stubRoomManager.Rooms[roomID].Players)
		if totalPlayers == 0 {
			t.Errorf("got %d want 1", totalPlayers)
		}

	})

}

func TestSendAndReceiveMessage(t *testing.T) {

	t.Run("test read message", func(t *testing.T) {

		clientConn, cleanup := mockConnHandler()
		defer cleanup()
		msg := []byte(`{
						  "action": "start",
						  "payload": {
							"tetrominos": [3, 7, 2, 1, 6, 4, 5, 2, 7, 1, 3, 6, 5, 4, 2, 1, 7, 3, 6, 5],
							"board": [
							  [0, 0, 0, 0, 0],
							  [0, 0, 0, 0, 0],
							  [0, 0, 0, 0, 0],
							  [0, 0, 0, 0, 0]
							]
						  }
						}`)

		err := clientConn.WriteMessage(websocket.TextMessage, msg)
		assertNoError(t, err)

		_, msgReceived, err := clientConn.ReadMessage()
		assertNoError(t, err)
		fmt.Println(string(msgReceived))
		var msgType data.Message
		assert.JSONEq(t, string(msg), string(msgReceived))

		err = json.Unmarshal(msgReceived, &msgType)
		assertNoError(t, err)

	})
}
func mockConnHandler() (*websocket.Conn, func()) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverConn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Errorf("upgrade fail %d", err)
			return
		}
		go func() {
			defer serverConn.Close()
			for {
				msgType, msg, err := serverConn.ReadMessage()
				if err != nil {
					return // đóng kết nối nếu serverConn của client ngắt
				}
				// Echo lại message về client
				err = serverConn.WriteMessage(msgType, msg)
				if err != nil {
					return
				}
			}
		}()

	}))

	url := "ws" + strings.TrimPrefix(server.URL, "http")
	clientConn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		panic(err)
	}
	cleanup := func() {
		clientConn.Close()
		server.Close()
	}
	return clientConn, cleanup
}

func TestConcurrentJoinSameRoom(t *testing.T) {
	t.Run("Test race condition when 3rd player join room", func(t *testing.T) {

	})
}
func TestHandleHighTraffic(t *testing.T) {

}
func newWsRequest(roomID, playerID string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/ws/match?roomid=%s&playerid=%s", roomID, playerID), nil)
	req.URL.Scheme = "ws"
	return req
}

//type StubRoomManager struct {
//	rooms map[string]*data.Room
//}
//func (s *StubRoomManager) JoinRoom(roomID string, key string) (data.RoomDTO, error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func NewStubRoomManager() *StubRoomManager {
//	return &StubRoomManager{
//		rooms: make(map[string]*data.Room),
//	}
//}
//
//func (s *StubRoomManager) GetAllDTO() ([]data.RoomDTO, error) {
//
//	panic("implement me")
//}
//func (s *StubRoomManager) Get(roomID string) (*data.Room, error) {
//	if room, exists := s.rooms[roomID]; exists {
//		return room, nil
//	}
//	return nil, errors.New("room not found")
//}
//
//func (s *StubRoomManager) AddPlayer(player *data.PlayerConn) {
//	room := s.rooms[player.GetRoomID()]
//	player.SetConn(conn)
//
//	room.join <- player
//}
//
//func (s *StubRoomManager) CreateRoom(key string) *data.Room {
//	roomID, _ := data.GenerateID(5)
//	room := &data.Room{
//		ID: roomID,
//	}
//	s.rooms[roomID] = room
//	return room
//}
