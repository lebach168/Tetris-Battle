package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
			RoomID  string         `json:"id"`
			Players []model.Player `json:"p"`
		}

		err := json.Unmarshal([]byte(rawJSON), &responseBody)

		assertNoError(t, err)
		assertRoomID(t, responseBody.RoomID, expectedLen)
		assertPlayerInRoom(t, responseBody.Players, expectedPlayer.ID)
		assertRoomSize(t, responseBody.Players, 1)
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

func newCreateRoomRequest(data interface{}) *http.Request {
	body, _ := json.Marshal(data)
	req, _ := http.NewRequest(http.MethodPost, "/rooms", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func assertRoomID(t testing.TB, got string, expectedLen int) {
	t.Helper()
	if len(got) != expectedLen {
		t.Errorf("room id len: %d , want len: %d", len(got), expectedLen)
	}

}

func assertNoError(t *testing.T, err error) {
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
