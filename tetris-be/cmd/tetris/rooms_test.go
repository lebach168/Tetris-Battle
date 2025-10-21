package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"tetris-be/internal/game"
)

func TestGetRooms(t *testing.T) {
	stubRoomManager := newStubRoomManager()
	server := NewServerHandler(nil, nil, stubRoomManager)
	t.Run("get all rooms info happy case", func(t *testing.T) {

		req := newGetRoomsRequest()
		response := httptest.NewRecorder()
		server.ServeHTTP(response, req)

		assertStatusCode(t, http.StatusOK, response.Code)
		assertContentType(t, "application/json", response.Header().Get("Content-Type"))
		var responseBody struct {
			Rooms []game.RoomDTO `json:"rooms"`
		}
		err := json.NewDecoder(response.Body).Decode(&responseBody)
		assertNoError(t, err)
		expectedRooms, _ := stubRoomManager.GetAllDTO()
		assertRooms(t, expectedRooms, responseBody.Rooms)
	})
}
func TestJoinRoom(t *testing.T) {
	stubRoomManager := newStubRoomManager()
	cfg := LoadConfig()
	server := NewServerHandler(nil, cfg, stubRoomManager)
	t.Run("join room happy case", func(t *testing.T) {
		in := struct {
			PlayerID string
			Key      string
		}{
			PlayerID: "player-x",
		}
		roomID := "ABC12"
		req := newJoinRoomRequest(in, roomID)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, req)
		assertStatusCode(t, http.StatusAccepted, response.Code)
		var responseBody struct {
			Room  game.RoomDTO `json:"room"`
			WsURL string       `json:"ws_url"`
		}
		err := json.NewDecoder(response.Body).Decode(&responseBody)
		assertNoError(t, err)
		assertWsURL(t, responseBody.WsURL)
		expectedRoom := game.RoomDTO{
			ID: roomID,
			Players: []game.PlayerDTO{
				{ID: "anon123"},
			},
		}
		assertRoom(t, expectedRoom, responseBody.Room)

	})

	t.Run("join into a room without roomID/create room", func(t *testing.T) {
		in := struct {
			PlayerID string
			Key      string
		}{
			PlayerID: "player-x",
		}

		req := newCreateRoomRequest(in)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, req)
		assertStatusCode(t, http.StatusAccepted, response.Code)

		var responseBody struct {
			Room  game.RoomDTO `json:"room"`
			WsURL string       `json:"ws_url"`
		}
		err := json.NewDecoder(response.Body).Decode(&responseBody)
		assertNoError(t, err)
		assertWsURL(t, responseBody.WsURL)
		expectedRoomID := "12345"
		assertRoomID(t, expectedRoomID, responseBody.Room.ID)
	})
	t.Run("join into a full room", func(t *testing.T) {
		in := struct {
			PlayerID string
			Key      string
		}{
			PlayerID: "player-x",
		}
		roomID := "DEF34"
		req := newJoinRoomRequest(in, roomID)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, req)
		assertStatusCode(t, http.StatusConflict, response.Code)
	})
	t.Run("join into non-exists room", func(t *testing.T) {
		in := struct {
			PlayerID string
			Key      string
		}{
			PlayerID: "player-x",
		}
		roomID := "DEF00"
		req := newJoinRoomRequest(in, roomID)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, req)
		assertStatusCode(t, http.StatusNotFound, response.Code)
	})
	t.Run("request missing playerid", func(t *testing.T) {
		in := struct {
			PlayerID string
			Key      string
		}{}
		roomID := "ABC12"
		req := newJoinRoomRequest(in, roomID)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, req)
		assertStatusCode(t, http.StatusBadRequest, response.Code)
	})
}

func newStubRoomManager() *game.InMemoryRoomManager {
	stubRoomManager := game.NewInMemoryRoomManager()
	rooms := []game.Room{
		{
			ID:  "ABC12",
			Key: "key-1",
			Players: map[*game.PlayerConn]bool{
				{ID: "anon123"}: true,
			},
		},
		{
			ID:  "DEF34",
			Key: "key-2",
			Players: map[*game.PlayerConn]bool{
				{ID: "player-2"}: true,
				{ID: "player-3"}: true,
			},
		},
		{
			ID:      "XYZ00",
			Key:     "key-3",
			Players: make(map[*game.PlayerConn]bool), // empty room
		},
	}
	for _, room := range rooms {
		stubRoomManager.Rooms[room.ID] = &room
	}
	return stubRoomManager
}
func newCreateRoomRequest(data interface{}) *http.Request {
	body, _ := json.Marshal(data)
	req := httptest.NewRequest(http.MethodPost, "/rooms", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func newGetRoomsRequest() *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/rooms", nil)
	return req
}

func newJoinRoomRequest(data interface{}, roomID string) *http.Request {
	body, _ := json.Marshal(data)
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/rooms?roomid=%s", roomID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}
func assertRoomIDLength(t testing.TB, got string, expectedLen int) {
	t.Helper()
	if len(got) != expectedLen {
		t.Errorf("room id len: %d , want len: %d", len(got), expectedLen)
	}

}
func assertRoomID(t testing.TB, got, want string) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("room id len: %d , want len: %d", len(got), len(want))
	}

}
func assertNoError(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("did not expect error but got: %v", err)
	}
}
func assertRoomSize(t *testing.T, players []game.PlayerConn, expectedSize int) {
	t.Helper()
	if len(players) != expectedSize {
		t.Errorf("expected %d players in room, got %d", expectedSize, len(players))
	}
}
func assertPlayerInRoom(t *testing.T, players []game.PlayerConn, expectedPlayerID string) {
	t.Helper()
	found := false
	for _, player := range players {
		if player.ID == expectedPlayerID {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("expected player %s to be in room, but got players: %v", expectedPlayerID, players)
	}
}
func assertStatusCode(t testing.TB, want, got int) {
	t.Helper()
	if got != want {
		t.Errorf("got status %v, want %v ", got, want)
	}
}
func assertRooms(t testing.TB, want, got []game.RoomDTO) {
	t.Helper()

	if len(got) != len(want) {
		t.Errorf("got %v want %v", got, want)
	}

}
func assertRoom(t testing.TB, want, got game.RoomDTO) {
	t.Helper()

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
}
func getRoomsFromResponse(body string) ([]game.Room, error) {
	data := make(map[string][]game.Room)
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		return nil, err
	}

	return data["rooms"], nil
}

func assertWsURL(t testing.TB, got string) {
	t.Helper()
	u, err := url.Parse(got)
	assertNoError(t, err)
	q := u.Query()
	roomID := q.Get("roomid")
	playerID := q.Get("playerid")

	if !strings.HasPrefix(got, "ws://") || len(roomID) < 5 || len(playerID) < 5 {
		t.Errorf("got %s, but want ws://host/ws/match?roomid=.....&playerid=.....", got)
	}
}
func assertContentType(t testing.TB, want string, got string) {
	t.Helper()
	if got != want {
		t.Errorf("got status %v, want %v ", got, want)
	}
}
