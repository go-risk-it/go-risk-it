package handlers

import (
	"encoding/json"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/state"
	"github.com/go-risk-it/go-risk-it/internal/web/fetchers/fetcher"
	"github.com/go-risk-it/go-risk-it/internal/web/fetchers/phase"
	"github.com/go-risk-it/go-risk-it/internal/web/ws/connection"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Options(
	fx.Invoke(
		HandleGameStateChanged,
		HandleMovePerformed,
		HandlePlayerConnected,
	),
)

type HandlerParams[T any] struct {
	fx.In

	PublicFetchers        []fetcher.Fetcher `group:"public_fetchers"`
	PrivateFetchers       []fetcher.Fetcher `group:"private_fetchers"`
	Log                   *zap.SugaredLogger
	Signal                T
	GameService           state.Service
	DeployPhaseFetcher    phase.DeployPhaseFetcher
	AttackPhaseFetcher    phase.AttackPhaseFetcher
	ConquerPhaseFetcher   phase.ConquerPhaseFetcher
	ReinforcePhaseFetcher phase.ReinforcePhaseFetcher
	CardsPhaseFetcher     phase.CardsPhaseFetcher
	MoveLogFetcher        fetcher.MoveLogFetcher
	ConnectionManager     connection.Manager
}

func fetchAllStatesAndPublish[T any](
	context ctx.GameContext,
	params HandlerParams[T],
	publisher func(ctx.GameContext, json.RawMessage),
) {
	go fetchGameState(
		context,
		params,
		publisher)

	for _, fetcher := range params.PublicFetchers {
		go fetchStateAndPublish(context, fetcher.FetchState, publisher)
	}

	for _, fetcher := range params.PrivateFetchers {
		go fetchStateAndPublish(
			context,
			fetcher.FetchState,
			params.ConnectionManager.WriteMessage,
		)
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
