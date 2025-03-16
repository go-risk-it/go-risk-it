package fetcher

import (
	"encoding/json"

	"github.com/go-risk-it/go-risk-it/internal/api/game/messaging"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/state"
	"github.com/go-risk-it/go-risk-it/internal/web/game/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/ws/message"
	"go.uber.org/fx"
)

type GameFetcher interface {
	Fetcher
}

type GameFetcherImpl struct {
	gameService     state.Service
	phaseController controller.PhaseController
}

var _ GameFetcher = (*GameFetcherImpl)(nil)

type GameFetcherResult struct {
	fx.Out

	GameFetcher GameFetcher
	Fetcher     Fetcher `group:"public_fetchers"`
}

func NewGameFetcher(
	gameService state.Service,
	phaseController controller.PhaseController,
) GameFetcherResult {
	res := &GameFetcherImpl{
		gameService:     gameService,
		phaseController: phaseController,
	}

	return GameFetcherResult{
		GameFetcher: res,
		Fetcher:     res,
	}
}

func (g *GameFetcherImpl) FetchState(ctx ctx.GameContext, stateChannel chan json.RawMessage) {
	gameState, err := g.gameService.GetGameState(ctx)
	if err != nil {
		ctx.Log().Errorf("failed to get game state: %v", err)

		return
	}

	switch gameState.Phase {
	case sqlc.GamePhaseTypeDEPLOY:
		FetchState(
			ctx,
			message.GameState,
			getGameFetcherFunc(gameState, g.phaseController.GetDeployPhaseState),
			stateChannel)
	case sqlc.GamePhaseTypeATTACK:
		FetchState(
			ctx,
			message.GameState,
			getGameFetcherFunc(gameState, g.phaseController.GetAttackPhaseState),
			stateChannel)
	case sqlc.GamePhaseTypeCONQUER:
		FetchState(
			ctx,
			message.GameState,
			getGameFetcherFunc(gameState, g.phaseController.GetConquerPhaseState),
			stateChannel)
	case sqlc.GamePhaseTypeREINFORCE:
		FetchState(
			ctx,
			message.GameState,
			getGameFetcherFunc(gameState, g.phaseController.GetReinforcePhaseState),
			stateChannel)
	case sqlc.GamePhaseTypeCARDS:
		FetchState(
			ctx,
			message.GameState,
			getGameFetcherFunc(gameState, g.phaseController.GetCardsPhaseState),
			stateChannel)
	default:
		ctx.Log().Errorf("unknown phase type: %v", gameState.Phase)

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
