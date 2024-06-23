package fetchers

import (
	"context"
	"encoding/json"

	"github.com/go-risk-it/go-risk-it/internal/web/ws/message"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Fetcher interface {
	FetchState(ctx context.Context, gameID int64, stateChannel chan json.RawMessage)
}

func FetchState[T any](
	ctx context.Context,
	log *zap.SugaredLogger,
	gameID int64,
	messageType message.Type,
	fetcherFunc func(context.Context, int64) (T, error),
	stateChannel chan json.RawMessage,
) {
	log.Infow("fetching state", "gameID", gameID, "messageType", messageType)

	state, err := fetcherFunc(ctx, gameID)
	if err != nil {
		log.Errorf("unable to fetch state: %v", err)
	}

	log.Debugw("got state", "gameID", gameID)

	rawResponse, err := message.BuildMessage(messageType, state)
	if err != nil {
		log.Errorf("unable to build message: %v", err)
	}

	stateChannel <- rawResponse
}

var Module = fx.Options(
	fx.Provide(
		NewGameFetcher,
		NewBoardFetcher,
		NewPlayerFetcher,
	),
)
