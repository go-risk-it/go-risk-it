package signals

import (
	"context"
	"log/slog"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/signals"
)

func HandleGameStateChanged(
	params HandlerParams[signals.GameStateChangedSignal],
) {
	params.Signal.AddListener(func(context context.Context, data signals.GameStateChangedData) {
		gameContext, ok := context.(ctx.GameContext)
		if !ok {
			slog.ErrorContext(context, "context is not game context")

			return
		}

		slog.InfoContext(
			gameContext,
			"handling game state changed",
			"fromPhase", data.FromPhase,
			"toPhase", data.ToPhase,
		)

		fetchAllStatesAndPublish(
			gameContext, params,
			params.ConnectionManager.Broadcast,
		)
	})
}
