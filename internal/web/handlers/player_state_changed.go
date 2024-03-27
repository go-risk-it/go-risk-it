package handlers

import (
	"context"

	"github.com/tomfran/go-risk-it/internal/signals"
	"github.com/tomfran/go-risk-it/internal/web/fetchers"
	"github.com/tomfran/go-risk-it/internal/web/ws/connection"
	"go.uber.org/zap"
)

func HandlePlayerStateChanged(
	log *zap.SugaredLogger,
	playerStateFetcher fetchers.PlayerFetcher,
	connectionManager connection.Manager,
	signal signals.PlayerStateChangedSignal,
) {
	signal.AddListener(func(ctx context.Context, data signals.PlayerStateChangedData) {
		log.Infow("handling player state changed", "data", data)

		fetchStateAndBroadcast(
			ctx,
			data.GameID,
			log,
			playerStateFetcher.FetchState,
			connectionManager.Broadcast)
	})
}
