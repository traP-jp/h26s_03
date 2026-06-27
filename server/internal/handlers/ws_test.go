package handlers_test

import (
	"errors"
	"fmt"
	"net"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"

	"github.com/traP-jp/h26s_03/server/internal/handlers"
)

func TestWebSocketBroadcasts(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		payload string
	}{
		{
			name:    "reaction",
			payload: `{"type":"reaction","username":"alice","reaction":"like"}`,
		},
		{
			name:    "vote",
			payload: `{"type":"vote"}`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			baseURL := startWebSocketTestServer(t)
			poll1Sender := dialWebSocket(t, baseURL, "1")
			poll1Viewer := dialWebSocket(t, baseURL, "1")

			if err := poll1Sender.WriteMessage(websocket.TextMessage, []byte(tc.payload)); err != nil {
				t.Fatalf("write websocket message: %v", err)
			}

			assertWebSocketMessage(t, poll1Sender, tc.payload)
			assertWebSocketMessage(t, poll1Viewer, tc.payload)
		})
	}
}

func TestWebSocketBroadcastsOnlyToPollSubscribers(t *testing.T) {
	t.Parallel()

	baseURL := startWebSocketTestServer(t)
	poll1Sender := dialWebSocket(t, baseURL, "1")
	poll1Viewer := dialWebSocket(t, baseURL, "1")
	poll2Viewer := dialWebSocket(t, baseURL, "2")

	payload := `{"type":"reaction","username":"alice","reaction":"like"}`
	if err := poll1Sender.WriteMessage(websocket.TextMessage, []byte(payload)); err != nil {
		t.Fatalf("write websocket message: %v", err)
	}

	assertWebSocketMessage(t, poll1Sender, payload)
	assertWebSocketMessage(t, poll1Viewer, payload)
	assertNoWebSocketMessage(t, poll2Viewer)
}

func TestWebSocketRequiresPollID(t *testing.T) {
	t.Parallel()

	baseURL := startWebSocketTestServer(t)
	_, resp, err := websocket.DefaultDialer.Dial(fmt.Sprintf("%s/api/ws", baseURL), nil)
	if err == nil {
		t.Fatal("unexpected successful websocket dial without poll_id")
	}
	if resp == nil {
		t.Fatalf("missing websocket response: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Fatalf("unexpected status: got=%d want=400", resp.StatusCode)
	}
}

func TestWebSocketDoesNotBroadcastInvalidMessages(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		payload string
	}{
		{
			name:    "reaction missing reaction",
			payload: `{"type":"reaction","username":"alice"}`,
		},
		{
			name:    "unknown type",
			payload: `{"type":"unknown","username":"alice"}`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			baseURL := startWebSocketTestServer(t)
			sender := dialWebSocket(t, baseURL, "1")
			viewer := dialWebSocket(t, baseURL, "1")

			if err := sender.WriteMessage(websocket.TextMessage, []byte(tc.payload)); err != nil {
				t.Fatalf("write websocket message: %v", err)
			}

			assertNoWebSocketMessage(t, viewer)
		})
	}
}

func startWebSocketTestServer(t *testing.T) string {
	t.Helper()

	e := echo.New()
	h := handlers.New(nil)
	e.GET("/api/ws", h.WebSocket)

	srv := httptest.NewServer(e)
	t.Cleanup(srv.Close)

	return "ws" + strings.TrimPrefix(srv.URL, "http")
}

func dialWebSocket(t *testing.T, baseURL string, pollID string) *websocket.Conn {
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

func assertWebSocketMessage(t *testing.T, conn *websocket.Conn, want string) {
	t.Helper()

	if err := conn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatalf("set read deadline: %v", err)
	}
	_, payload, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read websocket message: %v", err)
	}
	if string(payload) != want {
		t.Fatalf("unexpected websocket message: got=%s want=%s", string(payload), want)
	}
}

func assertNoWebSocketMessage(t *testing.T, conn *websocket.Conn) {
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
