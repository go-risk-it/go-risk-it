package handlers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/tomfran/go-risk-it/internal/web/ws/connection"
	"go.uber.org/zap"
)

func FetchStateAndBroadcast(
	ctx context.Context,
	gameID int64,
	log *zap.SugaredLogger,
	fetcher func(context.Context, int64, chan json.RawMessage),
	connectionManager connection.Manager,
) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	channel := make(chan json.RawMessage)
	fetcher(ctx, gameID, channel)

	select {
	case msg := <-channel:
		connectionManager.Broadcast(gameID, msg)
	case <-ctx.Done():
		log.Errorf("timeout while fetching state: %v", ctx.Err())
	}
}
