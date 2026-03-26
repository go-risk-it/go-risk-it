package signals

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/safego"
	"github.com/go-risk-it/go-risk-it/internal/web/lobby/fetcher"
	"github.com/go-risk-it/go-risk-it/internal/web/lobby/ws"
	"go.uber.org/fx"
)

const fetchTimeout = 10 * time.Second

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
	lobbyCtx ctx.LobbyContext,
	fetcher func(ctx.LobbyContext, chan json.RawMessage),
	publisher func(ctx.LobbyContext, json.RawMessage),
) {
	detached, cancel := ctx.DetachLobbyContextWithTimeout(lobbyCtx, fetchTimeout)
	defer cancel()

	channel := make(chan json.RawMessage, 1)
	safego.Go(detached, func() { fetcher(detached, channel) })

	select {
	case msg := <-channel:
		publisher(detached, msg)
	case <-detached.Done():
		slog.ErrorContext(detached, "timeout while fetching state", "timeout", fetchTimeout)
	}
}
