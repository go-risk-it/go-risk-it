package fetcher

import (
	"context"
	"encoding/json"
	"log/slog"
	"reflect"

	"github.com/go-risk-it/go-risk-it/internal/web/ws/message"
)

// FetchState fetches state using fetcherFunc and sends it as a message to stateChannel.
func FetchState[C context.Context, T any](
	ctx C,
	messageType message.Type,
	fetcherFunc func(C) (T, error),
	stateChannel chan json.RawMessage,
) {
	slog.InfoContext(ctx, "fetching state", "messageType", messageType)

	state, err := fetcherFunc(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "unable to fetch state", "error", err)

		return
	}

	slog.DebugContext(ctx, "got state, writing message", "type", reflect.TypeOf(state))

	rawResponse, err := message.BuildMessage(messageType, state)
	if err != nil {
		slog.ErrorContext(ctx, "unable to build message", "error", err)

		return
	}

	stateChannel <- rawResponse
}
