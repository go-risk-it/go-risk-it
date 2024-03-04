package player

import (
	"context"
	"encoding/json"

	"github.com/tomfran/go-risk-it/internal/web/controllers/player"
	"github.com/tomfran/go-risk-it/internal/web/ws/message"
	"go.uber.org/zap"
)

type Fetcher struct {
	log              *zap.SugaredLogger
	playerController player.Controller
}

func NewFetcher(
	log *zap.SugaredLogger,
	playerController player.Controller,
) *Fetcher {
	return &Fetcher{
		log:              log,
		playerController: playerController,
	}
}

func (f *Fetcher) FetchState(
	ctx context.Context,
	gameID int64,
	messageChannel chan json.RawMessage,
) {
	f.log.Infow("fetching player state", "gameID", gameID)

	state, err := f.playerController.GetPlayerState(ctx, gameID)
	if err != nil {
		f.log.Errorf("unable to fetch state: %v", err)
	}

	f.log.Infow("got state", "state", state)

	rawResponse, err := message.BuildMessage(message.PlayerState, state)
	if err != nil {
		f.log.Errorf("unable to build message: %v", err)
	}

	messageChannel <- rawResponse
}
