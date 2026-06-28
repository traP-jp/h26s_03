package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"

	"github.com/traP-jp/h26s_03/server/internal/middleware/authx"
)

func TestBroadcastPollClosedBroadcastsToPollSubscribers(t *testing.T) {
	t.Parallel()

	handler, baseURL := startInternalWebSocketTestServer(t)
	poll42Viewer := dialInternalWebSocket(t, baseURL, "42")
	poll42AnotherViewer := dialInternalWebSocket(t, baseURL, "42")
	poll2Viewer := dialInternalWebSocket(t, baseURL, "2")

	handler.broadcastPollClosed("42", 1)

	assertPollClosedWebSocketMessage(t, poll42Viewer, 42)
	assertPollClosedWebSocketMessage(t, poll42AnotherViewer, 42)
	assertNoInternalWebSocketMessage(t, poll2Viewer)
}

func startInternalWebSocketTestServer(t *testing.T) (*Handler, string) {
	t.Helper()

	e := echo.New()
	e.Use(authx.Soft())
	handler := New(nil)
	e.GET("/api/ws", handler.WebSocket)

	srv := httptest.NewServer(e)
	t.Cleanup(srv.Close)

	return handler, "ws" + strings.TrimPrefix(srv.URL, "http")
}

func dialInternalWebSocket(t *testing.T, baseURL string, pollID string) *websocket.Conn {
	t.Helper()

	conn, resp, err := websocket.DefaultDialer.Dial(fmt.Sprintf("%s/api/ws?poll_id=%s", baseURL, pollID), nil)
	if err != nil {
		if resp != nil {
			t.Fatalf("dial websocket: status=%s err=%v", resp.Status, err)
		}
		t.Fatalf("dial websocket: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	return conn
}

func assertPollClosedWebSocketMessage(t *testing.T, conn *websocket.Conn, pollID int64) {
	t.Helper()

	if err := conn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatalf("set read deadline: %v", err)
	}

	messageType, payload, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read websocket message: %v", err)
	}
	if messageType != websocket.TextMessage {
		t.Fatalf("unexpected websocket message type: got=%d want=%d", messageType, websocket.TextMessage)
	}

	var message pollStatusWebSocketMessage
	if err := json.Unmarshal(payload, &message); err != nil {
		t.Fatalf("decode websocket message: %v", err)
	}
	if message.Type != websocketMessageTypePollStatus {
		t.Fatalf("unexpected message type: got=%s want=%s", message.Type, websocketMessageTypePollStatus)
	}
	wantPollID := fmt.Sprint(pollID)
	if message.PollID != wantPollID {
		t.Fatalf("unexpected poll_id: got=%s want=%s", message.PollID, wantPollID)
	}
	if message.Status != pollStatusClosed {
		t.Fatalf("unexpected status: got=%s want=%s", message.Status, pollStatusClosed)
	}
	if message.Result == nil || *message.Result != 1 {
		t.Fatalf("unexpected result: got=%v want=1", message.Result)
	}
	if message.NotifiedAt.IsZero() {
		t.Fatal("unexpected notified_at: zero")
	}
}

func assertNoInternalWebSocketMessage(t *testing.T, conn *websocket.Conn) {
	t.Helper()

	if err := conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond)); err != nil {
		t.Fatalf("set read deadline: %v", err)
	}

	_, payload, err := conn.ReadMessage()
	if err == nil {
		t.Fatalf("unexpected websocket message: %s", string(payload))
	}
	var netErr net.Error
	if !errors.As(err, &netErr) || !netErr.Timeout() {
		t.Fatalf("unexpected websocket read error: %v", err)
	}
}
