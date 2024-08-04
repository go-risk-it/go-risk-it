package phase

import (
	"encoding/json"

	"github.com/go-risk-it/go-risk-it/internal/api/game/messaging"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/state"
	"go.uber.org/fx"
)

type Fetcher interface {
	FetchState(ctx ctx.GameContext, game *state.Game, stateChannel chan json.RawMessage)
}

func getFetcherFunc[T messaging.PhaseState](
	game *state.Game,
	fetcherFunc func(ctx.GameContext, *state.Game) (messaging.GameState[T], error),
) func(context ctx.GameContext) (messaging.GameState[T], error) {
	return func(cont ctx.GameContext) (messaging.GameState[T], error) {
		return fetcherFunc(cont, game)
	}
}

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(NewDeployPhaseFetcher, fx.As(new(DeployPhaseFetcher))),
		fx.Annotate(NewAttackPhaseFetcher, fx.As(new(AttackPhaseFetcher))),
		fx.Annotate(NewConquerPhaseFetcher, fx.As(new(ConquerPhaseFetcher))),
	),
)
