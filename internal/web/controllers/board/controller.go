package board

import (
	"github.com/tomfran/go-risk-it/internal/api/game/message"
	"github.com/tomfran/go-risk-it/internal/logic/board"
	"go.uber.org/zap"
)

type Controller interface {
	GetBoardState(gameID int64) (message.BoardState, error)
}

type ControllerImpl struct {
	log          *zap.SugaredLogger
	boardService board.Service
}

func New(log *zap.SugaredLogger, boardService board.Service) *ControllerImpl {
	return &ControllerImpl{log: log, boardService: boardService}
}

func (c *ControllerImpl) GetBoardState(
	gameID int64,
) (message.BoardState, error) {
	return message.BoardState{Regions: []message.Region{}}, nil
}
