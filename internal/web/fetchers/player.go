package fetchers

import (
	"encoding/json"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/web/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/ws/message"
	"go.uber.org/fx"
)

type PlayerFetcher interface {
	Fetcher
}
type PlayerFetcherImpl struct {
	playerController controller.PlayerController
}

type PlayerFetcherResult struct {
	fx.Out

	PlayerFetcher PlayerFetcher
	Fetcher       Fetcher `group:"fetchers"`
}

func NewPlayerFetcher(
	playerController controller.PlayerController,
) PlayerFetcherResult {
	res := &PlayerFetcherImpl{
		playerController: playerController,
	}

	return PlayerFetcherResult{
		PlayerFetcher: res,
		Fetcher:       res,
	}
}

func (f *PlayerFetcherImpl) FetchState(
	context ctx.GameContext,
	stateChannel chan json.RawMessage,
) {
	FetchState(
		context,
		message.PlayerState,
		f.playerController.GetPlayerState,
		stateChannel,
	)
}
