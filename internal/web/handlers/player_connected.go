package handlers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/signals"
	"github.com/go-risk-it/go-risk-it/internal/web/fetchers"
	"github.com/lesismal/nbio/nbhttp/websocket"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type PlayerConnectedHandlerParams struct {
	fx.In

	Fetchers []fetchers.Fetcher `group:"fetchers"`
	Log      *zap.SugaredLogger
	Signal   signals.PlayerConnectedSignal
}

func HandlePlayerConnected(
	params PlayerConnectedHandlerParams,
) {
	params.Signal.AddListener(func(ctx context.Context, data signals.PlayerConnectedData) {
		params.Log.Infow("handling player connected",
			"gameID", data.GameID,
			"remoteAddress", data.Connection.RemoteAddr().String())

		childCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		stateChannel := make(chan json.RawMessage, len(params.Fetchers))

		params.Log.Infow("fetching states", "count", len(params.Fetchers))
		for _, fetcher := range params.Fetchers {
			go fetcher.FetchState(childCtx, data.GameID, stateChannel)
		}

		for i := 0; i < len(params.Fetchers); i++ {
			select {
			case state := <-stateChannel:
				params.Log.Infow("got state, writing message", "gameID", data.GameID)

				err := data.Connection.WriteMessage(websocket.TextMessage, state)
				if err != nil {
					params.Log.Errorf("unable to write response: %v", err)
				}
			case <-childCtx.Done():
				params.Log.Errorf("unable to get all states: %v", childCtx.Err())

				return
			}
		}
	})
}
