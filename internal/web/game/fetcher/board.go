package fetcher

import (
	"encoding/json"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	sharedfetcher "github.com/go-risk-it/go-risk-it/internal/web/fetcher"
	"github.com/go-risk-it/go-risk-it/internal/web/game/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/ws/message"
	"go.uber.org/fx"
)

type boardFetcher struct {
	boardController *controller.BoardController
}

type BoardFetcherResult struct {
	fx.Out

	Fetcher Fetcher `group:"public_fetchers"`
}

func NewBoardFetcher(boardController *controller.BoardController) BoardFetcherResult {
	return BoardFetcherResult{
		Fetcher: &boardFetcher{
			boardController: boardController,
		},
	}
}

func (f *boardFetcher) FetchState(ctx ctx.GameContext, stateChannel chan json.RawMessage) {
	sharedfetcher.FetchState(
		ctx,
		message.BoardState,
		f.boardController.GetBoardState,
		stateChannel,
	)
}
