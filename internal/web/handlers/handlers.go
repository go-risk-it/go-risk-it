package handlers

import (
	"context"
	"encoding/json"
	"time"

	"go.uber.org/zap"
)

func fetchStateAndBroadcast(
	ctx context.Context,
	gameID int64,
	log *zap.SugaredLogger,
	fetcher func(context.Context, int64, chan json.RawMessage),
	broadcast func(int64, json.RawMessage),
) {
	childCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	channel := make(chan json.RawMessage)
	go fetcher(childCtx, gameID, channel)

	select {
	case msg := <-channel:
		broadcast(gameID, msg)
	case <-childCtx.Done():
		log.Errorf("timeout while fetching state: %v", ctx.Err())
	}
}
