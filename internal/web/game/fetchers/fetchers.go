package fetchers

import (
	fetcher2 "github.com/go-risk-it/go-risk-it/internal/web/game/fetchers/fetcher"
	"github.com/go-risk-it/go-risk-it/internal/web/game/fetchers/phase"
	"go.uber.org/fx"
)

var Module = fx.Options(
	phase.Module,
	fx.Provide(
		fetcher2.NewBoardFetcher,
		fetcher2.NewPlayerFetcher,
		fetcher2.NewCardFetcher,
		fetcher2.NewMoveLogFetcher,
	),
)
