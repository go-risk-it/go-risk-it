package rest

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	domainerrors "github.com/go-risk-it/go-risk-it/internal/logic/errors"
	gameWs "github.com/go-risk-it/go-risk-it/internal/web/game/ws"
	lobbyWs "github.com/go-risk-it/go-risk-it/internal/web/lobby/ws"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	restutils "github.com/go-risk-it/go-risk-it/internal/web/rest/utils"
	"github.com/go-risk-it/go-risk-it/internal/web/ws"
)

func NewWebSocketHandler(
	gameConnectionManager gameWs.Manager,
	lobbyConnectionManager lobbyWs.Manager,
	upgrader ws.Upgrader,
) *route.Route {
	handler := &webSocketHandler{
		gameConnectionManager:  gameConnectionManager,
		lobbyConnectionManager: lobbyConnectionManager,
		upgrader:               upgrader,
	}

	return route.New("/ws", true, handler)
}

type webSocketHandler struct {
	gameConnectionManager  gameWs.Manager
	lobbyConnectionManager lobbyWs.Manager
	upgrader               ws.Upgrader
}

func (h *webSocketHandler) ServeHTTP(
	writer http.ResponseWriter,
	request *http.Request,
) {
	slog.InfoContext(request.Context(), "Received request")

	userContext, ok := request.Context().(ctx.UserContext)
	if !ok {
		_ = restutils.WriteError(
			writer,
			errors.New("unable to extract user context"),
		)

		return
	}

	gameID, lobbyID, err := extractConnectionParams(request)
	if err != nil {
		_ = restutils.WriteError(
			writer,
			domainerrors.WrapValidationError(err, "invalid connection parameters"),
		)

		return
	}

	conn, err := h.upgrader.Upgrade(writer, request, nil)
	if err != nil {
		_ = restutils.WriteError(
			writer,
			fmt.Errorf("unable to upgrade websocket connection: %w", err),
		)

		return
	}

	if gameID > 0 {
		h.gameConnectionManager.ConnectPlayer(ctx.WithGameID(userContext, gameID), conn)
	} else {
		h.lobbyConnectionManager.ConnectPlayer(ctx.WithLobbyID(userContext, lobbyID), conn)
	}

	slog.InfoContext(request.Context(), "Upgraded", "remoteAddress", conn.RemoteAddr().String())
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
