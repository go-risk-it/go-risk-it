package fetchers

import (
	"github.com/go-risk-it/go-risk-it/internal/web/game/fetchers/fetcher"
	"github.com/go-risk-it/go-risk-it/internal/web/game/fetchers/phase"
	"go.uber.org/fx"
)

var Module = fx.Options(
	phase.Module,
	fx.Provide(
		fetcher.NewBoardFetcher,
		fetcher.NewPlayerFetcher,
		fetcher.NewCardFetcher,
		fetcher.NewMoveLogFetcher,
	),
)
