package web

import (
	"net/http"

	"github.com/tomfran/go-risk-it/internal/web/ws/connection/handler"
	"go.uber.org/zap"
)

func NewServeMux(wsHandler *handler.WebSocketHandler, log *zap.SugaredLogger) *http.ServeMux {
	mux := &http.ServeMux{}
	mux.HandleFunc("/ws", wsHandler.OnWebSocket)
	log.Infow("Created mux", "mux", mux)

	return mux
}
