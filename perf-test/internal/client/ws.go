package client

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/gamestate"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/metrics"
	"github.com/gorilla/websocket"
)

const (
	pingInterval     = 30 * time.Second
	writeWait        = 10 * time.Second
	reconnectBase    = 500 * time.Millisecond
	reconnectMax     = 15 * time.Second
	reconnectRetries = 5
)

// WS handles WebSocket connections to the game server.
type WS struct {
	mu   sync.Mutex
	conn *websocket.Conn

	// Connection parameters for reconnection.
	wsURL  string
	header http.Header

	view *gamestate.View
	done chan struct{}

	// closed is set to true when Close() is called intentionally.
	closed bool

	// pingStop stops the current connection's ping goroutine.
	pingStop chan struct{}

	collector *metrics.Collector
}

// ConnectWS establishes a WebSocket connection to a game.
func ConnectWS(
	baseURL string,
	gameID int64,
	token string,
	collector *metrics.Collector,
) (*WS, error) {
	wsURL := fmt.Sprintf("%s/ws?gameID=%d", baseURL, gameID)

	header := http.Header{}
	header.Set("Authorization", "Bearer "+token)

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
	if err != nil {
		return nil, fmt.Errorf("websocket dial: %w", err)
	}

	ws := &WS{
		conn:      conn,
		wsURL:     wsURL,
		header:    header,
		view:      gamestate.NewView(),
		done:      make(chan struct{}),
		pingStop:  make(chan struct{}),
		collector: collector,
	}

	go ws.readLoop()
	go ws.pingLoop(ws.pingStop)

	return ws, nil
}

// View returns the game state view updated by this connection.
func (ws *WS) View() *gamestate.View {
	return ws.view
}

// Done returns a channel closed when the connection is permanently terminated.
func (ws *WS) Done() <-chan struct{} {
	return ws.done
}

// Close closes the WebSocket connection. Safe to call multiple times.
func (ws *WS) Close() error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if ws.closed {
		return nil
	}

	ws.closed = true
	close(ws.pingStop)

	return ws.conn.Close()
}

func (ws *WS) readLoop() {
	defer close(ws.done)

	for {
		_, data, err := ws.conn.ReadMessage()
		if err != nil {
			// Check if this was an intentional close.
			ws.mu.Lock()
			intentional := ws.closed
			ws.mu.Unlock()

			if intentional {
				return
			}

			if websocket.IsUnexpectedCloseError(
				err,
				websocket.CloseGoingAway,
				websocket.CloseNormalClosure,
			) {
				log.Printf("ws read error: %v", err)
			}

			// Attempt reconnection.
			if !ws.reconnect() {
				return
			}

			continue
		}

		var msg gamestate.WSMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Printf("ws unmarshal error: %v", err)

			continue
		}

		if err := ws.view.Apply(msg); err != nil {
			log.Printf("ws apply error: %v", err)
		}
	}
}

// reconnect attempts to re-establish the WebSocket connection with exponential backoff.
// Returns true if reconnection succeeded, false if all attempts exhausted.
func (ws *WS) reconnect() bool {
	ws.mu.Lock()
	if ws.closed {
		ws.mu.Unlock()

		return false
	}

	// Stop the old ping goroutine.
	close(ws.pingStop)
	ws.mu.Unlock()

	backoff := reconnectBase

	for attempt := range reconnectRetries {
		log.Printf("ws reconnecting (attempt %d/%d, backoff %v)",
			attempt+1, reconnectRetries, backoff)

		if ws.collector != nil {
			ws.collector.RecordReconnect()
		}

		time.Sleep(backoff)

		ws.mu.Lock()
		if ws.closed {
			ws.mu.Unlock()

			return false
		}

		ws.mu.Unlock()

		conn, _, err := websocket.DefaultDialer.Dial(ws.wsURL, ws.header)
		if err != nil {
			log.Printf("ws reconnect failed: %v", err)
			backoff *= 2
			if backoff > reconnectMax {
				backoff = reconnectMax
			}

			continue
		}

		// Swap in the new connection and restart ping.
		ws.mu.Lock()
		ws.conn = conn
		ws.pingStop = make(chan struct{})
		ws.mu.Unlock()

		go ws.pingLoop(ws.pingStop)
		log.Printf("ws reconnected successfully")

		return true
	}

	log.Printf("ws reconnect failed after %d attempts, giving up", reconnectRetries)

	if ws.collector != nil {
		ws.collector.RecordReconnectFailure()
	}

	return false
}

func (ws *WS) pingLoop(stop chan struct{}) {
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ws.mu.Lock()
			conn := ws.conn
			ws.mu.Unlock()

			if err := conn.WriteControl(
				websocket.PingMessage, nil, time.Now().Add(writeWait),
			); err != nil {
				return
			}
		case <-stop:
			return
		}
	}
}
