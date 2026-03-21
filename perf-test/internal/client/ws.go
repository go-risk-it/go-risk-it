package client

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/gamestate"
	"github.com/gorilla/websocket"
)

const (
	pingInterval = 30 * time.Second
	writeWait    = 10 * time.Second
)

// WS handles WebSocket connections to the game server.
type WS struct {
	conn *websocket.Conn
	view *gamestate.View
	done chan struct{}
}

// ConnectWS establishes a WebSocket connection to a game.
func ConnectWS(baseURL string, gameID int64, token string) (*WS, error) {
	wsURL := fmt.Sprintf("%s/ws?gameID=%d", baseURL, gameID)

	header := http.Header{}
	header.Set("Authorization", "Bearer "+token)

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
	if err != nil {
		return nil, fmt.Errorf("websocket dial: %w", err)
	}

	ws := &WS{
		conn: conn,
		view: gamestate.NewView(),
		done: make(chan struct{}),
	}

	go ws.readLoop()
	go ws.pingLoop()

	return ws, nil
}

// View returns the game state view updated by this connection.
func (ws *WS) View() *gamestate.View {
	return ws.view
}

// Done returns a channel closed when the connection is terminated.
func (ws *WS) Done() <-chan struct{} {
	return ws.done
}

// Close closes the WebSocket connection.
func (ws *WS) Close() error {
	return ws.conn.Close()
}

func (ws *WS) readLoop() {
	defer close(ws.done)

	for {
		_, data, err := ws.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(
				err,
				websocket.CloseGoingAway,
				websocket.CloseNormalClosure,
			) {
				log.Printf("ws read error: %v", err)
			}

			return
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

func (ws *WS) pingLoop() {
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := ws.conn.WriteControl(
				websocket.PingMessage, nil, time.Now().Add(writeWait),
			); err != nil {
				return
			}
		case <-ws.done:
			return
		}
	}
}
