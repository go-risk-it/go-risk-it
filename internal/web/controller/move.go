package controller

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/api/game/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game"
	"github.com/go-risk-it/go-risk-it/internal/logic/move/orchestration"
	"github.com/go-risk-it/go-risk-it/internal/logic/move/performer/attack"
	"github.com/go-risk-it/go-risk-it/internal/logic/move/performer/deploy"
	"github.com/go-risk-it/go-risk-it/internal/logic/move/performer/service"
	"go.uber.org/zap"
)

type MoveController interface {
	PerformDeployMove(ctx ctx.MoveContext, deployMove request.DeployMove) error
	PerformAttackMove(ctx ctx.MoveContext, attackMove request.AttackMove) error
}

type MoveControllerImpl struct {
	log                  *zap.SugaredLogger
	attackService        attack.Service
	deployService        deploy.Service
	gameService          game.Service
	orchestrationService orchestration.Service
}

func NewMoveController(
	log *zap.SugaredLogger,
	attackService attack.Service,
	deployService deploy.Service,
	gameService game.Service,
	orchestrationService orchestration.Service,
) *MoveControllerImpl {
	return &MoveControllerImpl{
		log:                  log,
		attackService:        attackService,
		deployService:        deployService,
		gameService:          gameService,
		orchestrationService: orchestrationService,
	}
}

func getPerformerFunc[T any](
	performer service.Performer[T],
	move T,
) func(ctx ctx.MoveContext, querier db.Querier, game *sqlc.Game) error {
	return func(ctx ctx.MoveContext, querier db.Querier, game *sqlc.Game) error {
		err := performer.PerformQ(ctx, querier, game, move)
		if err != nil {
			return fmt.Errorf("unable to perform move: %w", err)
		}

		return nil
	}
}

func (c *MoveControllerImpl) PerformDeployMove(
	ctx ctx.MoveContext,
	deployMove request.DeployMove,
) error {
	err := c.orchestrationService.OrchestrateMove(
		ctx,
		sqlc.PhaseDEPLOY,
		getPerformerFunc(c.deployService, deploy.Move{
			RegionID:      deployMove.RegionID,
			CurrentTroops: deployMove.CurrentTroops,
			DesiredTroops: deployMove.DesiredTroops,
		}),
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
	err := c.orchestrationService.OrchestrateMove(
		ctx,
		sqlc.PhaseATTACK,
		getPerformerFunc(c.attackService, attack.Move{
			SourceRegionID:  attackMove.SourceRegionID,
			TargetRegionID:  attackMove.TargetRegionID,
			TroopsInSource:  attackMove.TroopsInSource,
			TroopsInTarget:  attackMove.TroopsInTarget,
			AttackingTroops: attackMove.AttackingTroops,
		}),
	)
	if err != nil {
		return fmt.Errorf("unable to perform attack move: %w", err)
	}

	return nil
}
