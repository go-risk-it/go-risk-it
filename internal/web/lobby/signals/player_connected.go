package signals

import (
	"context"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/logic/lobby/signals"
)

func HandlePlayerConnected(
	params HandlerParams[signals.PlayerConnectedSignal],
) {
	params.Signal.AddListener(func(context context.Context, data signals.PlayerConnectedData) {
		lobbyContext, ok := context.(ctx.LobbyContext)
		if !ok {
			params.Log.Errorw("context is not a lobby context", "context", context)

			return
		}

		lobbyContext.Log().Infow("handling player connected. fetching state and publishing")

		fetchStateAndPublish(
			lobbyContext,
			params.LobbyStateFetcher.FetchState,
			params.ConnectionManager.WriteMessage,
		)
	})
}
