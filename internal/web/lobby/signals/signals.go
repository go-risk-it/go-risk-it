package signals

import (
	"encoding/json"
	"log/slog"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/web/lobby/fetcher"
	"github.com/go-risk-it/go-risk-it/internal/web/lobby/ws"
	"go.uber.org/fx"
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
		slog.ErrorContext(ctx, "timeout while fetching state", "error", ctx.Err())
	}
}
