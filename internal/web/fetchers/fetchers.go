package fetchers

import (
	"context"
	"encoding/json"

	"github.com/tomfran/go-risk-it/internal/web/fetchers/board"
	"github.com/tomfran/go-risk-it/internal/web/fetchers/game"
	"github.com/tomfran/go-risk-it/internal/web/fetchers/player"
	"go.uber.org/fx"
)

type Fetcher interface {
	FetchState(ctx context.Context, gameID int64, stateChannel chan json.RawMessage)
}

func AsFetcher(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(Fetcher)),
		fx.ResultTags(`group:"fetchers"`),
	)
}

var Module = fx.Options(
	fx.Provide(
		AsFetcher(game.NewFetcher),
		AsFetcher(board.NewFetcher),
		AsFetcher(player.NewFetcher),
	),
)
