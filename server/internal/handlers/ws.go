package handlers

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"

	"github.com/traP-jp/h26s_03/server/internal/middleware/authx"
)

const (
	websocketMessageTypeReaction   = "reaction"
	websocketMessageTypeVote       = "vote"
	websocketMessageTypePollStatus = "poll_status"

	pollStatusClosed = "closed"
)

type websocketMessageEnvelope struct {
	Type string `json:"type"`
}

type pollStatusWebSocketMessage struct {
	Type       string    `json:"type"`
	PollID     string    `json:"poll_id"`
	Status     string    `json:"status"`
	Result     *int      `json:"result"`
	NotifiedAt time.Time `json:"notified_at"`
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

func (h *Handler) broadcastPollClosed(pollID string, result int) {
	if h.wsHub == nil {
		return
	}

	payload, err := json.Marshal(pollStatusWebSocketMessage{
		Type:       websocketMessageTypePollStatus,
		PollID:     pollID,
		Status:     pollStatusClosed,
		Result:     &result,
		NotifiedAt: time.Now().UTC(),
	})
	if err != nil {
		return
	}

	h.wsHub.broadcastToPoll(pollID, websocket.TextMessage, payload)
}

func (h *Handler) WebSocket(c echo.Context) error {
	pollID := c.QueryParam("poll_id")
	if pollID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "poll_id is required")
	}
	username, ok := authx.UserFromRequestContext(c.Request().Context())
	if !ok {
		username = anonymousUser
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
			message.Username = username
			payload, err := json.Marshal(message)
			if err != nil {
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
			message.Username = username
			payload, err := json.Marshal(message)
			if err != nil {
				return nil
			}
			h.wsHub.broadcastToPoll(pollID, messageType, payload)
		default:
			return nil
		}
	}
}
