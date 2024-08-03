package phase

import (
	"encoding/json"

	"github.com/go-risk-it/go-risk-it/internal/api/game/messaging"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/state"
	wsmessage "github.com/go-risk-it/go-risk-it/internal/web/ws/message"
	"go.uber.org/fx"
)

type Fetcher interface {
	FetchState(ctx ctx.GameContext, game *state.Game, stateChannel chan json.RawMessage)
}

func FetchState[T messaging.PhaseState](
	ctx ctx.GameContext,
	game *state.Game,
	messageType wsmessage.Type,
	fetcherFunc func(ctx.GameContext, *state.Game) (messaging.GameState[T], error),
	stateChannel chan json.RawMessage,
) {
	ctx.Log().Infow("fetching gameState", "messageType", messageType)

	gameState, err := fetcherFunc(ctx, game)
	if err != nil {
		ctx.Log().Errorf("unable to fetch gameState: %v", err)
	}

	ctx.Log().Debugw("got gameState", "gameState", gameState)

	rawResponse, err := wsmessage.BuildMessage(messageType, gameState)
	if err != nil {
		ctx.Log().Errorf("unable to build message: %v", err)
	}

	stateChannel <- rawResponse
}

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(NewDeployPhaseFetcher, fx.As(new(DeployPhaseFetcher))),
		fx.Annotate(NewAttackPhaseFetcher, fx.As(new(AttackPhaseFetcher))),
		fx.Annotate(NewConquerPhaseFetcher, fx.As(new(ConquerPhaseFetcher))),
	),
)
