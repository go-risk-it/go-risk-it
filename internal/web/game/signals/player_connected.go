package signals

import (
	"context"
	"log/slog"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/signals"
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

		slog.InfoContext( //nolint:contextcheck
			gameContext, "handling player connected",
		)

		go fetchAllStatesAndPublish(
			gameContext, params,
			params.ConnectionManager.WriteMessage,
		)

		slog.InfoContext( //nolint:contextcheck
			gameContext, "fetching move logs and publishing",
		)

		//nolint:contextcheck // deliberate context detach
		go fetchStateAndPublish(
			gameContext,
			params.MoveLogFetcher.FetchState,
			params.ConnectionManager.WriteMessage,
		)
	})
}
