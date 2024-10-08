package handlers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/state"
	"github.com/go-risk-it/go-risk-it/internal/logic/signals"
	"github.com/go-risk-it/go-risk-it/internal/web/fetchers/fetcher"
	"github.com/go-risk-it/go-risk-it/internal/web/fetchers/phase"
	"github.com/go-risk-it/go-risk-it/internal/web/ws/connection"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type PlayerConnectedHandlerParams struct {
	fx.In

	Fetchers              []fetcher.Fetcher `group:"fetchers"`
	Log                   *zap.SugaredLogger
	Signal                signals.PlayerConnectedSignal
	GameService           state.Service
	DeployPhaseFetcher    phase.DeployPhaseFetcher
	AttackPhaseFetcher    phase.AttackPhaseFetcher
	ConquerPhaseFetcher   phase.ConquerPhaseFetcher
	ReinforcePhaseFetcher phase.ReinforcePhaseFetcher
	CardsPhaseFetcher     phase.CardsPhaseFetcher
	ConnectionManager     connection.Manager
}

func HandlePlayerConnected(
	params PlayerConnectedHandlerParams,
) {
	params.Signal.AddListener(func(cont context.Context, data signals.PlayerConnectedData) {
		gameContext, ok := cont.(ctx.GameContext)
		if !ok {
			params.Log.Errorw("context is not game context", "context", cont)

			return
		}

		gameContext.Log().Infow("handling player connected")

		stateChannel := make(chan json.RawMessage, len(params.Fetchers))

		gameContext.Log().Infow("fetching states", "count", len(params.Fetchers))

		for _, fetcher := range params.Fetchers {
			go fetcher.FetchState(gameContext, stateChannel)
		}

		gameState, err := params.GameService.GetGameState(gameContext)
		if err != nil {
			gameContext.Log().Errorf("failed to get game state: %v", err)

			return
		}

		switch gameState.Phase {
		case sqlc.PhaseTypeDEPLOY:
			go params.DeployPhaseFetcher.FetchState(gameContext, gameState, stateChannel)
		case sqlc.PhaseTypeATTACK:
			go params.AttackPhaseFetcher.FetchState(gameContext, gameState, stateChannel)
		case sqlc.PhaseTypeCONQUER:
			go params.ConquerPhaseFetcher.FetchState(gameContext, gameState, stateChannel)
		case sqlc.PhaseTypeREINFORCE:
			go params.ReinforcePhaseFetcher.FetchState(gameContext, gameState, stateChannel)
		case sqlc.PhaseTypeCARDS:
			go params.CardsPhaseFetcher.FetchState(gameContext, gameState, stateChannel)
		default:
			gameContext.Log().Errorf("unknown phase type: %v", gameState.Phase)

			return
		}

		wait(gameContext, params.Fetchers, params.ConnectionManager, stateChannel)
	})
}

func wait(
	gameContext ctx.GameContext,
	fetchers []fetcher.Fetcher,
	connectionManager connection.Manager,
	stateChannel chan json.RawMessage,
) {
	childCtx, cancel := context.WithTimeout(gameContext, 10*time.Second)
	defer cancel()

	for range len(fetchers) + 1 {
		select {
		case state := <-stateChannel:
			gameContext.Log().Infow("got state, writing message")

			connectionManager.WriteMessage(gameContext, state)
		case <-childCtx.Done():
			gameContext.Log().Errorf("unable to get all states: %v", childCtx.Err())

			return
		}
	}
}
