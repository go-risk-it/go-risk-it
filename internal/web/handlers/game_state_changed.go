package handlers

import (
	"context"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/logic/signals"
)

func HandleGameStateChanged(
	params HandlerParams[signals.GameStateChangedSignal],
) {
	params.Signal.AddListener(func(context context.Context, data signals.GameStateChangedData) {
		gameContext, ok := context.(ctx.GameContext)
		if !ok {
			params.Log.Errorw("context is not game context", "context", context)

			return
		}

		gameContext.Log().Infow("handling game state changed")

		fetchAllStatesAndPublish(gameContext, params, params.ConnectionManager.Broadcast)
	})
}
