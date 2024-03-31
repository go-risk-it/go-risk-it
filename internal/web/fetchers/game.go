package fetchers

import (
	"context"
	"encoding/json"

	"github.com/go-risk-it/go-risk-it/internal/web/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/ws/message"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type GameFetcher interface {
	Fetcher
}

type GameFetcherImpl struct {
	log            *zap.SugaredLogger
	gameController controller.GameController
}

type GameFetcherResult struct {
	fx.Out

	GameFetcher GameFetcher
	Fetcher     Fetcher `group:"fetchers"`
}

func NewGameFetcher(
	log *zap.SugaredLogger,
	gameController controller.GameController,
) GameFetcherResult {
	res := &GameFetcherImpl{
		log:            log,
		gameController: gameController,
	}

	return GameFetcherResult{
		GameFetcher: res,
		Fetcher:     res,
	}
}

func (f *GameFetcherImpl) FetchState(
	ctx context.Context,
	gameID int64,
	messageChannel chan json.RawMessage,
) {
	FetchState(
		ctx,
		f.log,
		gameID,
		message.GameState,
		f.gameController.GetGameState,
		messageChannel,
	)
}
