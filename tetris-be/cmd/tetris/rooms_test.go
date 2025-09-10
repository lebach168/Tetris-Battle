package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
	"tetris-be/internal/model"
	store "tetris-be/internal/store/room"
)

func TestCreateRoom(t *testing.T) {
	stubStore := store.NewInmemRoomStore()
	server := NewServerHandler(nil, nil, stubStore)

	t.Run("happy case ", func(t *testing.T) {
		body := struct {
			PlayerID string `json:"playerID"`
		}{PlayerID: "anon123"}

		req := newCreateRoomRequest(body)

		expectedPlayer := model.Player{ID: "anon123"}

		response := httptest.NewRecorder()
		server.ServeHTTP(response, req)
		assertStatusCode(t, http.StatusCreated, response.Code)

		rawJSON := response.Body.String()
		expectedLen := 5
		var responseBody struct {
			WsURL string `json:"ws_url"`
			Room  struct {
				ID      string         `json:"id"`
				Players []model.Player `json:"p"`
			} `json:"room"`
		}

		err := json.Unmarshal([]byte(rawJSON), &responseBody)

		assertNoError(t, err)

		assertRoomIDLength(t, responseBody.Room.ID, expectedLen)

		assertWsURL(t, responseBody.WsURL)
		assertPlayerInRoom(t, responseBody.Room.Players, expectedPlayer.ID)
		assertRoomSize(t, responseBody.Room.Players, 1)
	})
	t.Run("empty player id", func(t *testing.T) {
		body := struct {
			PlayerID string `json:"playerID"`
		}{PlayerID: ""}
		req := newCreateRoomRequest(body)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, req)
		assertStatusCode(t, http.StatusBadRequest, response.Code)
	})
	t.Run("missing playerID field", func(t *testing.T) {
		body := struct {
		}{}
		req := newCreateRoomRequest(body)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, req)
		assertStatusCode(t, http.StatusBadRequest, response.Code)
	})
}
func TestGetRooms(t *testing.T) {
	stubStore := store.NewInmemRoomStore()
	rooms := []model.Room{
		{
			ID: "ABC12",
			P: []model.Player{
				{ID: "p1"},
				{ID: "p2"},
			},
		},
		{
			ID: "DEF34",
			P: []model.Player{
				{ID: "player3"},
			},
		},
	}
	for _, room := range rooms {
		stubStore.AddRoom(room)
	}
	server := NewServerHandler(nil, nil, stubStore)
	t.Run("get all rooms info happy case", func(t *testing.T) {

		req := newGetRoomsRequest()
		response := httptest.NewRecorder()

		server.ServeHTTP(response, req)

		assertStatusCode(t, http.StatusOK, response.Code)

		roomsFromResponse, err := getRoomsFromResponse(response.Body.String())
		assertNoError(t, err)
		assertRooms(t, rooms, roomsFromResponse)
	})
}
func TestJoinRoom(t *testing.T) {

	rooms := []model.Room{
		{
			ID: "ABC12",
			P: []model.Player{
				{ID: "p1"},
				{ID: "p2"},
			},
			Key: "KEY01",
		},
		{
			ID: "DEF34",
			P: []model.Player{
				{ID: "anon123"},
			},
			Key: "KEY02",
		},
	}

	t.Run("join room happy case", func(t *testing.T) {
		stubStore := store.NewInmemRoomStore()
		server := NewServerHandler(nil, nil, stubStore)
		for _, room := range rooms {
			stubStore.AddRoom(room)
		}
		body := struct {
			PlayerID string `json:"playerID"`
		}{PlayerID: "anon456"}
		roomID := "DEF34"
		req := newJoinRoomRequest(body, roomID)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, req)

		assertStatusCode(t, http.StatusAccepted, response.Code)
		var responseBody struct {
			WsURL string `json:"ws_url"`
			Room  struct {
				ID      string         `json:"id"`
				Players []model.Player `json:"p"`
			} `json:"room"`
		}
		err := json.NewDecoder(response.Body).Decode(&responseBody)

		assertNoError(t, err)
		assertWsURL(t, responseBody.WsURL)
		assertRoomID(t, responseBody.Room.ID, roomID)
		assertRoomSize(t, responseBody.Room.Players, 2)

	})
	t.Run("join into a full room", func(t *testing.T) {
		stubStore := store.NewInmemRoomStore()
		server := NewServerHandler(nil, nil, stubStore)
		for _, room := range rooms {
			stubStore.AddRoom(room)
		}
		body := struct {
			PlayerID string `json:"playerID"`
		}{PlayerID: "anon456"}
		roomID := "ABC12"
		req := newJoinRoomRequest(body, roomID)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, req)
		assertStatusCode(t, http.StatusConflict, response.Code)
	})
	t.Run("join into non-exists room", func(t *testing.T) {
		stubStore := store.NewInmemRoomStore()
		server := NewServerHandler(nil, nil, stubStore)
		for _, room := range rooms {
			stubStore.AddRoom(room)
		}
		body := struct {
			PlayerID string `json:"playerID"`
		}{PlayerID: "anon456"}
		roomID := "DEF12"
		req := newJoinRoomRequest(body, roomID)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, req)
		assertStatusCode(t, http.StatusConflict, response.Code)
	})

}

func newCreateRoomRequest(data interface{}) *http.Request {
	body, _ := json.Marshal(data)
	req, _ := http.NewRequest(http.MethodPost, "/rooms", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func newGetRoomsRequest() *http.Request {
	req, _ := http.NewRequest(http.MethodGet, "/rooms", nil)
	return req
}

func newJoinRoomRequest(data interface{}, roomID string) *http.Request {
	body, _ := json.Marshal(data)
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/rooms/%s", roomID), bytes.NewReader(body))
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
	if got != want {
		t.Errorf("room id len: %s , want len: %s", got, want)
	}

}
func assertNoError(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("did not expect error but got: %v", err)
	}
}
func assertRoomSize(t *testing.T, players []model.Player, expectedSize int) {
	t.Helper()
	if len(players) != expectedSize {
		t.Errorf("expected %d players in room, got %d", expectedSize, len(players))
	}
}
func assertPlayerInRoom(t *testing.T, players []model.Player, expectedPlayerID string) {
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
func assertRooms(t testing.TB, want, got []model.Room) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
}
func getRoomsFromResponse(body string) ([]model.Room, error) {
	data := make(map[string][]model.Room)
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
	key := q.Get("key")
	if len(roomID) < 5 || len(key) < 5 {
		t.Errorf("got %s, but want ws://match/?roomid=.....&key=.....", got)
	}
}
