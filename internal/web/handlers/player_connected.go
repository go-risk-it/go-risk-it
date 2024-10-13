package handlers

import (
	"context"

	"github.com/go-risk-it/go-risk-it/internal/logic/signals"
)

func HandlePlayerConnected(
	params HandlerParams[signals.PlayerConnectedSignal],
) {
	params.Signal.AddListener(func(context context.Context, data signals.PlayerConnectedData) {
		fetchAllStatesAndPublish(context, params, params.ConnectionManager.WriteMessage)
	})
}
