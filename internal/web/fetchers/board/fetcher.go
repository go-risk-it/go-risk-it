package board

import (
	"context"
	"encoding/json"

	"github.com/tomfran/go-risk-it/internal/web/controllers/board"
	"github.com/tomfran/go-risk-it/internal/web/ws/message"
	"go.uber.org/zap"
)

type Fetcher struct {
	log             *zap.SugaredLogger
	boardController board.Controller
}

func NewFetcher(
	log *zap.SugaredLogger,
	boardController board.Controller,
) *Fetcher {
	return &Fetcher{
		log:             log,
		boardController: boardController,
	}
}

func (f *Fetcher) FetchState(
	ctx context.Context,
	gameID int64,
	messageChannel chan json.RawMessage,
) {
	f.log.Infow("fetching board state", "gameID", gameID)

	state, err := f.boardController.GetBoardState(ctx, gameID)
	if err != nil {
		f.log.Errorf("unable to fetch state: %v", err)
	}

	f.log.Infow("got state", "state", state)

	rawResponse, err := message.BuildMessage(message.BoardState, state)
	if err != nil {
		f.log.Errorf("unable to build message: %v", err)
	}

	messageChannel <- rawResponse
}
