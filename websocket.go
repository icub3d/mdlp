package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

func WebSocketHandler(ch <-chan struct{}) http.Handler {
	return http.HandlerFunc(func(wr http.ResponseWriter, r *http.Request) {
		ws, err := NewWebSocket(wr, r)
		if err != nil {
			log.Println(err)
			return
		}
		defer ws.Close()
		ws.Start(ch)
	})
}

type WebSocket struct {
	ws     *websocket.Conn
	closed chan struct{}
}

func NewWebSocket(w http.ResponseWriter, r *http.Request) (*WebSocket, error) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(*http.Request) bool { return true },
	}

	// Setup websocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	return &WebSocket{conn, make(chan struct{})}, nil
}

func (w *WebSocket) Close() error {
	return w.ws.Close()
}

func (w *WebSocket) readLoop() {
	// We largely ignore read messages but websockets require
	// us to handle ping/pong and close messages.
	defer w.Close()
	defer close(w.closed)
	w.ws.SetReadLimit(512)
	w.ws.SetReadDeadline(time.Now().Add(60 * time.Second))
	w.ws.SetPongHandler(func(string) error {
		w.ws.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	for {
		_, _, err := w.ws.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (w *WebSocket) ping() {
	ticker := time.NewTicker(50 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := w.ws.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second)); err != nil {
				log.Println("ping:", err)
			}
		case <-w.closed:
			return
		}
	}
}

func (w *WebSocket) Start(changed <-chan struct{}) {
	go w.readLoop()
	go w.ping()
	for {
		select {
		case <-w.closed:
			return
		case <-changed:
			err := w.ws.WriteMessage(websocket.TextMessage, []byte("reload"))
			if err != nil {
				log.Println(err)
				return
			}
		}
	}
}
