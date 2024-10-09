package handlers

import (
	"encoding/json"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
)

func fetchStateAndBroadcast(
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
