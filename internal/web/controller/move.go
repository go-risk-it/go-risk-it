package controller

import (
	"context"
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/api/game/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/logic/move/attack"
	"github.com/go-risk-it/go-risk-it/internal/logic/move/deploy"
	"github.com/go-risk-it/go-risk-it/internal/logic/move/move"
	"go.uber.org/zap"
)

type MoveController interface {
	PerformDeployMove(ctx context.Context, gameID int64, deployMove request.DeployMove) error
	PerformAttackMove(
		ctx context.Context, gameID int64, userID string, attackMove request.AttackMove,
	) error
}

type MoveControllerImpl struct {
	log           *zap.SugaredLogger
	deployService deploy.Service
	attackService attack.Service
}

func NewMoveController(
	log *zap.SugaredLogger,
	deployService deploy.Service,
	attackService attack.Service,
) *MoveControllerImpl {
	return &MoveControllerImpl{
		log:           log,
		deployService: deployService,
		attackService: attackService,
	}
}

func performMove[T any](
	ctx context.Context,
	move move.Move[T],
	performer move.Performer[T],
) error {
	// validate move
	if err := performer.Perform(ctx, move); err != nil {
		return fmt.Errorf("unable to perform move: %w", err)
	}

	return nil
}

func (c *MoveControllerImpl) PerformDeployMove(
	ctx context.Context, gameID int64, userID string, deployMove request.DeployMove,
) error {
	err := performMove(ctx, move.Move[deploy.MoveData]{
		UserID: userID,
		GameID: gameID,
		Payload: deploy.MoveData{
			RegionID:      deployMove.RegionID,
			CurrentTroops: deployMove.CurrentTroops,
			DesiredTroops: deployMove.DesiredTroops,
		},
	}, c.deployService)
	if err != nil {
		return fmt.Errorf("unable to perform deploy move: %w", err)
	}

	return nil
}

func (c *MoveControllerImpl) PerformAttackMove(
	ctx context.Context, gameID int64, userID string, attackMove request.AttackMove,
) error {
	err := performMove(ctx, move.Move[attack.MoveData]{
		UserID: userID,
		GameID: gameID,
		Payload: attack.MoveData{
			SourceRegionID:  attackMove.SourceRegionID,
			TargetRegionID:  attackMove.TargetRegionID,
			TroopsInSource:  attackMove.TroopsInSource,
			TroopsInTarget:  attackMove.TroopsInTarget,
			AttackingTroops: attackMove.AttackingTroops,
		},
	}, c.attackService)
	if err != nil {
		return fmt.Errorf("unable to perform attack move: %w", err)
	}

	return nil
}
