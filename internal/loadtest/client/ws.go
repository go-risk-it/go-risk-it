package client

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/loadtest/gamestate"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/metrics"
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

	// gen is the connection generation counter. Each (readLoop, pingLoop) pair
	// is scoped to a generation. Incremented on successful reconnect.
	gen uint64

	// genDone signals the current generation's pingLoop to stop.
	// Closed by reconnect (to stop old pingLoop) or by Close (to stop all).
	genDone chan struct{}

	collector *metrics.Collector
}

// ConnectWS establishes a WebSocket connection to a game.
func ConnectWS(
	baseURL string,
	gameID int64,
	token string,
	collector *metrics.Collector,
) (*WS, error) {
	wsURL := fmt.Sprintf("%s/api/v1/games/%d/ws", baseURL, gameID)

	header := http.Header{}
	header.Set("Authorization", "Bearer "+token)

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
	if err != nil {
		return nil, fmt.Errorf("websocket dial: %w", err)
	}

	genDone := make(chan struct{})

	ws := &WS{
		conn:      conn,
		wsURL:     wsURL,
		header:    header,
		view:      gamestate.NewView(),
		done:      make(chan struct{}),
		genDone:   genDone,
		collector: collector,
	}

	go ws.readLoop(0, genDone, conn)
	go ws.pingLoop(genDone, conn)

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

// Disrupt force-closes the underlying connection without marking the WS as
// intentionally closed. This triggers the readLoop's auto-reconnect logic.
func (ws *WS) Disrupt() {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if ws.closed {
		return
	}

	// Close the underlying conn; readLoop will get a read error and reconnect.
	ws.conn.Close()
}

// Close closes the WebSocket connection. Safe to call multiple times.
func (ws *WS) Close() error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if ws.closed {
		return nil
	}

	ws.closed = true
	close(ws.genDone)

	return ws.conn.Close()
}

// readLoop reads messages from the WebSocket connection. Each readLoop
// invocation is scoped to a connection generation: it takes its own
// generation number, genDone channel, and connection reference as parameters
// and never reads ws.conn directly.
func (ws *WS) readLoop(myGen uint64, myGenDone <-chan struct{}, myConn *websocket.Conn) {
	shouldCloseDone := true

	defer func() {
		if shouldCloseDone {
			close(ws.done)
		}
	}()

	for {
		_, data, err := myConn.ReadMessage()
		if err != nil {
			ws.mu.Lock()
			if ws.closed {
				ws.mu.Unlock()

				return // shouldCloseDone stays true — permanent shutdown.
			}

			if myGen < ws.gen {
				// This readLoop was superseded by a newer generation.
				// The successor owns done.
				ws.mu.Unlock()
				shouldCloseDone = false

				return
			}

			ws.mu.Unlock()

			if websocket.IsUnexpectedCloseError(
				err,
				websocket.CloseGoingAway,
				websocket.CloseNormalClosure,
			) {
				log.Printf("ws read error (gen %d): %v", myGen, err)
			}

			// Attempt reconnection. If it succeeds, a new readLoop is
			// launched by reconnect — this one must exit without closing done.
			if ws.reconnect(myGen) {
				shouldCloseDone = false

				return
			}

			// Reconnect exhausted or closed during backoff.
			return
		}

		if ws.view == nil {
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

// reconnect attempts to re-establish the WebSocket connection with exponential
// backoff. If successful, it spawns a new readLoop and pingLoop for the new
// generation. Returns true if reconnection succeeded.
func (ws *WS) reconnect(fromGen uint64) bool {
	ws.mu.Lock()

	if ws.closed {
		ws.mu.Unlock()

		return false
	}

	if fromGen != ws.gen {
		// Another goroutine already reconnected — this call is stale.
		ws.mu.Unlock()

		return false
	}

	// Stop the old generation's pingLoop and immediately replace genDone
	// with a fresh channel so Close() never sees a closed channel.
	close(ws.genDone)
	ws.genDone = make(chan struct{})
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

		// Swap in the new connection and start a new generation.
		ws.mu.Lock()
		if ws.closed {
			conn.Close()
			ws.mu.Unlock()

			return false
		}

		ws.gen++
		ws.conn = conn

		newGenDone := make(chan struct{})
		ws.genDone = newGenDone
		newGen := ws.gen

		ws.mu.Unlock()

		go ws.readLoop(newGen, newGenDone, conn)
		go ws.pingLoop(newGenDone, conn)

		log.Printf("ws reconnected successfully (gen %d)", newGen)

		return true
	}

	log.Printf("ws reconnect failed after %d attempts, giving up", reconnectRetries)

	if ws.collector != nil {
		ws.collector.RecordReconnectFailure()
	}

	return false
}

func (ws *WS) pingLoop(myGenDone <-chan struct{}, myConn *websocket.Conn) {
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := myConn.WriteControl(
				websocket.PingMessage, nil, time.Now().Add(writeWait),
			); err != nil {
				log.Printf("ws ping failed, forcing reconnect: %v", err)
				myConn.Close()

				return
			}
		case <-myGenDone:
			return
		}
	}
}
