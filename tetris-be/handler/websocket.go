package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type WSMessage struct {
	Type    string          `json:"type"`    // e.g. "key", "checksum", etc.
	To      string          `json:"To"`      // receiver : all |  player1ID | player2ID
	Payload json.RawMessage `json:"payload"` //

}
type KeyPayload struct {
	Keys []string `json:"keys"` // e.g. ["left", "rotate", "drop"]
}
type ChecksumPayload struct {
	Hash string `json:"hash"` // hash của board (ví dụ md5 hoặc sha1)
}
type SyncBoardPayload struct {
	Board       [][]int     `json:"board"`       // matrix trạng thái
	ActivePiece interface{} `json:"activePiece"` // nếu muốn sync luôn piece
	Score       int         `json:"score"`
	Timestamp   int64       `json:"timestamp"` // optional
}
type GarbagePayload struct {
	//From  string `json:"from"`
	Lines int `json:"lines"` // số dòng rác gửi
}
type GameOverPayload struct {
	From string `json:"clientID"`
}
type StartPayload struct {
	StartAt int64 `json:"startAt"` // timestamp để tất cả bắt đầu đồng bộ
}

type Client struct {
	ID     string
	RoomID string
	Conn   *websocket.Conn
	Send   chan []byte // channel để gửi ngược ra client
}

var (
	roomClients   = make(map[string][]*Client) // map[roomId][]*Client
	roomClientsMu sync.RWMutex
)
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade failed:", err)
		return
	}
	fmt.Println("Create websocket connection...")
	conn.SetReadLimit(1024)
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	//close connection after use
	defer conn.Close()
	roomID := r.URL.Query().Get("room")
	client := &Client{
		Conn:   conn,
		Send:   make(chan []byte, 2),
		RoomID: roomID,
		ID:     r.URL.Query().Get("player"),
	}

	roomClientsMu.Lock()
	roomClients[roomID] = append(roomClients[roomID], client)
	roomClientsMu.Unlock()

	//catch end broadcast message
	var wg sync.WaitGroup
    wg.Add(2) // Chờ 2 goroutine
    go func() {
        defer wg.Done()
        client.messageReceiver()
    }()
    go func() {
        defer wg.Done()
        client.messageSender()
    }()
    wg.Wait()
	// Flow:  ClientJS -message-> Server(WebsocketHandler) --> messageReceiver
	// 																	|
	// 																	↓
	// 			 messageSender <-- channel Send <-- Broadcast<--messageHandler

	//cleanup function
	defer func(roomID string) {
		conn.Close()
		roomClientsMu.Lock()
		if clients, ok := roomClients[client.RoomID]; ok {
			for i, c := range clients {
				if c == client {
					roomClients[client.RoomID] = append(clients[:i], clients[i+1:]...)
					break
				}
			}
			if len(roomClients[client.RoomID]) == 0 {
				delete(roomClients, client.RoomID)
				//DeleteRoom(roomID)
			}
		}
		
		roomClientsMu.Unlock()
		close(client.Send)
	}(roomID)

}

func handleWSMessage(sender *Client, message []byte) {
	var msg WSMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("Invalid message: %v", err)
		return
	}

	switch msg.Type {
	case "key":
		var payload KeyPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			log.Println("Invalid key payload")
			return
		}
		broadcastToOthers(sender, message)
	//TODO
	// case "checksum":
	//     var payload ChecksumPayload
	//     // Xóa roomId khỏi struct ChecksumPayload
	//     broadcastToOthers(sender, message)

	// case "sync_board":
	// 	// full state sync
	// 	broadcastToOthers(sender, sender.RoomID, message)

	case "sent_garbage":
		broadcastToOthers(sender, message)

	case "game_over":
		broadcastToAll(sender.RoomID, message)

	case "start":
		broadcastToAll(sender.RoomID, message)

	default:
		log.Println("Unknown message type:", msg.Type)
	}
}

// Upgrade http to WS connection

func (c *Client) messageReceiver() {
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			break // Ngắt kết nối nếu có lỗi
		}

		handleWSMessage(c, message)
	}
}
func (c *Client) messageSender() {
	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.Conn.WriteMessage(websocket.TextMessage, message)
		}
	}

}

//TODO

func broadcastToOthers(sender *Client, msg []byte) {
	for _, client := range roomClients[sender.RoomID] {
		if client != sender {
			select {
			case client.Send <- msg:
			default:
				close(client.Send)
				delete(roomClients, client.ID)
			}
		}
	}
}

func broadcastToAll(roomId string, msg []byte) {
	for _, client := range roomClients[roomId] {
		select {
		case client.Send <- msg:
		default:
			close(client.Send)
			delete(roomClients, client.ID)
		}
	}
}
