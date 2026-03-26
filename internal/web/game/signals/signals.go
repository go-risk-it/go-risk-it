package signals

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/safego"
	"github.com/go-risk-it/go-risk-it/internal/web/game/fetcher"
	"github.com/go-risk-it/go-risk-it/internal/web/game/ws"
	"go.uber.org/fx"
)

const fetchTimeout = 10 * time.Second

var Module = fx.Options(
	fx.Invoke(
		HandleGameStateChanged,
		HandleMovePerformed,
		HandlePlayerConnected,
	),
)

type HandlerParams[T any] struct {
	fx.In

	PublicFetchers    []fetcher.Fetcher `group:"public_fetchers"`
	PrivateFetchers   []fetcher.Fetcher `group:"private_fetchers"`
	Signal            T
	MoveLogFetcher    fetcher.MoveLogFetcher
	ConnectionManager ws.Manager
}

func fetchAllStatesAndPublish[T any](
	context ctx.GameContext,
	params HandlerParams[T],
	publisher func(ctx.GameContext, json.RawMessage),
) {
	for _, f := range params.PublicFetchers {
		safego.Go(context, func() {
			fetchStateAndPublish(context, f.FetchState, publisher)
		})
	}

	for _, f := range params.PrivateFetchers {
		safego.Go(context, func() {
			fetchStateAndPublish(
				context,
				f.FetchState,
				params.ConnectionManager.WriteMessage,
			)
		})
	}
}

func fetchStateAndPublish(
	gameCtx ctx.GameContext,
	fetcher func(ctx.GameContext, chan json.RawMessage),
	publisher func(ctx.GameContext, json.RawMessage),
) {
	detached, cancel := ctx.DetachGameContextWithTimeout(gameCtx, fetchTimeout)
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
