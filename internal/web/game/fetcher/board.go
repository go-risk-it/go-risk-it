package fetcher

import (
	"encoding/json"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	sharedfetcher "github.com/go-risk-it/go-risk-it/internal/web/fetcher"
	"github.com/go-risk-it/go-risk-it/internal/web/game/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/ws/message"
	"go.uber.org/fx"
)

type BoardFetcher interface {
	Fetcher
}

type BoardFetcherImpl struct {
	boardController controller.BoardController
}

var _ BoardFetcher = (*BoardFetcherImpl)(nil)

type BoardFetcherResult struct {
	fx.Out

	BoardFetcher BoardFetcher
	Fetcher      Fetcher `group:"public_fetchers"`
}

func NewBoardFetcher(boardController controller.BoardController) BoardFetcherResult {
	res := &BoardFetcherImpl{
		boardController: boardController,
	}

	return BoardFetcherResult{
		BoardFetcher: res,
		Fetcher:      res,
	}
}

func (f *BoardFetcherImpl) FetchState(ctx ctx.GameContext, stateChannel chan json.RawMessage) {
	sharedfetcher.FetchState(
		ctx,
		message.BoardState,
		f.boardController.GetBoardState,
		stateChannel,
	)
}
