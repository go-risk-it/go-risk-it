package web

import (
	"net/http"

	"github.com/tomfran/go-risk-it/internal/web/handlers"
	"go.uber.org/zap"
)

func NewServeMux(wsHandler *handlers.WebSocketHandler, log *zap.SugaredLogger) *http.ServeMux {
	mux := &http.ServeMux{}
	mux.HandleFunc("/ws", wsHandler.OnWebSocket)
	log.Infow("Created mux", "mux", mux)

	return mux
}
