package upgrader

import (
	"net/http"

	"github.com/lesismal/nbio/nbhttp/websocket"
	"github.com/tomfran/go-risk-it/internal/web/ws/connection/manager"
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
	connectionManager manager.Manager
	messageHandler    message.Handler
}

func New(
	log *zap.SugaredLogger,
	messageHandler message.Handler,
	connectionManager manager.Manager,
	args ...interface{},
) *UpgraderImpl {
	//exhaustruct:ignore
	upgrader := UpgraderImpl{
		Upgrader: &websocket.Upgrader{
			// resolve cross-origin problems
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		log:               log,
		connectionManager: connectionManager,
		messageHandler:    messageHandler,
	}

	upgrader.OnOpen(func(connection *websocket.Conn) {
		connectionManager.ConnectPlayer(connection, 0)
	})

	upgrader.OnMessage(messageHandler.OnMessage)

	upgrader.OnClose(func(connection *websocket.Conn, err error) {
		if err != nil {
			log.Errorw(
				"Connection closed",
				"remoteAddress",
				connection.RemoteAddr().String(),
				"error",
				err,
			)
		} else {
			log.Infow("Connection closed", "remoteAddress", connection.RemoteAddr().String())
		}
		connectionManager.DisconnectPlayer(connection, 0)
	})

	return &upgrader
}
