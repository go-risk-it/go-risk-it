package fetcher

import (
	"encoding/json"
	"reflect"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/web/ws/message"
)

// FetchState fetches state using fetcherFunc and sends it as a message to stateChannel.
func FetchState[C ctx.LogContext, T any](
	logContext C,
	messageType message.Type,
	fetcherFunc func(C) (T, error),
	stateChannel chan json.RawMessage,
) {
	logContext.Log().Infow("fetching state", "messageType", messageType)

	state, err := fetcherFunc(logContext)
	if err != nil {
		logContext.Log().Errorf("unable to fetch state: %v", err)
	}

	logContext.Log().Debugf("got state %v, writing message", reflect.TypeOf(state))

	rawResponse, err := message.BuildMessage(messageType, state)
	if err != nil {
		logContext.Log().Errorf("unable to build message: %v", err)
	}

	stateChannel <- rawResponse
}
