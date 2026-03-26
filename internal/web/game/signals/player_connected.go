package signals

import (
	"context"
	"log/slog"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/signals"
	"github.com/go-risk-it/go-risk-it/internal/safego"
)

func HandlePlayerConnected(
	params HandlerParams[signals.PlayerConnectedSignal],
) {
	params.Signal.AddListener(func(context context.Context, data signals.PlayerConnectedData) {
		gameContext, ok := context.(ctx.GameContext)
		if !ok {
			slog.ErrorContext(context, "context is not game context")

			return
		}

		slog.InfoContext(
			gameContext, "handling player connected",
		)

		safego.Go(gameContext, func() {
			fetchAllStatesAndPublish(
				gameContext, params,
				params.ConnectionManager.WriteMessage,
			)
		})

		slog.InfoContext(
			gameContext, "fetching move logs and publishing",
		)

		safego.Go(gameContext, func() {
			fetchStateAndPublish(
				gameContext,
				params.MoveLogFetcher.FetchState,
				params.ConnectionManager.WriteMessage,
			)
		})
	})
}
