package fetcher

import (
	"encoding/json"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	sharedfetcher "github.com/go-risk-it/go-risk-it/internal/web/fetcher"
	"github.com/go-risk-it/go-risk-it/internal/web/game/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/ws/message"
	"go.uber.org/fx"
)

type playerFetcher struct {
	playerController *controller.PlayerController
}

type PlayerFetcherResult struct {
	fx.Out

	Fetcher Fetcher `group:"public_fetchers"`
}

func NewPlayerFetcher(
	playerController *controller.PlayerController,
) PlayerFetcherResult {
	return PlayerFetcherResult{
		Fetcher: &playerFetcher{
			playerController: playerController,
		},
	}
}

func (f *playerFetcher) FetchState(
	context ctx.GameContext,
	stateChannel chan json.RawMessage,
) {
	sharedfetcher.FetchState(
		context,
		message.PlayerState,
		f.playerController.GetPlayerState,
		stateChannel,
	)
}
