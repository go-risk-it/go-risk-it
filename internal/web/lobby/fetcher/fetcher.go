package fetcher

import (
	"encoding/json"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(NewLobbyStateFetcher),
)

type Fetcher interface {
	FetchState(ctx ctx.LobbyContext, stateChannel chan json.RawMessage)
}
