package signals

import (
	"encoding/json"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/web/game/fetcher"
	"github.com/go-risk-it/go-risk-it/internal/web/game/ws"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

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
	Log               *zap.SugaredLogger
	Signal            T
	MoveLogFetcher    fetcher.MoveLogFetcher
	ConnectionManager ws.Manager
}

func fetchAllStatesAndPublish[T any](
	context ctx.GameContext,
	params HandlerParams[T],
	publisher func(ctx.GameContext, json.RawMessage),
) {
	for _, fetcher := range params.PublicFetchers {
		go fetchStateAndPublish(context, fetcher.FetchState, publisher)
	}

	for _, fetcher := range params.PrivateFetchers {
		go fetchStateAndPublish(
			context,
			fetcher.FetchState,
			params.ConnectionManager.WriteMessage,
		)
	}
}

func fetchStateAndPublish(
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
