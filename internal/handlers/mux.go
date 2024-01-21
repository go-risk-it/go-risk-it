package handlers

import (
	"net/http"

	"go.uber.org/zap"
)

func NewServeMux(wsHandler *WebSocketHandler, log *zap.SugaredLogger) *http.ServeMux {
	mux := &http.ServeMux{}
	mux.HandleFunc("/ws", wsHandler.OnWebSocket)
	log.Infow("Created mux", "mux", mux)
	return mux
}
