package fetcher

import (
	"encoding/json"
	"reflect"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/web/ws/message"
	"go.uber.org/fx"
)

type Fetcher interface {
	FetchState(ctx ctx.GameContext, stateChannel chan json.RawMessage)
}

func FetchState[T any](
	ctx ctx.GameContext,
	messageType message.Type,
	fetcherFunc func(ctx.GameContext) (T, error),
	stateChannel chan json.RawMessage,
) {
	ctx.Log().Infow("fetching state", "messageType", messageType)

	state, err := fetcherFunc(ctx)
	if err != nil {
		ctx.Log().Errorf("unable to fetch state: %v", err)
	}

	ctx.Log().Debugf("got state %v, writing message", reflect.TypeOf(state))

	rawResponse, err := message.BuildMessage(messageType, state)
	if err != nil {
		ctx.Log().Errorf("unable to build message: %v", err)
	}

	stateChannel <- rawResponse
}

var Module = fx.Options(
	fx.Provide(
		NewBoardFetcher,
		NewPlayerFetcher,
		NewCardFetcher,
		NewMoveLogFetcher,
		NewMissionFetcher,
		NewGameFetcher,
	),
)
