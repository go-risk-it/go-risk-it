package game

import (
	"github.com/tomfran/go-risk-it/internal/api/game/message/request"
	gameApi "github.com/tomfran/go-risk-it/internal/api/game/message/response"
	"github.com/tomfran/go-risk-it/internal/logic/game"
	"go.uber.org/zap"
)

type Controller interface {
	GetGameState(request request.GameStateRequest) (gameApi.GameStateResponse, error)
}

type ControllerImpl struct {
	log         *zap.SugaredLogger
	gameService game.Service
}

func New(log *zap.SugaredLogger, gameService game.Service) *ControllerImpl {
	return &ControllerImpl{log: log, gameService: gameService}
}

func (c *ControllerImpl) GetGameState(
	request request.GameStateRequest,
) (gameApi.GameStateResponse, error) {
	return gameApi.GameStateResponse{UserID: 1, GameID: 1}, nil
}
