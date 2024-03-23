package fetchers

import (
	"context"
	"encoding/json"

	"github.com/tomfran/go-risk-it/internal/web/controller"
	"github.com/tomfran/go-risk-it/internal/web/ws/message"
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

func (f *BoardFetcherImpl) FetchState(
	ctx context.Context,
	gameID int64,
	messageChannel chan json.RawMessage,
) {
	FetchState(
		ctx,
		f.log,
		gameID,
		message.BoardState,
		f.boardController.GetBoardState,
		messageChannel,
	)
}
