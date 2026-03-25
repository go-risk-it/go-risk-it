package fetcher

import (
	"encoding/json"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"go.uber.org/fx"
)

type Fetcher interface {
	FetchState(ctx ctx.GameContext, stateChannel chan json.RawMessage)
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
