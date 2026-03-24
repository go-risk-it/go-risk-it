package signals

import (
	"context"
	"log/slog"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/logic/lobby/signals"
)

func HandleLobbyStateChanged(
	params HandlerParams[signals.LobbyStateChangedSignal],
) {
	params.Signal.AddListener(func(context context.Context, _ signals.LobbyStateChangedData) {
		lobbyContext, ok := context.(ctx.LobbyContext)
		if !ok {
			slog.ErrorContext(context, "context is not a lobby context")

			return
		}

		slog.InfoContext(
			lobbyContext, "handling lobby state changed",
		)

		fetchStateAndPublish(
			lobbyContext,
			params.LobbyStateFetcher.FetchState,
			params.ConnectionManager.Broadcast,
		)
	})
}
