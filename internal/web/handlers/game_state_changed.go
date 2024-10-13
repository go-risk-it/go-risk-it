package handlers

import (
	"context"
	"github.com/go-risk-it/go-risk-it/internal/logic/signals"
)

func HandleGameStateChanged(
	params HandlerParams[signals.GameStateChangedSignal],
) {
	params.Signal.AddListener(func(context context.Context, data signals.GameStateChangedData) {
		fetchAllStatesAndPublish(context, params, params.ConnectionManager.Broadcast)
	})
}
