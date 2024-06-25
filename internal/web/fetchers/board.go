package fetchers

import (
	"encoding/json"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/web/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/ws/message"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type BoardFetcher interface {
	Fetcher
}

type BoardFetcherImpl struct {
	log             *zap.SugaredLogger
	boardController controller.BoardController
}

type BoardFetcherResult struct {
	fx.Out

	BoardFetcher BoardFetcher
	Fetcher      Fetcher `group:"fetchers"`
}

func NewBoardFetcher(
	log *zap.SugaredLogger,
	boardController controller.BoardController,
) BoardFetcherResult {
	res := &BoardFetcherImpl{
		log:             log,
		boardController: boardController,
	}

	return BoardFetcherResult{
		BoardFetcher: res,
		Fetcher:      res,
	}
}

func (f *BoardFetcherImpl) FetchState(ctx ctx.GameContext, stateChannel chan json.RawMessage) {
	FetchState(
		ctx,
		message.BoardState,
		f.boardController.GetBoardState,
		stateChannel,
	)
}
