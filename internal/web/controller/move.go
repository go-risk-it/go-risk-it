package controller

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/api/game/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/attack"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/conquer"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/deploy"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/orchestration"
)

type MoveController interface {
	PerformDeployMove(ctx ctx.GameContext, deployMove request.DeployMove) error
	PerformAttackMove(ctx ctx.GameContext, attackMove request.AttackMove) error
	PerformConquerMove(ctx ctx.GameContext, conquerMove request.ConquerMove) error
}

type MoveControllerImpl struct {
	deployOrchestrator  orchestration.DeployOrchestrator
	attackOrchestrator  orchestration.AttackOrchestrator
	conquerOrchestrator orchestration.ConquerOrchestrator
}

var _ MoveController = (*MoveControllerImpl)(nil)

func NewMoveController(
	deployOrchestrator orchestration.DeployOrchestrator,
	attackOrchestrator orchestration.AttackOrchestrator,
	conquerOrchestrator orchestration.ConquerOrchestrator,
) *MoveControllerImpl {
	return &MoveControllerImpl{
		deployOrchestrator:  deployOrchestrator,
		attackOrchestrator:  attackOrchestrator,
		conquerOrchestrator: conquerOrchestrator,
	}
}

func (c *MoveControllerImpl) PerformDeployMove(
	ctx ctx.GameContext,
	deployMove request.DeployMove,
) error {
	move := deploy.Move{
		RegionID:      deployMove.RegionID,
		CurrentTroops: deployMove.CurrentTroops,
		DesiredTroops: deployMove.DesiredTroops,
	}

	err := c.deployOrchestrator.OrchestrateMove(ctx, move)
	if err != nil {
		return fmt.Errorf("unable to perform deploy move: %w", err)
	}

	return nil
}

func (c *MoveControllerImpl) PerformAttackMove(
	ctx ctx.GameContext,
	attackMove request.AttackMove,
) error {
	move := attack.Move{
		AttackingRegionID: attackMove.SourceRegionID,
		DefendingRegionID: attackMove.TargetRegionID,
		TroopsInSource:    attackMove.TroopsInSource,
		TroopsInTarget:    attackMove.TroopsInTarget,
		AttackingTroops:   attackMove.AttackingTroops,
	}

	err := c.attackOrchestrator.OrchestrateMove(ctx, move)
	if err != nil {
		return fmt.Errorf("unable to perform attack move: %w", err)
	}

	return nil
}

func (c *MoveControllerImpl) PerformConquerMove(
	ctx ctx.GameContext,
	conquerMove request.ConquerMove,
) error {
	move := conquer.Move{
		Troops: conquerMove.Troops,
	}

	err := c.conquerOrchestrator.OrchestrateMove(ctx, move)
	if err != nil {
		return fmt.Errorf("unable to perform conquer move: %w", err)
	}

	return nil
}
