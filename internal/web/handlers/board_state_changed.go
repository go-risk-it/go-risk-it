package handlers

import (
	"context"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/logic/signals"
	"github.com/go-risk-it/go-risk-it/internal/web/fetchers/fetcher"
	"github.com/go-risk-it/go-risk-it/internal/web/ws/connection"
	"go.uber.org/zap"
)

func HandleBoardStateChanged(
	log *zap.SugaredLogger,
	boardStateFetcher fetcher.BoardFetcher,
	connectionManager connection.Manager,
	signal signals.BoardStateChangedSignal,
) {
	signal.AddListener(func(context context.Context, data signals.BoardStateChangedData) {
		ctx := ctx.WithGameID(ctx.WithLog(context, log), data.GameID)

		ctx.Log().Infow("handling board state changed")
		fetchStateAndBroadcast(
			ctx,
			boardStateFetcher.FetchState,
			connectionManager.Broadcast)
	})
}
