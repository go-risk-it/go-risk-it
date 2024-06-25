package fetchers

import (
	"encoding/json"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/web/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/ws/message"
	"go.uber.org/fx"
)

type GameFetcher interface {
	Fetcher
}

type GameFetcherImpl struct {
	gameController controller.GameController
}

var _ GameFetcher = (*GameFetcherImpl)(nil)

type GameFetcherResult struct {
	fx.Out

	GameFetcher GameFetcher
	Fetcher     Fetcher `group:"fetchers"`
}

func NewGameFetcher(
	gameController controller.GameController,
) GameFetcherResult {
	res := &GameFetcherImpl{
		gameController: gameController,
	}

	return GameFetcherResult{
		GameFetcher: res,
		Fetcher:     res,
	}
}

func (f *GameFetcherImpl) FetchState(context ctx.GameContext, stateChannel chan json.RawMessage) {
	FetchState(
		context,
		message.GameState,
		f.gameController.GetGameState,
		stateChannel,
	)
}
