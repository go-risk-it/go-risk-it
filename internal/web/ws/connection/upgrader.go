package connection

import (
	"net/http"

	"github.com/lesismal/nbio/nbhttp/websocket"
	"github.com/tomfran/go-risk-it/internal/web/ws/message"
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
	log               *zap.SugaredLogger
	connectionManager Manager
	messageHandler    message.Handler
}

func New(
	log *zap.SugaredLogger,
	messageHandler message.Handler,
	connectionManager Manager,
	args ...interface{},
) *UpgraderImpl {
	//exhaustruct:ignore
	upgrader := UpgraderImpl{
		Upgrader:          websocket.NewUpgrader(),
		log:               log,
		connectionManager: connectionManager,
		messageHandler:    messageHandler,
	}

	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}

	upgrader.OnOpen(func(connection *websocket.Conn) {
		connectionManager.ConnectPlayer(connection, 1)
	})

	upgrader.OnMessage(messageHandler.OnMessage)

	upgrader.OnClose(func(connection *websocket.Conn, err error) {
		log.Infow("Connection closed", "remoteAddress", connection.RemoteAddr().String())
		connectionManager.DisconnectPlayer(connection, 1)
	})

	return &upgrader
}
