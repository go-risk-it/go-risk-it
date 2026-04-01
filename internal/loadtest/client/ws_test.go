package client

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testWSServer creates an httptest server with a gorilla/websocket upgrader.
// The handler is called for each accepted WS connection. Returns the server
// and a WS-scheme URL (ws://...) ready for dialing.
func testWSServer(t *testing.T, handler func(*websocket.Conn)) (*httptest.Server, string) {
	t.Helper()

	upgrader := websocket.Upgrader{
		CheckOrigin: func(_ *http.Request) bool { return true },
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Logf("upgrade error: %v", err)

			return
		}

		handler(conn)
	}))

	t.Cleanup(srv.Close)

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")

	return srv, wsURL
}

// connectTestWS dials the test server and returns a WS client.
// Uses minimal constants for fast test execution.
func connectTestWS(t *testing.T, wsURL string) *WS {
	t.Helper()

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)

	genDone := make(chan struct{})

	ws := &WS{
		conn:    conn,
		wsURL:   wsURL,
		header:  http.Header{},
		view:    nil, // tests don't need view processing
		done:    make(chan struct{}),
		genDone: genDone,
	}

	// Capture genDone locally before launching goroutines to avoid a race
	// between readLoop calling reconnect (which writes ws.genDone) and the
	// second go statement reading ws.genDone.
	go ws.readLoop(0, genDone, conn)
	go ws.pingLoop(genDone, conn)

	return ws
}

func TestWS_ReadLoop_ExitsOnClose(t *testing.T) {
	t.Parallel()

	_, wsURL := testWSServer(t, func(conn *websocket.Conn) {
		// Echo server — keeps connection alive until client closes.
		for {
			mt, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}

			_ = conn.WriteMessage(mt, msg)
		}
	})

	ws := connectTestWS(t, wsURL)

	// Close should terminate readLoop and eventually close done.
	err := ws.Close()
	require.NoError(t, err)

	select {
	case <-ws.Done():
		// Success — done was closed.
	case <-time.After(2 * time.Second):
		t.Fatal("done channel not closed after Close()")
	}
}

func TestWS_ReadLoop_ReconnectsOnError(t *testing.T) {
	t.Parallel()

	var connCount atomic.Int64

	_, wsURL := testWSServer(t, func(conn *websocket.Conn) {
		n := connCount.Add(1)
		if n == 1 {
			// First connection: close immediately to trigger reconnect.
			conn.Close()

			return
		}

		// Second connection: stay alive.
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				return
			}
		}
	})

	ws := connectTestWS(t, wsURL)
	defer ws.Close()

	// Wait for reconnection to establish a second connection.
	require.Eventually(t, func() bool {
		return connCount.Load() >= 2
	}, 5*time.Second, 50*time.Millisecond, "expected reconnection to establish second connection")

	// done should NOT be closed — reconnect succeeded.
	select {
	case <-ws.Done():
		t.Fatal("done channel should not be closed after successful reconnect")
	default:
		// Good — still open.
	}
}

func TestWS_PingStop_NoDoubleClose(t *testing.T) {
	t.Parallel()

	_, wsURL := testWSServer(t, func(conn *websocket.Conn) {
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				return
			}
		}
	})

	ws := connectTestWS(t, wsURL)

	// Disrupt triggers reconnect (closes conn, readLoop gets error,
	// reconnect closes genDone). Then Close also closes genDone.
	// This must not panic.
	ws.Disrupt()

	// Give reconnect a moment to start.
	time.Sleep(100 * time.Millisecond)

	// Close during or after reconnect — must not double-close genDone.
	assert.NotPanics(t, func() {
		_ = ws.Close()
	})

	select {
	case <-ws.Done():
		// Good.
	case <-time.After(5 * time.Second):
		t.Fatal("done channel not closed after Disrupt+Close")
	}
}

func TestWS_Close_DuringReconnectBackoff(t *testing.T) {
	t.Parallel()

	var connCount atomic.Int64

	_, wsURL := testWSServer(t, func(conn *websocket.Conn) {
		n := connCount.Add(1)
		if n == 1 {
			// Close immediately to trigger reconnect.
			conn.Close()

			return
		}

		// Subsequent connections: keep alive.
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				return
			}
		}
	})

	ws := connectTestWS(t, wsURL)

	// Wait for the first connection to close and reconnect to begin.
	time.Sleep(200 * time.Millisecond)

	// Close during backoff sleep.
	err := ws.Close()
	require.NoError(t, err)

	select {
	case <-ws.Done():
		// Good — done closed.
	case <-time.After(5 * time.Second):
		t.Fatal("done channel not closed after Close during reconnect backoff")
	}
}

func TestWS_ConcurrentDisruptAndClose(t *testing.T) {
	t.Parallel()

	_, wsURL := testWSServer(t, func(conn *websocket.Conn) {
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				return
			}
		}
	})

	ws := connectTestWS(t, wsURL)

	// Hammer both Disrupt and Close concurrently from multiple goroutines.
	var wg sync.WaitGroup

	const goroutines = 10

	wg.Add(goroutines)

	for range goroutines {
		go func() {
			defer wg.Done()

			ws.Disrupt()
		}()
	}

	// Also close concurrently.

	wg.Go(func() {
		_ = ws.Close()
	})

	wg.Wait()

	select {
	case <-ws.Done():
		// Good.
	case <-time.After(5 * time.Second):
		t.Fatal("done channel not closed after concurrent Disrupt+Close")
	}
}

func TestWS_ReconnectExhaustion_ClosesDone(t *testing.T) {
	t.Parallel()

	// Server accepts the first connection only. After that, it returns
	// HTTP 503 so reconnect dials fail at the WebSocket upgrade level.
	var connCount atomic.Int64

	upgrader := websocket.Upgrader{
		CheckOrigin: func(_ *http.Request) bool { return true },
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := connCount.Add(1)
		if n > 1 {
			// Reject reconnect attempts.
			w.WriteHeader(http.StatusServiceUnavailable)

			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		// Close immediately to trigger readLoop error.
		conn.Close()
	}))

	t.Cleanup(srv.Close)

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	ws := connectTestWS(t, wsURL)

	// All reconnect attempts will fail because the server rejects them.
	// done should be closed after exhaustion.
	select {
	case <-ws.Done():
		// Good — done closed after reconnect exhaustion.
	case <-time.After(30 * time.Second):
		t.Fatal("done channel not closed after reconnect exhaustion")
	}
}
