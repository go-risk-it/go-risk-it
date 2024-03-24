package handlers

import (
	"context"

	"github.com/tomfran/go-risk-it/internal/signals"
	"github.com/tomfran/go-risk-it/internal/web/fetchers"
	"github.com/tomfran/go-risk-it/internal/web/ws/connection"
	"go.uber.org/zap"
)

func HandleGameStateChanged(
	log *zap.SugaredLogger,
	gameStateFetcher fetchers.GameFetcher,
	connectionManager connection.Manager,
	signal signals.GameStateChangedSignal,
) {
	signal.AddListener(func(ctx context.Context, data signals.GameStateChangedData) {
		log.Infow("handling game state changed", "gameID", data.GameID)

		FetchStateAndBroadcast(
			ctx,
			data.GameID,
			log,
			gameStateFetcher.FetchState,
			connectionManager)
	})
}
