package handlers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/tomfran/go-risk-it/internal/signals"
	"github.com/tomfran/go-risk-it/internal/web/fetchers"
	"github.com/tomfran/go-risk-it/internal/web/ws/connection"
	"go.uber.org/zap"
)

func HandleBoardStateChanged(
	log *zap.SugaredLogger,
	boardStateFetcher fetchers.BoardFetcher,
	connectionManager connection.Manager,
	signal signals.BoardStateChangedSignal,
) {
	signal.AddListener(func(ctx context.Context, data signals.BoardStateChangedData) {
		log.Infow("handling board state changed", "data", data)

		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		channel := make(chan json.RawMessage)
		boardStateFetcher.FetchState(ctx, data.GameID, channel)

		select {
		case message := <-channel:
			connectionManager.Broadcast(data.GameID, message)
		case <-ctx.Done():
			log.Errorf("timeout while fetching board state: %v", ctx.Err())
		}
	})
}
