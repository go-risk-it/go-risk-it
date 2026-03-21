package controller

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/api/game/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/attack"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/cards"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/conquer"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/deploy"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/orchestration"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/reinforce"
)

type MoveController interface {
	PerformDeployMove(ctx ctx.GameContext, deployMove request.DeployMove) error
	PerformAttackMove(ctx ctx.GameContext, attackMove request.AttackMove) error
	PerformConquerMove(ctx ctx.GameContext, conquerMove request.ConquerMove) error
	PerformReinforceMove(ctx ctx.GameContext, reinforceMove request.ReinforceMove) error
	PerformCardsMove(ctx ctx.GameContext, cardsMove request.CardsMove) error
}

type MoveControllerImpl struct {
	deployOrchestrator    orchestration.DeployOrchestrator
	attackOrchestrator    orchestration.AttackOrchestrator
	conquerOrchestrator   orchestration.ConquerOrchestrator
	reinforceOrchestrator orchestration.ReinforceOrchestrator
	cardsOrchestrator     orchestration.CardsOrchestrator
}

var _ MoveController = (*MoveControllerImpl)(nil)

func NewMoveController(
	deployOrchestrator orchestration.DeployOrchestrator,
	attackOrchestrator orchestration.AttackOrchestrator,
	conquerOrchestrator orchestration.ConquerOrchestrator,
	reinforceOrchestrator orchestration.ReinforceOrchestrator,
	cardsOrchestrator orchestration.CardsOrchestrator,
) *MoveControllerImpl {
	return &MoveControllerImpl{
		deployOrchestrator:    deployOrchestrator,
		attackOrchestrator:    attackOrchestrator,
		conquerOrchestrator:   conquerOrchestrator,
		reinforceOrchestrator: reinforceOrchestrator,
		cardsOrchestrator:     cardsOrchestrator,
	}
}

func performMove[Move any](
	ctx ctx.GameContext,
	move Move,
	orchestrator orchestration.Orchestrator[Move, any],
) error {
	if err := orchestrator.OrchestrateMove(ctx, move); err != nil {
		return fmt.Errorf("unable to perform move: %w", err)
	}

	return nil
}

func (c *MoveControllerImpl) PerformDeployMove(
	ctx ctx.GameContext,
	deployMove request.DeployMove,
) error {
	return performMove(ctx, mapDeployMove(deployMove), c.deployOrchestrator)
}

func (c *MoveControllerImpl) PerformAttackMove(
	ctx ctx.GameContext,
	attackMove request.AttackMove,
) error {
	return performMove(ctx, mapAttackMove(attackMove), c.attackOrchestrator)
}

func (c *MoveControllerImpl) PerformConquerMove(
	ctx ctx.GameContext,
	conquerMove request.ConquerMove,
) error {
	return performMove(ctx, mapConquerMove(conquerMove), c.conquerOrchestrator)
}

func (c *MoveControllerImpl) PerformReinforceMove(
	ctx ctx.GameContext,
	reinforceMove request.ReinforceMove,
) error {
	return performMove(ctx, mapReinforceMove(reinforceMove), c.reinforceOrchestrator)
}

func (c *MoveControllerImpl) PerformCardsMove(
	ctx ctx.GameContext,
	cardsMove request.CardsMove,
) error {
	return performMove(ctx, mapCardsMove(cardsMove), c.cardsOrchestrator)
}

func mapDeployMove(req request.DeployMove) deploy.Move {
	return deploy.Move{
		RegionID:      req.RegionID,
		CurrentTroops: req.CurrentTroops,
		DesiredTroops: req.DesiredTroops,
	}
}

func mapAttackMove(req request.AttackMove) attack.Move {
	return attack.Move{
		AttackingRegionID: req.SourceRegionID,
		DefendingRegionID: req.TargetRegionID,
		TroopsInSource:    req.TroopsInSource,
		TroopsInTarget:    req.TroopsInTarget,
		AttackingTroops:   req.AttackingTroops,
	}
}

func mapConquerMove(req request.ConquerMove) conquer.Move {
	return conquer.Move{
		Troops: req.Troops,
	}
}

func mapReinforceMove(req request.ReinforceMove) reinforce.Move {
	return reinforce.Move{
		SourceRegionID: req.SourceRegionID,
		TargetRegionID: req.TargetRegionID,
		TroopsInSource: req.TroopsInSource,
		TroopsInTarget: req.TroopsInTarget,
		MovingTroops:   req.MovingTroops,
	}
}

func mapCardsMove(req request.CardsMove) cards.Move {
	combinations := make([]cards.CardCombination, len(req.Combinations))
	for i, combination := range req.Combinations {
		combinations[i] = cards.CardCombination{
			CardIDs: combination.CardIDs,
		}
	}

	return cards.Move{
		Combinations: combinations,
	}
}
