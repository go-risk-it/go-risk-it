package ws

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/kernel/config"
	"github.com/lesismal/nbio/nbhttp/websocket"
)

const (
	pingInterval = 10 * time.Second
)

type Upgrader interface {
	Upgrade(
		w http.ResponseWriter,
		r *http.Request,
		responseHeader http.Header,
		args ...any,
	) (*websocket.Conn, error)
}

type upgrader struct {
	*websocket.Upgrader
}

var _ Upgrader = (*upgrader)(nil)

func New(serverConfig config.ServerConfig, _ ...any) Upgrader {
	wsUpgrader := upgrader{
		Upgrader: websocket.NewUpgrader(),
	}
	wsUpgrader.Subprotocols = []string{"risk-it.websocket.auth.token"}

	allowedOrigins := make(map[string]bool, len(serverConfig.AllowedOrigins))
	for _, origin := range serverConfig.AllowedOrigins {
		allowedOrigins[origin] = true
	}

	wsUpgrader.CheckOrigin = func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}

		return allowedOrigins[origin]
	}

	wsUpgrader.OnOpen(func(connection *websocket.Conn) {
		slog.Info("connection opened", "remoteAddress", connection.RemoteAddr().String())
		connection.Keepalive(pingInterval)
	})

	wsUpgrader.SetPingHandler(func(conn *websocket.Conn, data string) {
		_ = conn.WriteMessage(websocket.PongMessage, []byte(data))
	})

	wsUpgrader.OnMessage(nil)

	wsUpgrader.OnClose(func(connection *websocket.Conn, err error) {
		if err != nil {
			slog.Info("connection closed with error", "error", err)
		} else {
			slog.Info("connection closed", "remoteAddress", connection.RemoteAddr().String())
		}
	})

	return &wsUpgrader
}
