package ws

import (
	"net/http"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/config"
	"github.com/lesismal/nbio/nbhttp/websocket"
	"go.uber.org/zap"
)

const (
	// pingInterval is how often the server sends a WebSocket ping to each client.
	// Must be well below the idle timeout (30s in nbio's default task pool) to
	// prevent the close/handle race that causes nil-pointer panics.
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

type UpgraderImpl struct {
	*websocket.Upgrader
}

func New(log *zap.SugaredLogger, serverConfig config.ServerConfig, _ ...any) *UpgraderImpl {
	//exhaustruct:ignore
	upgrader := UpgraderImpl{
		Upgrader: websocket.NewUpgrader(),
	}
	upgrader.Subprotocols = []string{"risk-it.websocket.auth.token"}

	allowedOrigins := make(map[string]bool, len(serverConfig.AllowedOrigins))
	for _, origin := range serverConfig.AllowedOrigins {
		allowedOrigins[origin] = true
	}

	upgrader.CheckOrigin = func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}

		return allowedOrigins[origin]
	}

	upgrader.OnOpen(func(connection *websocket.Conn) {
		log.Infow("Connection opened", "remoteAddress", connection.RemoteAddr().String())
		connection.Keepalive(pingInterval)
	})

	upgrader.SetPingHandler(func(conn *websocket.Conn, data string) {
		_ = conn.WriteMessage(websocket.PongMessage, []byte(data))
	})

	upgrader.OnMessage(nil)

	upgrader.OnClose(func(connection *websocket.Conn, err error) {
		if err != nil {
			log.Infow("Connection closed with error", "error", err)
		} else {
			log.Infow("Connection closed", "remoteAddress", connection.RemoteAddr().String())
		}
	})

	return &upgrader
}
