package signals

import (
	"context"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/signals"
)

func HandlePlayerConnected(
	params HandlerParams[signals.PlayerConnectedSignal],
) {
	params.Signal.AddListener(func(context context.Context, data signals.PlayerConnectedData) {
		gameContext, ok := context.(ctx.GameContext)
		if !ok {
			params.Log.Errorw("context is not game context", "context", context)

			return
		}

		gameContext.Log().Infow("handling player connected. fetching all states and publishing")

		fetchAllStatesAndPublish(gameContext, params, params.ConnectionManager.WriteMessage)

		gameContext.Log().Infow("fetching move logs and publishing")

		fetchStateAndPublish(
			gameContext,
			params.MoveLogFetcher.FetchState,
			params.ConnectionManager.WriteMessage,
		)
	})
}
