package controller

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/api/game/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/attack"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/deploy"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/orchestration"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/phase"
)

type MoveController interface {
	PerformDeployMove(ctx ctx.MoveContext, deployMove request.DeployMove) error
	PerformAttackMove(ctx ctx.MoveContext, attackMove request.AttackMove) error
}

type MoveControllerImpl struct {
	attackService       attack.Service
	deployService       deploy.Service
	phaseService        phase.Service
	deployOrchestrator  orchestration.DeployOrchestrator
	attackOrchestrator  orchestration.AttackOrchestrator
	conquerOrchestrator orchestration.ConquerOrchestrator
}

var _ MoveController = (*MoveControllerImpl)(nil)

func NewMoveController(
	attackService attack.Service,
	deployService deploy.Service,
	phaseService phase.Service,
	deployOrchestrator orchestration.DeployOrchestrator,
	attackOrchestrator orchestration.AttackOrchestrator,
	conquerOrchestrator orchestration.ConquerOrchestrator,
) *MoveControllerImpl {
	return &MoveControllerImpl{
		attackService:       attackService,
		deployService:       deployService,
		phaseService:        phaseService,
		deployOrchestrator:  deployOrchestrator,
		attackOrchestrator:  attackOrchestrator,
		conquerOrchestrator: conquerOrchestrator,
	}
}

func (c *MoveControllerImpl) PerformDeployMove(
	ctx ctx.MoveContext,
	deployMove request.DeployMove,
) error {
	move := deploy.Move{
		RegionID:      deployMove.RegionID,
		CurrentTroops: deployMove.CurrentTroops,
		DesiredTroops: deployMove.DesiredTroops,
	}

	err := c.deployOrchestrator.OrchestrateMove(
		ctx,
		sqlc.PhaseTypeDEPLOY,
		c.deployService,
		move,
	)
	if err != nil {
		return fmt.Errorf("unable to perform deploy move: %w", err)
	}

	return nil
}

func (c *MoveControllerImpl) PerformAttackMove(
	ctx ctx.MoveContext,
	attackMove request.AttackMove,
) error {
	move := attack.Move{
		AttackingRegionID: attackMove.SourceRegionID,
		DefendingRegionID: attackMove.TargetRegionID,
		TroopsInSource:    attackMove.TroopsInSource,
		TroopsInTarget:    attackMove.TroopsInTarget,
		AttackingTroops:   attackMove.AttackingTroops,
	}

	err := c.attackOrchestrator.OrchestrateMove(
		ctx,
		sqlc.PhaseTypeATTACK,
		c.attackService,
		move,
	)
	if err != nil {
		return fmt.Errorf("unable to perform attack move: %w", err)
	}

	return nil
}
