package controller

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/api/game/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/attack"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/deploy"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/orchestration"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/service"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/phase"
)

type MoveController interface {
	PerformDeployMove(ctx ctx.MoveContext, deployMove request.DeployMove) error
	PerformAttackMove(ctx ctx.MoveContext, attackMove request.AttackMove) error
}

type MoveControllerImpl struct {
	attackService        attack.Service
	deployService        deploy.Service
	phaseService         phase.Service
	orchestrationService orchestration.Service
}

var _ MoveController = (*MoveControllerImpl)(nil)

func NewMoveController(
	attackService attack.Service,
	deployService deploy.Service,
	phaseService phase.Service,
	orchestrationService orchestration.Service,
) *MoveControllerImpl {
	return &MoveControllerImpl{
		attackService:        attackService,
		deployService:        deployService,
		phaseService:         phaseService,
		orchestrationService: orchestrationService,
	}
}

func getPerformerFunc[T any](
	performer service.Performer[T],
	move T,
) func(ctx ctx.MoveContext, querier db.Querier) error {
	return func(ctx ctx.MoveContext, querier db.Querier) error {
		err := performer.PerformQ(ctx, querier, move)
		if err != nil {
			return fmt.Errorf("unable to perform move: %w", err)
		}

		return nil
	}
}

func getAdvancerFunc[T any](
	advancer service.Advancer[T],
	move T,
) func(ctx ctx.MoveContext, querier db.Querier, phaseType sqlc.PhaseType) error {
	return func(ctx ctx.MoveContext, querier db.Querier, phaseType sqlc.PhaseType) error {
		err := advancer.AdvanceQ(ctx, querier, phaseType, move)
		if err != nil {
			return fmt.Errorf("unable to advance move: %w", err)
		}

		return nil
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

	err := c.orchestrationService.OrchestrateMove(
		ctx,
		sqlc.PhaseTypeDEPLOY,
		getPerformerFunc(c.deployService, move),
		c.deployService.Walk,
		getAdvancerFunc(c.deployService, move),
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

	err := c.orchestrationService.OrchestrateMove(
		ctx,
		sqlc.PhaseTypeATTACK,
		getPerformerFunc(c.attackService, move),
		c.attackService.Walk,
		getAdvancerFunc(c.attackService, move),
	)
	if err != nil {
		return fmt.Errorf("unable to perform attack move: %w", err)
	}

	return nil
}
