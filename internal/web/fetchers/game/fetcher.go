package game

import (
	"context"
	"encoding/json"

	"github.com/tomfran/go-risk-it/internal/web/controllers/game"
	"github.com/tomfran/go-risk-it/internal/web/ws/message"
	"go.uber.org/zap"
)

type Fetcher struct {
	log            *zap.SugaredLogger
	gameController game.Controller
}

func NewFetcher(
	log *zap.SugaredLogger,
	gameController game.Controller,
) *Fetcher {
	return &Fetcher{
		log:            log,
		gameController: gameController,
	}
}

func (f *Fetcher) FetchState(
	ctx context.Context,
	gameID int64,
	messageChannel chan json.RawMessage,
) {
	f.log.Infow("fetching game state", "gameID", gameID)

	state, err := f.gameController.GetGameState(ctx, gameID)
	if err != nil {
		f.log.Errorf("unable to fetch state: %v", err)
	}

	f.log.Infow("got state", "state", state)

	rawResponse, err := message.BuildMessage(message.GameState, state)
	if err != nil {
		f.log.Errorf("unable to build message: %v", err)
	}

	messageChannel <- rawResponse
}
