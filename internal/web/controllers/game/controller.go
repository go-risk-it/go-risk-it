package game

import (
	gameApi "github.com/tomfran/go-risk-it/internal/api/game"
	"github.com/tomfran/go-risk-it/internal/logic/game"
	"go.uber.org/zap"
)

type Controller interface {
	GetGameState(request gameApi.GameStateRequest) gameApi.GameStateResponse
}

type ControllerImpl struct {
	log         *zap.SugaredLogger
	gameService game.Service
}

func New(log *zap.SugaredLogger, gameService game.Service) *ControllerImpl {
	return &ControllerImpl{log: log, gameService: gameService}
}

func (c *ControllerImpl) GetGameState(request gameApi.GameStateRequest) gameApi.GameStateResponse {
	return gameApi.GameStateResponse{UserID: 1, GameID: 1}
}
