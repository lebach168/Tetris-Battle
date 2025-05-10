package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"tetris-be/util"
	"time"

	"github.com/gorilla/websocket"
)

type WSMessage struct {
	Type string `json:"type"` // e.g. "key", "checksum", etc.
	//To      string          `json:"to"`      // receiver : all |  player1ID | player2ID
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
	Timestamp int64 `json:"timestamp"`
	StartAt   int64 `json:"startAt"` // thời gian để tất cả bắt đầu đồng bộ
}
type StartReceiveMsgPayload struct {
	Timestamp int64 `json:"timestamp"`
}
type InitPayload struct {
	ListBlock []int `json:"listBlock"`
}
type Client struct {
	ID     string
	RoomID string
	Conn   *websocket.Conn
	Send   chan []byte // channel để gửi ra client
	once   sync.Once
	quit   chan struct{}
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
	conn.SetReadLimit(1024 * 10)
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	roomID := r.URL.Query().Get("room")
	playerId := r.URL.Query().Get("player")
	client := &Client{
		Conn:   conn,
		Send:   make(chan []byte, 256),
		RoomID: roomID,
		ID:     playerId,
		once:   sync.Once{},
		quit:   make(chan struct{}),
	}

	registerClient(client)
	if len(roomClients[roomID]) == 2 {
		msg := []byte(`{"type":"ready","payload":{}}`)
		broadcastToAll(roomID, msg)
	}
	fmt.Printf("Player %s joined room %s (clients: %d)\n", playerId, roomID, len(roomClients[roomID]))
	defer client.cleanup()

	//catch end broadcast message

	// Flow:  ClientJS -message-> Server(WebsocketHandler) --> messageReceiver
	// 																	|
	// 																	↓
	// 			 messageSender <-- channel Send <-- Broadcast<--messageHandler

	//cleanup function
	go client.messageReceiver()
	go client.messageSender()
	<-client.quit //channel bắt tín hiệu dừng connection
}
func registerClient(c *Client) {
	roomClientsMu.Lock()
	defer roomClientsMu.Unlock()

	clients := roomClients[c.RoomID]
	for i, old := range clients {
		if old.ID == c.ID {
			// replace old client
			old.cleanup()
			clients[i] = c
			roomClients[c.RoomID] = clients
			log.Printf("Replaced client %s in room %s", c.ID, c.RoomID)
			return
		}
	}

	// append if not found
	roomClients[c.RoomID] = append(clients, c)
	log.Printf("Added client %s to room %s (now %d)", c.ID, c.RoomID, len(roomClients[c.RoomID]))
}

func handleWSMessage(sender *Client, rawMessage []byte) {
	var wsMsg WSMessage
	util.JSONDecode(rawMessage, &wsMsg)

	switch wsMsg.Type {
	case "init":
		//FE send init mỗi mỗi khi nhận được msg ready.
		broadcastToOthers(sender, rawMessage)
		//FE bắt init ở bên useTetrisEventSource

	case "key_down":

		broadcastToOthers(sender, rawMessage)

	case "key_up":

		broadcastToOthers(sender, rawMessage)

	//TODO
	// case "checksum":
	//     var payload ChecksumPayload
	//     // Xóa roomId khỏi struct ChecksumPayload
	//     broadcastToOthers(sender, rawMessage)

	// case "sync_board":
	// 	// full state sync
	// 	broadcastToOthers(sender, sender.RoomID, rawMessage)

	case "sent_garbage":
		broadcastToOthers(sender, rawMessage)

	case "game_over":
		broadcastToAll(sender.RoomID, rawMessage)

	case "start":
		var ts StartReceiveMsgPayload
		util.JSONDecode(wsMsg.Payload, &ts)
		payload := StartPayload{
			Timestamp: ts.Timestamp,
			StartAt:   time.Now().Add(2 * time.Second).UnixMilli(),
		}
		msg := util.JSONEncode(map[string]interface{}{
			"type":    "start",
			"payload": payload,
		})

		broadcastToAll(sender.RoomID, msg)
	case "out_room":
		log.Println("out room here")
		close(sender.quit)
	default:
		log.Println("Unknown rawMessage type:", wsMsg.Type)
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
			if c.Conn == nil {
				return
			}
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
				log.Printf("Client %s channel full, skipping message", client.ID)
			}
		}
	}
}

func broadcastToAll(roomId string, msg []byte) {
	for _, client := range roomClients[roomId] {
		select {
		case client.Send <- msg:
		default:
			log.Printf("Client %s channel full, skipping message", client.ID)
		}
	}
}
func (c *Client) cleanup() {
	c.once.Do(func() {
		log.Printf("[DISCONNECT] Client %s left room %s\n", c.ID, c.RoomID)
		if c.Conn != nil {
			c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
			c.Conn.Close()
			c.Conn = nil
		}
		if c.Send != nil {
			close(c.Send)
		}
		if c.quit != nil {
			close(c.quit)
		}
		roomClientsMu.Lock()
		defer roomClientsMu.Unlock()
		if clients, ok := roomClients[c.RoomID]; ok {
			for i, client := range clients {
				if client == c {
					roomClients[c.RoomID] = append(clients[:i], clients[i+1:]...)
					break
				}
			}
			if len(roomClients[c.RoomID]) < 2 {
				msg := []byte(`{"type":"unready","payload":{}}`)
				log.Printf("[ROOM STATUS] Room %s now has %d players\n", c.RoomID, len(roomClients[c.RoomID]))
				broadcastToAll(c.RoomID, msg)
			}
			if len(roomClients[c.RoomID]) == 0 {
				delete(roomClients, c.RoomID)
				log.Printf("[ROOM CLOSED] Room %s deleted (no players left)\n", c.RoomID)
				//DeleteRoom(roomID)
			}
		}
	})
}
