package signals

import (
	"context"
	"log/slog"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/logic/lobby/signals"
)

func HandlePlayerConnected(
	params HandlerParams[signals.PlayerConnectedSignal],
) {
	params.Signal.AddListener(func(context context.Context, data signals.PlayerConnectedData) {
		lobbyContext, ok := context.(ctx.LobbyContext)
		if !ok {
			slog.ErrorContext(context, "context is not a lobby context")

			return
		}

		slog.InfoContext(
			lobbyContext, "handling player connected",
		)

		fetchStateAndPublish(
			lobbyContext,
			params.LobbyStateFetcher.FetchState,
			params.ConnectionManager.WriteMessage,
		)
	})
}
