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
	"github.com/lesismal/nbio/nbhttp/websocket"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type PlayerConnectedHandlerParams struct {
	fx.In

	Fetchers            []fetcher.Fetcher `group:"fetchers"`
	Log                 *zap.SugaredLogger
	Signal              signals.PlayerConnectedSignal
	GameService         state.Service
	DeployPhaseFetcher  phase.DeployPhaseFetcher
	AttackPhaseFetcher  phase.AttackPhaseFetcher
	ConquerPhaseFetcher phase.ConquerPhaseFetcher
}

func HandlePlayerConnected(
	params PlayerConnectedHandlerParams,
) {
	params.Signal.AddListener(func(cont context.Context, data signals.PlayerConnectedData) {
		childCtx, cancel := context.WithTimeout(cont, 10*time.Second)
		defer cancel()

		gameContext := ctx.WithGameID(ctx.WithLog(childCtx, params.Log), data.GameID)

		gameContext.Log().Infow("handling player connected",
			"remoteAddress", data.Connection.RemoteAddr().String())

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
		default:
			gameContext.Log().Errorf("unknown phase type: %v", gameState.Phase)

			return
		}

		wait(params, stateChannel, gameContext, data, childCtx)
	})
}

func wait(
	params PlayerConnectedHandlerParams,
	stateChannel chan json.RawMessage,
	gameContext ctx.GameContext,
	data signals.PlayerConnectedData,
	ctx context.Context,
) {
	for range len(params.Fetchers) + 1 {
		select {
		case state := <-stateChannel:
			gameContext.Log().Infow("got state, writing message")

			err := data.Connection.WriteMessage(websocket.TextMessage, state)
			if err != nil {
				gameContext.Log().Errorf("unable to write response: %v", err)
			}
		case <-ctx.Done():
			gameContext.Log().Errorf("unable to get all states: %v", ctx.Err())

			return
		}
	}
}
