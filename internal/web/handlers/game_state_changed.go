package handlers

import (
	"context"
	"encoding/json"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/state"
	"github.com/go-risk-it/go-risk-it/internal/logic/signals"
	"github.com/go-risk-it/go-risk-it/internal/web/fetchers/phase"
	"github.com/go-risk-it/go-risk-it/internal/web/ws/connection"
	"go.uber.org/zap"
)

func HandleGameStateChanged(
	log *zap.SugaredLogger,
	gameService state.Service,
	deployPhaseFetcher phase.DeployPhaseFetcher,
	attackPhaseFetcher phase.AttackPhaseFetcher,
	conquerPhaseFetcher phase.ConquerPhaseFetcher,
	connectionManager connection.Manager,
	signal signals.GameStateChangedSignal,
) {
	signal.AddListener(func(context context.Context, data signals.GameStateChangedData) {
		gameContext := ctx.WithGameID(ctx.WithLog(context, log), data.GameID)

		gameState, err := gameService.GetGameState(gameContext)
		if err != nil {
			gameContext.Log().Errorf("failed to get game state: %v", err)

			return
		}

		var fetcher phase.Fetcher

		switch gameState.Phase {
		case sqlc.PhaseTypeDEPLOY:
			fetcher = deployPhaseFetcher
		case sqlc.PhaseTypeATTACK:
			fetcher = attackPhaseFetcher
		case sqlc.PhaseTypeCONQUER:
			fetcher = conquerPhaseFetcher
		default:
			gameContext.Log().Errorf("unknown phase type: %v", gameState.Phase)

			return
		}

		fetchStateAndBroadcast(
			gameContext,
			func(ctx ctx.GameContext, channel chan json.RawMessage) {
				fetcher.FetchState(ctx, gameState, channel)
			},
			connectionManager.Broadcast)
	})
}
