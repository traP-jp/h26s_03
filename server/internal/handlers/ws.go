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
	conns map[*websocket.Conn]string
}

func newWebsocketHub() *websocketHub {
	return &websocketHub{conns: make(map[*websocket.Conn]string)}
}

func (h *websocketHub) add(conn *websocket.Conn, pollID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.conns[conn] = pollID
}

func (h *websocketHub) remove(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.conns, conn)
}

func (h *websocketHub) broadcastToPoll(pollID string, messageType int, payload []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for conn, connPollID := range h.conns {
		if connPollID != pollID {
			continue
		}
		_ = conn.WriteMessage(messageType, payload)
	}
}

var websocketUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (h *Handler) WebSocket(c echo.Context) error {
	pollID := c.QueryParam("poll_id")
	if pollID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "poll_id is required")
	}

	conn, err := websocketUpgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to upgrade to websocket: "+err.Error())
	}
	defer func() {
		h.wsHub.remove(conn)
		_ = conn.Close()
	}()

	h.wsHub.add(conn, pollID)

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
			h.wsHub.broadcastToPoll(pollID, messageType, payload)
		case websocketMessageTypeVote:
			var message voteWebSocketMessage
			if err := json.Unmarshal(payload, &message); err != nil {
				return nil
			}
			if err := message.validate(); err != nil {
				return nil
			}
			h.wsHub.broadcastToPoll(pollID, messageType, payload)
		default:
			return nil
		}
	}
}
