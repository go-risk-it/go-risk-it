package phase

import (
	"encoding/json"

	"github.com/go-risk-it/go-risk-it/internal/api/game/message"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/state"
	wsmessage "github.com/go-risk-it/go-risk-it/internal/web/ws/message"
	"go.uber.org/fx"
)

type Fetcher interface {
	FetchState(ctx ctx.GameContext, game *state.Game, stateChannel chan json.RawMessage)
}

func FetchState[T message.PhaseState](
	ctx ctx.GameContext,
	game *state.Game,
	messageType wsmessage.Type,
	fetcherFunc func(ctx.GameContext, *state.Game) (message.GameState[T], error),
	stateChannel chan json.RawMessage,
) {
	ctx.Log().Infow("fetching gameState", "messageType", messageType)

	gameState, err := fetcherFunc(ctx, game)
	if err != nil {
		ctx.Log().Errorf("unable to fetch gameState: %v", err)
	}

	ctx.Log().Debugw("got gameState")

	rawResponse, err := wsmessage.BuildMessage(messageType, gameState)
	if err != nil {
		ctx.Log().Errorf("unable to build message: %v", err)
	}

	stateChannel <- rawResponse
}

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(NewDeployPhaseFetcher, fx.As(new(Fetcher))),
	),
)
