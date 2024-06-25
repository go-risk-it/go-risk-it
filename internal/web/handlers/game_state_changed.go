package handlers

import (
	"context"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/logic/signals"
	"github.com/go-risk-it/go-risk-it/internal/web/fetchers"
	"github.com/go-risk-it/go-risk-it/internal/web/ws/connection"
	"go.uber.org/zap"
)

func HandleGameStateChanged(
	log *zap.SugaredLogger,
	gameStateFetcher fetchers.GameFetcher,
	connectionManager connection.Manager,
	signal signals.GameStateChangedSignal,
) {
	signal.AddListener(func(context context.Context, data signals.GameStateChangedData) {
		ctx := ctx.WithGameID(ctx.WithLog(context, log), data.GameID)

		ctx.Log().Infow("handling game state changed")
		fetchStateAndBroadcast(
			ctx,
			gameStateFetcher.FetchState,
			connectionManager.Broadcast)
	})
}
