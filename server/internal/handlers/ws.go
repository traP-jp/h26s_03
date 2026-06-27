package handlers

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

const (
	websocketMessageTypeReaction = "reaction"
	websocketMessageTypeVote     = "vote"
)

type websocketMessageEnvelope struct {
	Type string `json:"type"`
}

type websocketHub struct {
	mu    sync.Mutex
	conns map[*websocket.Conn]struct{}
}

func newWebsocketHub() *websocketHub {
	return &websocketHub{conns: make(map[*websocket.Conn]struct{})}
}

func (h *websocketHub) add(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.conns[conn] = struct{}{}
}

func (h *websocketHub) remove(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.conns, conn)
}

func (h *websocketHub) broadcast(messageType int, payload []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for conn := range h.conns {
		_ = conn.WriteMessage(messageType, payload)
	}
}

var websocketUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (h *Handler) WebSocket(c echo.Context) error {
	conn, err := websocketUpgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to upgrade to websocket: "+err.Error())
	}
	defer func() {
		h.wsHub.remove(conn)
		_ = conn.Close()
	}()

	h.wsHub.add(conn)

	for {
		messageType, payload, err := conn.ReadMessage()
		if err != nil {
			return nil
		}

		var envelope websocketMessageEnvelope
		if err := json.Unmarshal(payload, &envelope); err != nil {
			return nil
		}

		switch envelope.Type {
		case websocketMessageTypeReaction:
			var message reactionWebSocketMessage
			if err := json.Unmarshal(payload, &message); err != nil {
				return nil
			}
			if err := message.validate(); err != nil {
				return nil
			}
			h.wsHub.broadcast(messageType, payload)
		case websocketMessageTypeVote:
			var message voteWebSocketMessage
			if err := json.Unmarshal(payload, &message); err != nil {
				return nil
			}
			if err := message.validate(); err != nil {
				return nil
			}
			h.wsHub.broadcast(messageType, payload)
		default:
			return nil
		}
	}
}
