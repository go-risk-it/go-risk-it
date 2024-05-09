package rest

import (
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/web/ws/connection"
	"go.uber.org/zap"
)

type WebSocketUpgraderHandler interface {
	Pattern() string
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

type WebSocketUpgraderHandlerImpl struct {
	upgrader connection.Upgrader
	log      *zap.SugaredLogger
}

func NewWebSocketHandler(
	upgrader connection.Upgrader,
	log *zap.SugaredLogger,
) *WebSocketUpgraderHandlerImpl {
	return &WebSocketUpgraderHandlerImpl{upgrader: upgrader, log: log}
}

func (h *WebSocketUpgraderHandlerImpl) Pattern() string {
	return "/ws"
}

func (h *WebSocketUpgraderHandlerImpl) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.log.Infow("Received request")

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.log.Errorw("Unable to upgrade connection", "error", err)
	}

	h.log.Infow("Upgraded:", "remoteAddress", conn.RemoteAddr().String())
}
