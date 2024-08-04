package handlers

import (
	"context"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/logic/signals"
	"github.com/go-risk-it/go-risk-it/internal/web/fetchers/fetcher"
	"github.com/go-risk-it/go-risk-it/internal/web/ws/connection"
	"go.uber.org/zap"
)

func HandlePlayerStateChanged(
	log *zap.SugaredLogger,
	playerStateFetcher fetcher.PlayerFetcher,
	connectionManager connection.Manager,
	signal signals.PlayerStateChangedSignal,
) {
	signal.AddListener(func(context context.Context, data signals.PlayerStateChangedData) {
		ctx := ctx.WithGameID(ctx.WithLog(context, log), data.GameID)

		ctx.Log().Infow("handling player state changed")
		fetchStateAndBroadcast(
			ctx,
			playerStateFetcher.FetchState,
			connectionManager.Broadcast)
	})
}
