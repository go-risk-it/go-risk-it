package fetcher

import (
	"encoding/json"
	"log/slog"

	"github.com/go-risk-it/go-risk-it/internal/api/game/messaging"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/state"
	sharedfetcher "github.com/go-risk-it/go-risk-it/internal/web/fetcher"
	"github.com/go-risk-it/go-risk-it/internal/web/game/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/ws/message"
	"go.uber.org/fx"
)

type gameFetcher struct {
	gameService     state.Service
	phaseController *controller.PhaseController
}

type GameFetcherResult struct {
	fx.Out

	Fetcher Fetcher `group:"public_fetchers"`
}

func NewGameFetcher(
	gameService state.Service,
	phaseController *controller.PhaseController,
) GameFetcherResult {
	return GameFetcherResult{
		Fetcher: &gameFetcher{
			gameService:     gameService,
			phaseController: phaseController,
		},
	}
}

func (g *gameFetcher) FetchState(ctx ctx.GameContext, stateChannel chan json.RawMessage) {
	gameState, err := g.gameService.GetGameState(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get game state", "error", err)

		return
	}

	switch gameState.Phase {
	case sqlc.GamePhaseTypeDEPLOY:
		sharedfetcher.FetchState(
			ctx,
			message.GameState,
			getGameFetcherFunc(gameState, g.phaseController.GetDeployPhaseState),
			stateChannel)
	case sqlc.GamePhaseTypeATTACK:
		sharedfetcher.FetchState(
			ctx,
			message.GameState,
			getGameFetcherFunc(gameState, g.phaseController.GetAttackPhaseState),
			stateChannel)
	case sqlc.GamePhaseTypeCONQUER:
		sharedfetcher.FetchState(
			ctx,
			message.GameState,
			getGameFetcherFunc(gameState, g.phaseController.GetConquerPhaseState),
			stateChannel)
	case sqlc.GamePhaseTypeREINFORCE:
		sharedfetcher.FetchState(
			ctx,
			message.GameState,
			getGameFetcherFunc(gameState, g.phaseController.GetReinforcePhaseState),
			stateChannel)
	case sqlc.GamePhaseTypeCARDS:
		sharedfetcher.FetchState(
			ctx,
			message.GameState,
			getGameFetcherFunc(gameState, g.phaseController.GetCardsPhaseState),
			stateChannel)
	default:
		slog.ErrorContext(ctx, "unknown phase type", "phase", gameState.Phase)

		return
	}
}

func getGameFetcherFunc[T messaging.PhaseState](
	game *state.Game,
	fetcherFunc func(ctx.GameContext, *state.Game) (messaging.GameState[T], error),
) func(context ctx.GameContext) (messaging.GameState[T], error) {
	return func(cont ctx.GameContext) (messaging.GameState[T], error) {
		return fetcherFunc(cont, game)
	}
}
