package game

//import (
//	"fmt"
//	"github.com/gorilla/websocket"
//	"log"
//	"time"
//)
//
//type PlayerConnLog struct {
//	id       string
//	conn     *websocket.Conn
//	outbound chan []byte // serialize writes qua 1 goroutine
//}
//
//func NewPlayerConnLogWithQueue(id string, ws *websocket.Conn) *PlayerConnLog {
//	return &PlayerConnLog{
//		id:       id,
//		conn:     ws,
//		outbound: make(chan []byte, outboundSize),
//	}
//}
//
//func (p *PlayerConnLog) Close() {
//	close(p.outbound) // sẽ làm writePump thoát
//	_ = p.conn.Close()
//}
//
//// Enqueue gửi message an toàn (không ghi trực tiếp vào conn từ nhiều goroutine)
//func (p *PlayerConnLog) Enqueue(msg []byte) {
//	select {
//	case p.outbound <- msg:
//	default:
//		// backpressure: queue đầy -> log và drop (hoặc kick)
//		log.Printf("[ws][%s] outbound queue full -> dropping message", p.id)
//		return
//	}
//}
//
//func (p *PlayerConnLog) WritePump() {
//	ticker := time.NewTicker(pingPeriod)
//	defer func() {
//		ticker.Stop()
//		_ = p.conn.Close()
//	}()
//
//	for {
//		select {
//		case msg, ok := <-p.outbound:
//			_ = p.conn.SetWriteDeadline(time.Now().Add(writeWait))
//			if !ok {
//				// channel đóng: gửi close tickFrame
//				log.Printf("[ws][%s] outbound closed -> send close tickFrame", p.id)
//				_ = p.conn.WriteMessage(websocket.CloseMessage, []byte{})
//				return
//			}
//			w, err := p.conn.NextWriter(websocket.TextMessage)
//			if err != nil {
//				log.Printf("[ws][%s] %s", p.id, classifyErr("NextWriter error", err))
//				return
//			}
//			if _, err = w.Write(msg); err != nil {
//				log.Printf("[ws][%s] %s", p.id, classifyErr("Write error", err))
//				_ = w.Close()
//				return
//			}
//			if err := w.Close(); err != nil {
//				log.Printf("[ws][%s] %s", p.id, classifyErr("Close writer error", err))
//
//				return
//			}
//		case <-ticker.C:
//			_ = p.conn.SetWriteDeadline(time.Now().Add(writeWait))
//			if err := p.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
//				log.Printf("[ws][%s] %s", p.id, classifyErr("Ping write error", err))
//				return
//			}
//		}
//	}
//}
//
//func (p *PlayerConnLog) ReadPump(onMessage func([]byte)) {
//	defer func() {
//		_ = p.conn.Close()
//	}()
//
//	p.conn.SetReadLimit(maxMessageSize)
//	_ = p.conn.SetReadDeadline(time.Now().Add(pongWait))
//	p.conn.SetPongHandler(func(string) error {
//		// nhận pong -> gia hạn deadline
//		fmt.Printf("receive pong at: %v \n", time.Now())
//		return p.conn.SetReadDeadline(time.Now().Add(pongWait))
//	})
//
//	for {
//		mt, message, err := p.conn.ReadMessage()
//		if err != nil {
//			log.Printf("[ws][%s] %s", p.id, classifyErr("ReadMessage error", err))
//			return
//		}
//		if mt != websocket.TextMessage {
//			// bỏ qua binary/others
//			continue
//		}
//		onMessage(message)
//	}
//}
