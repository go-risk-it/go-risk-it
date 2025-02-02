package signals

import (
	"context"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/logic/lobby/signals"
)

func HandleLobbyStateChanged(
	params HandlerParams[signals.LobbyStateChangedSignal],
) {
	params.Signal.AddListener(func(context context.Context, _ signals.LobbyStateChangedData) {
		lobbyContext, ok := context.(ctx.LobbyContext)
		if !ok {
			params.Log.Errorw("context is not a lobby context", "context", context)

			return
		}

		lobbyContext.Log().Infow("handling lobby state changed. fetching state and publishing")

		fetchStateAndPublish(
			lobbyContext,
			params.LobbyStateFetcher.FetchState,
			params.ConnectionManager.Broadcast,
		)
	})
}
