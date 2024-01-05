package nbio

import (
	"go-risk-it/handlers"
	"go.uber.org/zap"
	"net/http"
)

func NewServeMux(wsHandler *handlers.WebSocketHandler, log *zap.SugaredLogger) *http.ServeMux {
	mux := &http.ServeMux{}
	mux.HandleFunc("/ws", wsHandler.OnWebSocket)
	log.Infow("Created mux", "mux", mux)
	return mux
}
