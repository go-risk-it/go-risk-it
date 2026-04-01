package routes

import (
	"fmt"
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
	"github.com/go-risk-it/go-risk-it/internal/lobby/ctx"
	lobbyWs "github.com/go-risk-it/go-risk-it/internal/lobby/ws"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	"github.com/go-risk-it/go-risk-it/internal/web/ws"
	"go.opentelemetry.io/otel/attribute"
)

func ProvideRoutes(
	creationCtrl *CreationController,
	managementCtrl *ManagementController,
	startCtrl *StartController,
	lobbyConnectionManager lobbyWs.Manager,
	upgrader ws.Upgrader,
) []*route.Route {
	return []*route.Route{
		route.CreateHandler("POST /api/v1/lobbies", creationCtrl.CreateLobby),
		route.QueryHandler("GET /api/v1/lobbies/summary", managementCtrl.GetUserLobbies),
		route.DomainCommand(
			"POST /api/v1/lobbies/{id}/join",
			ctx.WithLobbyID,
			managementCtrl.JoinLobby,
		),
		route.DomainVoid(
			"POST /api/v1/lobbies/{id}/start",
			ctx.WithLobbyID,
			startCtrl.StartGame,
		),
		route.DomainWS(
			"GET /api/v1/lobbies/{id}/ws",
			func(r *http.Request) (ctx.LobbyContext, error) {
				return route.BuildDomainContext(r, ctx.WithLobbyID)
			},
			connectLobbyWS(lobbyConnectionManager, upgrader),
		),
	}
}

func connectLobbyWS(
	lobbyConnectionManager lobbyWs.Manager,
	upgrader ws.Upgrader,
) func(http.ResponseWriter, *http.Request, ctx.LobbyContext) error {
	return func(
		writer http.ResponseWriter,
		request *http.Request,
		lobbyCtx ctx.LobbyContext,
	) error {
		conn, err := upgrader.Upgrade(writer, request, nil)
		if err != nil {
			return fmt.Errorf("unable to upgrade websocket connection: %w", err)
		}

		lobbyConnectionManager.ConnectPlayer(lobbyCtx, conn)
		observe.Info(
			lobbyCtx,
			"Lobby WS upgraded",
			attribute.String("remoteAddress", conn.RemoteAddr().String()),
		)

		return nil
	}
}
