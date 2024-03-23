package handlers

import (
	"context"
	"encoding/json"

	"github.com/tomfran/go-risk-it/internal/signals"
	"github.com/tomfran/go-risk-it/internal/web/fetchers"
	"github.com/tomfran/go-risk-it/internal/web/ws/connection"
)

func HandleBoardStateChanged(
	boardStateFetcher fetchers.BoardFetcher,
	connectionManager connection.Manager,
	signal signals.BoardStateChangedSignal,
) {
	signal.AddListener(func(ctx context.Context, data signals.BoardStateChangedData) {
		channel := make(chan json.RawMessage)
		boardStateFetcher.FetchState(ctx, data.GameID, channel)
		connectionManager.Broadcast(data.GameID, <-channel)
	})
}
