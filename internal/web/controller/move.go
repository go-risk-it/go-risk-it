package controller

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/api/game/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game"
	"github.com/go-risk-it/go-risk-it/internal/logic/move/attack"
	"github.com/go-risk-it/go-risk-it/internal/logic/move/deploy"
	"github.com/go-risk-it/go-risk-it/internal/logic/orchestration"
	"github.com/go-risk-it/go-risk-it/internal/riskcontext"
	"go.uber.org/zap"
)

type MoveController interface {
	PerformDeployMove(ctx riskcontext.MoveContext, deployMove request.DeployMove) error
	PerformAttackMove(ctx riskcontext.MoveContext, attackMove request.AttackMove) error
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

func (c *MoveControllerImpl) PerformDeployMove(
	ctx riskcontext.MoveContext,
	deployMove request.DeployMove,
) error {
	err := c.orchestrationService.OrchestrateMove(
		ctx,
		sqlc.PhaseDEPLOY,
		func(ctx riskcontext.MoveContext, querier db.Querier, game *sqlc.Game) error {
			err := c.deployService.PerformQ(ctx, querier, game, deploy.Move{
				RegionID:      deployMove.RegionID,
				CurrentTroops: deployMove.CurrentTroops,
				DesiredTroops: deployMove.DesiredTroops,
			})
			if err != nil {
				return fmt.Errorf("unable to perform deploy move: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return fmt.Errorf("unable to perform deploy move: %w", err)
	}

	return nil
}

func (c *MoveControllerImpl) PerformAttackMove(
	ctx riskcontext.MoveContext,
	attackMove request.AttackMove,
) error {
	err := c.orchestrationService.OrchestrateMove(
		ctx,
		sqlc.PhaseATTACK,
		func(ctx riskcontext.MoveContext, querier db.Querier, game *sqlc.Game) error {
			err := c.attackService.PerformQ(ctx, querier, game, attack.Move{
				SourceRegionID:  attackMove.SourceRegionID,
				TargetRegionID:  attackMove.TargetRegionID,
				TroopsInSource:  attackMove.TroopsInSource,
				TroopsInTarget:  attackMove.TroopsInTarget,
				AttackingTroops: attackMove.AttackingTroops,
			})
			if err != nil {
				return fmt.Errorf("unable to perform attack move: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return fmt.Errorf("unable to perform attack move: %w", err)
	}

	return nil
}
