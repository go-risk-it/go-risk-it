package handlers

import (
	"context"
	"encoding/json"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/state"
	"github.com/go-risk-it/go-risk-it/internal/web/fetchers/fetcher"
	"github.com/go-risk-it/go-risk-it/internal/web/fetchers/phase"
	"github.com/go-risk-it/go-risk-it/internal/web/ws/connection"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
)

var Module = fx.Options(
	fx.Invoke(
		HandleGameStateChanged,
		HandlePlayerConnected,
	),
)

type HandlerParams[T any] struct {
	fx.In

	Fetchers              []fetcher.Fetcher `group:"fetchers"`
	Log                   *zap.SugaredLogger
	Signal                T
	GameService           state.Service
	DeployPhaseFetcher    phase.DeployPhaseFetcher
	AttackPhaseFetcher    phase.AttackPhaseFetcher
	ConquerPhaseFetcher   phase.ConquerPhaseFetcher
	ReinforcePhaseFetcher phase.ReinforcePhaseFetcher
	CardsPhaseFetcher     phase.CardsPhaseFetcher
	ConnectionManager     connection.Manager
}

func fetchAllStatesAndPublish[T any](
	context context.Context,
	params HandlerParams[T],
	publisher func(ctx.GameContext, json.RawMessage),
) {
	gameContext, ok := context.(ctx.GameContext)
	if !ok {
		params.Log.Errorw("context is not game context", "context", context)

		return
	}

	gameContext.Log().Infow("handling player connected")

	go fetchGameState(
		gameContext,
		params,
		publisher)

	for _, fetcher := range params.Fetchers {
		go fetchStateAndPublish(gameContext, fetcher.FetchState, publisher)
	}
}

func fetchGameState[T any](
	gameContext ctx.GameContext,
	params HandlerParams[T],
	publisher func(ctx.GameContext, json.RawMessage),
) {
	gameState, err := params.GameService.GetGameState(gameContext)
	if err != nil {
		gameContext.Log().Errorf("failed to get game state: %v", err)

		return
	}

	var fetcher phase.Fetcher

	switch gameState.Phase {
	case sqlc.PhaseTypeDEPLOY:
		fetcher = params.DeployPhaseFetcher
	case sqlc.PhaseTypeATTACK:
		fetcher = params.AttackPhaseFetcher
	case sqlc.PhaseTypeCONQUER:
		fetcher = params.ConquerPhaseFetcher
	case sqlc.PhaseTypeREINFORCE:
		fetcher = params.ReinforcePhaseFetcher
	case sqlc.PhaseTypeCARDS:
		fetcher = params.CardsPhaseFetcher
	default:
		gameContext.Log().Errorf("unknown phase type: %v", gameState.Phase)

		return
	}

	fetchStateAndPublish(
		gameContext,
		func(ctx ctx.GameContext, channel chan json.RawMessage) {
			fetcher.FetchState(ctx, gameState, channel)
		},
		publisher)
}

func fetchStateAndPublish(
	ctx ctx.GameContext,
	fetcher func(ctx.GameContext, chan json.RawMessage),
	publisher func(ctx.GameContext, json.RawMessage),
) {
	channel := make(chan json.RawMessage)
	go fetcher(ctx, channel)

	select {
	case msg := <-channel:
		publisher(ctx, msg)
	case <-ctx.Done():
		ctx.Log().Errorf("timeout while fetching state: %v", ctx.Err())
	}
}
