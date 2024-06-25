package handlers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/logic/signals"
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
	params.Signal.AddListener(func(cont context.Context, data signals.PlayerConnectedData) {
		childCtx, cancel := context.WithTimeout(cont, 10*time.Second)
		defer cancel()

		gameContext := ctx.WithGameID(ctx.WithLog(childCtx, params.Log), data.GameID)

		gameContext.Log().Infow("handling player connected",
			"remoteAddress", data.Connection.RemoteAddr().String())

		stateChannel := make(chan json.RawMessage, len(params.Fetchers))

		gameContext.Log().Infow("fetching states", "count", len(params.Fetchers))
		for _, fetcher := range params.Fetchers {
			go fetcher.FetchState(gameContext, stateChannel)
		}

		for i := 0; i < len(params.Fetchers); i++ {
			select {
			case state := <-stateChannel:
				gameContext.Log().Infow("got state, writing message")

				err := data.Connection.WriteMessage(websocket.TextMessage, state)
				if err != nil {
					gameContext.Log().Errorf("unable to write response: %v", err)
				}
			case <-childCtx.Done():
				gameContext.Log().Errorf("unable to get all states: %v", childCtx.Err())

				return
			}
		}
	})
}
