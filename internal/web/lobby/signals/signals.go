package signals

import (
	"encoding/json"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/web/lobby/fetcher"
	"github.com/go-risk-it/go-risk-it/internal/web/lobby/ws"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Options(
	fx.Invoke(
		HandlePlayerConnected,
		HandleLobbyStateChanged,
	),
)

type HandlerParams[T any] struct {
	fx.In

	LobbyStateFetcher fetcher.LobbyStateFetcher
	Log               *zap.SugaredLogger
	Signal            T
	ConnectionManager ws.Manager
}

func fetchStateAndPublish(
	ctx ctx.LobbyContext,
	fetcher func(ctx.LobbyContext, chan json.RawMessage),
	publisher func(ctx.LobbyContext, json.RawMessage),
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
