package board

import (
	"github.com/tomfran/go-risk-it/internal/api/game/message/request"
	gameApi "github.com/tomfran/go-risk-it/internal/api/game/message/response"
	"github.com/tomfran/go-risk-it/internal/logic/board"
	"go.uber.org/zap"
)

type Controller interface {
	GetBoardState(request request.BoardStateRequest) (gameApi.BoardStateResponse, error)
}

type ControllerImpl struct {
	log          *zap.SugaredLogger
	boardService board.Service
}

func New(log *zap.SugaredLogger, boardService board.Service) *ControllerImpl {
	return &ControllerImpl{log: log, boardService: boardService}
}

func (c *ControllerImpl) GetBoardState(
	request request.BoardStateRequest,
) (gameApi.BoardStateResponse, error) {
	return gameApi.BoardStateResponse{Regions: []gameApi.Region{}}, nil
}
