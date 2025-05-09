package handler

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"tetris-be/util"
)

type Room struct {
	RoomId  string `json:"roomId"`
	Player1 string `json:"p1"` // Client ID
	Player2 string `json:"p2"` //

}

var (
	rooms   = make(map[string]*Room)
	roomsMu sync.RWMutex
)

// fixed mock data
func init() {
	roomsMu.Lock()
	rooms["123456"] = &Room{
		RoomId:  "123456",
		Player1: "p1",
		Player2: "p2",
	}
	roomsMu.Unlock()
	fmt.Println("init data for test...")
}

const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ123456789"
const roomIDLength = 6

func generateRoomID() (string, error) {
	b := make([]byte, roomIDLength)
	maxAttempts := 10

	for attempts := 0; attempts < maxAttempts; attempts++ {
		for i := range b {
			randIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
			if err != nil {
				return "", fmt.Errorf("Failed to generate random index")
			}
			b[i] = chars[randIndex.Int64()]
		}

		roomID := string(b)
		if rooms[roomID] == nil {
			return roomID, nil
		}
	}

	return "", fmt.Errorf("Failed to generate unique room ID after multiple attempts")
}

func ListRooms(w http.ResponseWriter, r *http.Request) {
	fmt.Println("List room...")
	roomsMu.RLock()
	var roomList []Room
	for _, v := range rooms {
		roomList = append(roomList, *v)
	}
	roomsMu.RUnlock()
	util.RespondWithJSON(w, http.StatusOK, roomList)
}

type CreateRoomRequest struct {
	PlayerID string `json:"playerId"`
}

func CreateRoom(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Creating room...")

	// Validate content type
	if r.Header.Get("Content-Type") != "application/json" {
		util.RespondWithError(w, http.StatusBadRequest, "Invalid content type")
		return
	}

	// Parse request body
	var req CreateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	defer r.Body.Close()

	// Set default PlayerID if empty
	playerId := req.PlayerID
	if playerId == "" {
		playerId = "guest1"
	}

	// Generate Room ID
	roomsMu.Lock()
	defer roomsMu.Unlock()

	roomID, err := generateRoomID()
	if err != nil {
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to generate room ID")
		return
	}

	// Create new room
	room := &Room{
		RoomId:  roomID,
		Player1: playerId,
		Player2: "",
	}
	rooms[roomID] = room
	// Send response
	util.RespondWithJSON(w, http.StatusOK, room)
}

// TODO join room
type JoinRoomRequest struct {
	PlayerID string `json:"playerId"`
	RoomID   string `json:"roomId"`
}

func JoinRoom(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Joining room...")

	// Validate content type
	if r.Header.Get("Content-Type") != "application/json" {
		util.RespondWithError(w, http.StatusBadRequest, "Invalid content type")
		return
	}

	// Parse request body
	var req JoinRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	defer r.Body.Close()

	// Set default PlayerID if empty
	playerID := req.PlayerID
	if playerID == "" {
		playerID = "guest2"
	}

	roomsMu.Lock()
	defer roomsMu.Unlock()

	room, exists := rooms[req.RoomID]
	if !exists {
		util.RespondWithError(w, http.StatusNotFound, "Room not found")
		return
	}

	// Check if room is full
	if room.Player2 != "" {
		util.RespondWithError(w, http.StatusConflict, "Room is full")
		return
	}
	// Join room
	room.Player2 = playerID

	util.RespondWithJSON(w, http.StatusOK, room)
}
func DeleteRoom(roomID string) {
	roomsMu.Lock()
	_, ok := rooms[roomID]
	if ok {
		delete(rooms, roomID)
	}

	roomsMu.Unlock()
}
