package controller

import (
	"context"
	"fmt"

	"github.com/tomfran/go-risk-it/internal/api/game/rest/request"
	"github.com/tomfran/go-risk-it/internal/logic/move"
	"go.uber.org/zap"
)

type MoveController interface {
	PerformDeployMove(ctx context.Context, deployMove request.DeployMove) error
}

type MoveControllerImpl struct {
	log         *zap.SugaredLogger
	moveService move.Service
}

func NewMoveController(
	log *zap.SugaredLogger,
	moveService move.Service,
) *MoveControllerImpl {
	return &MoveControllerImpl{log: log, moveService: moveService}
}

func (c *MoveControllerImpl) PerformDeployMove(
	ctx context.Context, deployMove request.DeployMove,
) error {
	if err := c.moveService.PerformDeployMoveWithTx(
		ctx,
		deployMove.GameID,
		deployMove.PlayerID,
		deployMove.RegionID,
		deployMove.Troops,
	); err != nil {
		return fmt.Errorf("unable to perform deploy move: %w", err)
	}

	return nil
}
