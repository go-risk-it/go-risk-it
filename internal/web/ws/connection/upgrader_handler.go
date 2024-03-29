package connection

import (
	"net/http"

	"go.uber.org/zap"
)

type WebSocketUpgraderHandler struct {
	upgrader Upgrader
	log      *zap.SugaredLogger
}

func NewWebSocketHandler(
	upgrader Upgrader,
	logger *zap.SugaredLogger,
) *WebSocketUpgraderHandler {
	return &WebSocketUpgraderHandler{upgrader: upgrader, log: logger}
}

func (h *WebSocketUpgraderHandler) OnWebSocket(w http.ResponseWriter, r *http.Request) {
	h.log.Infow("Received request")

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		panic(err)
	}

	h.log.Infow("Upgraded:", "remoteAddress", conn.RemoteAddr().String())
}
