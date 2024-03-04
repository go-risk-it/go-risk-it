package handler

import (
	"net/http"

	"github.com/tomfran/go-risk-it/internal/web/ws/connection/upgrader"
	"go.uber.org/zap"
)

type WebSocketHandler struct {
	upgrader upgrader.Upgrader
	log      *zap.SugaredLogger
}

func NewWebSocketHandler(
	upgrader upgrader.Upgrader,
	logger *zap.SugaredLogger,
) *WebSocketHandler {
	return &WebSocketHandler{upgrader: upgrader, log: logger}
}

func (h *WebSocketHandler) OnWebSocket(w http.ResponseWriter, r *http.Request) {
	h.log.Infow("Received request")

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		panic(err)
	}

	h.log.Infow("Upgraded:", "remoteAddress", conn.RemoteAddr().String())
}
