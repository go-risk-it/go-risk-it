package web

import (
	"net/http"

	"github.com/tomfran/go-risk-it/internal/web/rest"
	"github.com/tomfran/go-risk-it/internal/web/ws/connection"
	"go.uber.org/zap"
)

func NewServeMux(wsHandler *connection.WebSocketUpgraderHandler,
	restHandler rest.Handler,
	log *zap.SugaredLogger,
) *http.ServeMux {
	mux := &http.ServeMux{}
	mux.HandleFunc("/ws", wsHandler.OnWebSocket)
	mux.HandleFunc("POST /api/1/game", restHandler.OnCreateGame)
	mux.HandleFunc("POST /api/1/game/{id}/move/deploy", restHandler.OnMoveDeploy)
	log.Infow("Created mux", "mux", mux)

	return mux
}
