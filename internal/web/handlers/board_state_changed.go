package handlers

import (
	"context"

	"github.com/tomfran/go-risk-it/internal/signals"
	"github.com/tomfran/go-risk-it/internal/web/fetchers"
	"github.com/tomfran/go-risk-it/internal/web/ws/connection"
	"go.uber.org/zap"
)

func HandleBoardStateChanged(
	log *zap.SugaredLogger,
	boardStateFetcher fetchers.BoardFetcher,
	connectionManager connection.Manager,
	signal signals.BoardStateChangedSignal,
) {
	signal.AddListener(func(ctx context.Context, data signals.BoardStateChangedData) {
		log.Infow("handling board state changed", "data", data)

		fetchStateAndBroadcast(
			ctx,
			data.GameID,
			log,
			boardStateFetcher.FetchState,
			connectionManager.Broadcast)
	})
}
