package rest

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	gameWs "github.com/go-risk-it/go-risk-it/internal/web/game/ws"
	lobbyWs "github.com/go-risk-it/go-risk-it/internal/web/lobby/ws"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	"github.com/go-risk-it/go-risk-it/internal/web/ws"
	"go.uber.org/zap"
)

type WebSocketUpgraderHandler interface {
	route.Route
}

type WebSocketUpgraderHandlerImpl struct {
	gameConnectionManager  gameWs.Manager
	lobbyConnectionManager lobbyWs.Manager
	upgrader               ws.Upgrader
	log                    *zap.SugaredLogger
}

var _ WebSocketUpgraderHandler = (*WebSocketUpgraderHandlerImpl)(nil)

func NewWebSocketHandler(
	gameConnectionManager gameWs.Manager,
	lobbyConnectionManager lobbyWs.Manager,
	upgrader ws.Upgrader,
	log *zap.SugaredLogger,
) *WebSocketUpgraderHandlerImpl {
	return &WebSocketUpgraderHandlerImpl{
		gameConnectionManager:  gameConnectionManager,
		lobbyConnectionManager: lobbyConnectionManager,
		upgrader:               upgrader,
		log:                    log,
	}
}

func (h *WebSocketUpgraderHandlerImpl) Pattern() string {
	return "/ws"
}

func (h *WebSocketUpgraderHandlerImpl) RequiresAuth() bool {
	return true
}

func (h *WebSocketUpgraderHandlerImpl) ServeHTTP(
	writer http.ResponseWriter,
	request *http.Request,
) {
	h.log.Infow("Received request")

	userContext, ok := request.Context().(ctx.UserContext)
	if !ok {
		http.Error(writer, "unable to extract user context", http.StatusInternalServerError)

		return
	}

	gameID, lobbyID, err := extractConnectionParams(request)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)

		return
	}

	conn, err := h.upgrader.Upgrade(writer, request, nil)
	if err != nil {
		http.Error(
			writer,
			"unable to upgrade websocket connection",
			http.StatusInternalServerError,
		)

		return
	}

	if gameID > 0 {
		h.gameConnectionManager.ConnectPlayer(ctx.WithGameID(userContext, gameID), conn)
	} else {
		h.lobbyConnectionManager.ConnectPlayer(ctx.WithLobbyID(userContext, lobbyID), conn)
	}

	h.log.Infow("Upgraded:", "remoteAddress", conn.RemoteAddr().String())
}

func extractConnectionParams(r *http.Request) (int64, int64, error) {
	query := r.URL.Query()
	gameIDStr := query.Get("gameID")
	lobbyIDStr := query.Get("lobbyID")

	if gameIDStr != "" && lobbyIDStr != "" {
		return 0, 0, errors.New("only one of gameID or lobbyID should be provided")
	}

	if gameIDStr == "" && lobbyIDStr == "" {
		return 0, 0, errors.New("either gameID or lobbyID must be provided")
	}

	if gameIDStr != "" {
		gameID, err := strconv.ParseInt(gameIDStr, 10, 64)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid gameID format: %w", err)
		}

		return gameID, 0, nil
	}

	lobbyID, err := strconv.ParseInt(lobbyIDStr, 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid lobbyID format: %w", err)
	}

	return 0, lobbyID, nil
}
