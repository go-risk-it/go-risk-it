package fetchers

import (
	"context"
	"encoding/json"

	"github.com/tomfran/go-risk-it/internal/web/controller"
	"github.com/tomfran/go-risk-it/internal/web/ws/message"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type PlayerFetcher interface {
	Fetcher
}
type PlayerFetcherImpl struct {
	log              *zap.SugaredLogger
	playerController controller.PlayerController
}

type PlayerFetcherResult struct {
	fx.Out

	PlayerFetcher PlayerFetcher
	Fetcher       Fetcher `group:"fetchers"`
}

func NewPlayerFetcher(
	log *zap.SugaredLogger,
	playerController controller.PlayerController,
) PlayerFetcherResult {
	res := &PlayerFetcherImpl{
		log:              log,
		playerController: playerController,
	}

	return PlayerFetcherResult{
		PlayerFetcher: res,
		Fetcher:       res,
	}
}

func (f *PlayerFetcherImpl) FetchState(
	ctx context.Context,
	gameID int64,
	messageChannel chan json.RawMessage,
) {
	FetchState(
		ctx,
		f.log,
		gameID,
		message.PlayerState,
		f.playerController.GetPlayerState,
		messageChannel,
	)
}
