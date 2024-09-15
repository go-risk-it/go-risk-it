package handlers

import (
	"context"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/logic/signals"
	"github.com/go-risk-it/go-risk-it/internal/web/fetchers/fetcher"
	"github.com/go-risk-it/go-risk-it/internal/web/ws/connection"
)

func HandleBoardStateChanged(
	boardStateFetcher fetcher.BoardFetcher,
	connectionManager connection.Manager,
	signal signals.BoardStateChangedSignal,
) {
	signal.AddListener(func(context context.Context, data signals.BoardStateChangedData) {
		gameContext, ok := context.(ctx.GameContext)
		if !ok {
			return
		}

		gameContext.Log().Infow("handling board state changed")
		fetchStateAndBroadcast(
			gameContext,
			boardStateFetcher.FetchState,
			connectionManager.Broadcast)
	})
}
