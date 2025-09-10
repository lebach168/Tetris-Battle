package main

import (
	"fmt"
	"net/http"
	"tetris-be/internal/model"
	"tetris-be/internal/store/room"
	"tetris-be/internal/validator"
)

type input struct {
	PlayerID string
}

func getAllRoomsHandler(roomStore store.RoomsStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var rooms []*model.Room
		rooms, err := roomStore.GetAll()
		if err != nil {
			serverErrorResponse(w, r, err)
			return
		}
		result := envelope{"rooms": rooms}
		encode(w, http.StatusOK, result, nil)
	})
}

func createRoomHandler(roomsStore store.RoomsStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		in, err := decode[input](r)
		if err != nil {
			errorResponse(w, r, http.StatusBadRequest, "invalid JSON")
			return
		}

		v := validator.New()
		v.Check(in.PlayerID != "", "playerID", "playerID must be provided")
		v.Check(len(in.PlayerID) <= 15, "playerID", "invalid request body")

		if !v.Valid() {
			badRequestResponse(w, r, fmt.Errorf("invalid request"))
			return
		}
		var roomID string
		var room *model.Room
		maxRetries := 10
		for i := 0; i < maxRetries; i++ {
			roomID, err = generateID(5)
			room, err = roomsStore.CreateRoom(roomID, in.PlayerID)
			if err == nil {
				break
			}
		}

		if err != nil {
			serverErrorResponse(w, r, fmt.Errorf("failed to create room after %d attempts: %v", maxRetries, err))
			return
		}
		wsURL := fmt.Sprintf("ws://match/?roomid=%s&key=%s", room.ID, room.Key)
		encode(w, http.StatusCreated, envelope{"room": room, "ws_url": wsURL}, nil)
	})
}
func joinRoomHandler(roomsStore store.RoomsStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		in, err := decode[input](r)
		if err != nil {
			errorResponse(w, r, http.StatusBadRequest, "invalid JSON")
			return
		}
		roomID := r.PathValue("roomID")
		room, err := roomsStore.JoinRoom(roomID, in.PlayerID)
		if err != nil {
			if err.Error() == "room does not exists" || err.Error() == "room is full" || err.Error() == "player already joined" {
				errorResponse(w, r, http.StatusConflict, err.Error())
			} else {
				serverErrorResponse(w, r, err)
			}
			return
		}

		wsURL := fmt.Sprintf("ws://match/?roomid=%s&key=%s", room.ID, room.Key)
		encode(w, http.StatusAccepted, envelope{"room": room, "ws_url": wsURL}, nil)
	})
}
