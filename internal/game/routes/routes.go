package routes

import (
	"fmt"
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/player"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/state"
	gameWs "github.com/go-risk-it/go-risk-it/internal/game/ws"
	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	"github.com/go-risk-it/go-risk-it/internal/web/ws"
	"go.opentelemetry.io/otel/attribute"
)

func ProvideRoutes(
	gameCtrl *GameController,
	advCtrl *AdvancementController,
	moveCtrl *MoveController,
	gameConnectionManager gameWs.Gateway,
	upgrader ws.Upgrader,
	gameStateService state.Service,
	playerService player.Service,
) []*route.Route {
	return []*route.Route{
		route.CreateHandler("POST /api/v1/games", gameCtrl.CreateGame),
		route.QueryHandler("GET /api/v1/games/summary", gameCtrl.GetUserGames),
		route.DomainCommand(
			"POST /api/v1/games/{id}/advancements",
			ctx.WithGameID,
			advCtrl.Advance,
		),
		route.DomainCommand(
			"POST /api/v1/games/{id}/moves/deployments",
			ctx.WithGameID,
			moveCtrl.PerformDeployMove,
		),
		route.DomainCommand(
			"POST /api/v1/games/{id}/moves/attacks",
			ctx.WithGameID,
			moveCtrl.PerformAttackMove,
		),
		route.DomainCommand(
			"POST /api/v1/games/{id}/moves/conquers",
			ctx.WithGameID,
			moveCtrl.PerformConquerMove,
		),
		route.DomainCommand(
			"POST /api/v1/games/{id}/moves/reinforcements",
			ctx.WithGameID,
			moveCtrl.PerformReinforceMove,
		),
		route.DomainCommand(
			"POST /api/v1/games/{id}/moves/cards",
			ctx.WithGameID,
			moveCtrl.PerformCardsMove,
		),
		route.DomainWS(
			"GET /api/v1/games/{id}/ws",
			func(r *http.Request) (ctx.GameContext, error) {
				return route.BuildDomainContext(r, ctx.WithGameID)
			},
			connectGameWS(gameConnectionManager, upgrader, gameStateService, playerService),
		),
	}
}

func connectGameWS(
	gameConnectionManager gameWs.Gateway,
	upgrader ws.Upgrader,
	gameStateService state.Service,
	playerService player.Service,
) func(http.ResponseWriter, *http.Request, ctx.GameContext) error {
	return func(writer http.ResponseWriter, request *http.Request, gameCtx ctx.GameContext) error {
		if err := ValidateGameWSConnection(gameCtx, gameStateService, playerService); err != nil {
			return err
		}

		conn, err := upgrader.Upgrade(writer, request, nil)
		if err != nil {
			return fmt.Errorf("unable to upgrade websocket connection: %w", err)
		}

		gameConnectionManager.ConnectPlayer(gameCtx, conn)
		observe.Info(
			request.Context(),
			"Game WS upgraded",
			attribute.String("remoteAddress", conn.RemoteAddr().String()),
		)

		return nil
	}
}
