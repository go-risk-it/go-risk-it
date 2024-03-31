package controller

import (
	"context"
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/api/game/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/logic/move/deploy"
	"go.uber.org/zap"
)

type MoveController interface {
	PerformDeployMove(ctx context.Context, gameID int64, deployMove request.DeployMove) error
}

type MoveControllerImpl struct {
	log           *zap.SugaredLogger
	deployService deploy.Service
}

func NewMoveController(
	log *zap.SugaredLogger,
	deployService deploy.Service,
) *MoveControllerImpl {
	return &MoveControllerImpl{log: log, deployService: deployService}
}

func (c *MoveControllerImpl) PerformDeployMove(
	ctx context.Context, gameID int64, deployMove request.DeployMove,
) error {
	if err := c.deployService.PerformDeployMoveWithTx(
		ctx,
		gameID,
		deployMove.PlayerID,
		deployMove.RegionID,
		deployMove.Troops,
	); err != nil {
		return fmt.Errorf("unable to perform deploy move: %w", err)
	}

	return nil
}
