package main

import (
	"fmt"
	"net/http"
	"tetris-be/internal/data"
	"tetris-be/internal/validator"
)

type input struct {
	PlayerID string
	Key      string
}

func getAllRoomsHandler(roomManager data.RoomManager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := roomManager.GetAll()
		if err != nil {
			serverErrorResponse(w, r, err)
			return
		}
		encode(w, http.StatusOK, envelope{"rooms": data}, nil)
	})
}

func joinRoomHandler(roomManager data.RoomManager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//read param
		roomID := readString(r.URL.Query(), "roomid", "")
		//read body
		in, err := decode[input](r)
		if err != nil {
			serverErrorResponse(w, r, err)
			return
		}

		v := validator.New()
		ValidateInput(v, in)

		if !v.Valid() {
			failedValidationResponse(w, r, v.Errors)
			return
		}

		//call service - join or create
		var data data.RoomDTO
		switch roomID {
		case "":
			data, err = roomManager.CreateRoom(in.Key)
		default:
			data, err = roomManager.JoinRoom(roomID, in.Key)
		}
		//catch error
		if err != nil {
			switch {
			case err.Error() == "not found":
				notFoundResponse(w, r)
			case err.Error() == "room is full":
				conflictResponse(w, r)
			default:
				serverErrorResponse(w, r, err)
			}
		}
		//send response { wsurl:...,room:...}
		wsURL := fmt.Sprintf("ws://match?roomid=%s&playerid=%s", data.ID, in.PlayerID)
		encode(w, http.StatusAccepted, envelope{"room": data, "ws_url": wsURL}, nil)
	})
}

func ValidateInput(v *validator.Validator, in input) {
	v.Check(in.PlayerID != "", "playerID", "playerID must be provided")
	v.Check(len(in.PlayerID) <= 15, "playerID", "invalid request body")
	v.Check(len(in.Key) <= 7, "key", "wrong key")
}
