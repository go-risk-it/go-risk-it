package handlers

import (
	"net/http"

	"github.com/lesismal/nbio/nbhttp/websocket"
	"go.uber.org/zap"
)

type WebSocketHandler struct {
	upgrader *websocket.Upgrader
	log      *zap.SugaredLogger
}

func NewWebSocketHandler(
	upgrader *websocket.Upgrader,
	logger *zap.SugaredLogger,
) *WebSocketHandler {
	return &WebSocketHandler{upgrader: upgrader, log: logger}
}

func (wsHandler *WebSocketHandler) OnWebSocket(w http.ResponseWriter, r *http.Request) {
	wsHandler.log.Infow("Received request")

	conn, err := wsHandler.upgrader.Upgrade(w, r, nil)
	if err != nil {
		panic(err)
	}

	wsHandler.log.Infow("Upgraded:", "remoteAddress", conn.RemoteAddr().String())
}
