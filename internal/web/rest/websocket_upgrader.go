package rest

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	"github.com/go-risk-it/go-risk-it/internal/web/ws/connection"
	"go.uber.org/zap"
)

type WebSocketUpgraderHandler interface {
	route.Route
}

type WebSocketUpgraderHandlerImpl struct {
	connectionManager connection.Manager
	upgrader          connection.Upgrader
	log               *zap.SugaredLogger
}

var _ WebSocketUpgraderHandler = (*WebSocketUpgraderHandlerImpl)(nil)

func NewWebSocketHandler(
	connectionManager connection.Manager,
	upgrader connection.Upgrader,
	log *zap.SugaredLogger,
) *WebSocketUpgraderHandlerImpl {
	return &WebSocketUpgraderHandlerImpl{
		connectionManager: connectionManager,
		upgrader:          upgrader,
		log:               log,
	}
}

func (h *WebSocketUpgraderHandlerImpl) Pattern() string {
	return "/wss"
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

	gameID, err := extractGameID(request)
	if err != nil {
		http.Error(writer, "unable to extract gameID from query parameters", http.StatusBadRequest)

		return
	}

	gameContext := ctx.WithGameID(userContext, gameID)

	conn, err := h.upgrader.Upgrade(writer, request, nil)
	if err != nil {
		http.Error(
			writer,
			"unable to upgrade websocket connection",
			http.StatusInternalServerError,
		)

		return
	}

	h.connectionManager.ConnectPlayer(gameContext, conn)

	h.log.Infow("Upgraded:", "remoteAddress", conn.RemoteAddr().String(), "gameID", gameID)
}

func extractGameID(r *http.Request) (int64, error) {
	gameID := r.URL.Query().Get("gameID")
	if gameID == "" {
		return -1, fmt.Errorf("no parameter matching gameID: %s", r.URL.Query())
	}

	gameIDInt, err := strconv.ParseInt(gameID, 10, 64)
	if err != nil {
		return -1, fmt.Errorf("unable to parse gameID: %w", err)
	}

	return gameIDInt, nil
}
