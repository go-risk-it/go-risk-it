package handlers

import (
	"context"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/logic/signals"
	"github.com/go-risk-it/go-risk-it/internal/web/fetchers/fetcher"
	"github.com/go-risk-it/go-risk-it/internal/web/ws/connection"
)

func HandlePlayerStateChanged(
	playerStateFetcher fetcher.PlayerFetcher,
	connectionManager connection.Manager,
	signal signals.PlayerStateChangedSignal,
) {
	signal.AddListener(func(context context.Context, data signals.PlayerStateChangedData) {
		gameContext, ok := context.(ctx.GameContext)
		if !ok {
			return
		}

		gameContext.Log().Infow("handling player state changed")
		fetchStateAndBroadcast(
			gameContext,
			playerStateFetcher.FetchState,
			connectionManager.Broadcast)
	})
}
