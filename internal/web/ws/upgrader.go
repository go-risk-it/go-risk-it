package ws

import (
	"net/http"

	"github.com/lesismal/nbio/nbhttp/websocket"
	"go.uber.org/zap"
)

type Upgrader interface {
	Upgrade(
		w http.ResponseWriter,
		r *http.Request,
		responseHeader http.Header,
		args ...interface{},
	) (*websocket.Conn, error)
}

type UpgraderImpl struct {
	*websocket.Upgrader
}

func New(log *zap.SugaredLogger, _ ...interface{}) *UpgraderImpl {
	//exhaustruct:ignore
	upgrader := UpgraderImpl{
		Upgrader: websocket.NewUpgrader(),
	}
	upgrader.Subprotocols = []string{"risk-it.websocket.auth.token"}

	upgrader.CheckOrigin = func(r *http.Request) bool {
		// plz fix
		return true
	}

	upgrader.OnOpen(func(connection *websocket.Conn) {
		log.Infow("Connection opened", "remoteAddress", connection.RemoteAddr().String())
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
