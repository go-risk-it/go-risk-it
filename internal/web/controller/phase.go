package controller

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/api/game/messaging"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/conquer"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/deploy"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/state"
)

type PhaseController interface {
	GetDeployPhaseState(
		ctx ctx.GameContext,
		game *state.Game,
	) (messaging.GameState[messaging.DeployPhaseState], error)
	GetAttackPhaseState(
		ctx ctx.GameContext,
		game *state.Game,
	) (messaging.GameState[messaging.EmptyState], error)
	GetConquerPhaseState(
		ctx ctx.GameContext,
		game *state.Game,
	) (messaging.GameState[messaging.ConquerPhaseState], error)
}

type PhaseControllerImpl struct {
	conquerService conquer.Service
	deployService  deploy.Service
}

var _ PhaseController = (*PhaseControllerImpl)(nil)

func NewPhaseController(
	conquerService conquer.Service,
	deployService deploy.Service,
) *PhaseControllerImpl {
	return &PhaseControllerImpl{
		conquerService: conquerService,
		deployService:  deployService,
	}
}

func (c *PhaseControllerImpl) GetDeployPhaseState(
	ctx ctx.GameContext,
	game *state.Game,
) (messaging.GameState[messaging.DeployPhaseState], error) {
	ctx.Log().Infow("fetching deploy phase state")

	deployableTroops, err := c.deployService.GetDeployableTroops(ctx)
	if err != nil {
		return messaging.GameState[messaging.DeployPhaseState]{}, fmt.Errorf(
			"failed to get deployable troops: %w",
			err,
		)
	}

	return messaging.GameState[messaging.DeployPhaseState]{
		ID:   game.ID,
		Turn: game.Turn,
		Phase: messaging.Phase[messaging.DeployPhaseState]{
			Type: messaging.Deploy,
			State: messaging.DeployPhaseState{
				DeployableTroops: deployableTroops,
			},
		},
	}, nil
}

func (c *PhaseControllerImpl) GetAttackPhaseState(
	ctx ctx.GameContext,
	game *state.Game,
) (messaging.GameState[messaging.EmptyState], error) {
	ctx.Log().Infow("fetching attack phase phaseState")

	return c.getEmptyPhaseState(ctx, game, messaging.Attack), nil
}

func (c *PhaseControllerImpl) GetConquerPhaseState(
	ctx ctx.GameContext,
	game *state.Game,
) (messaging.GameState[messaging.ConquerPhaseState], error) {
	ctx.Log().Infow("fetching conquer phase phaseState")

	conquerPhase, err := c.conquerService.GetPhaseState(ctx)
	if err != nil {
		return messaging.GameState[messaging.ConquerPhaseState]{}, fmt.Errorf(
			"failed to get conquer phase state: %w",
			err,
		)
	}

	return messaging.GameState[messaging.ConquerPhaseState]{
		ID:   game.ID,
		Turn: game.Turn,
		Phase: messaging.Phase[messaging.ConquerPhaseState]{
			Type: messaging.Conquer,
			State: messaging.ConquerPhaseState{
				AttackingRegionID: conquerPhase.SourceRegion,
				DefendingRegionID: conquerPhase.TargetRegion,
				MinTroopsToMove:   conquerPhase.MinimumTroops,
			},
		},
	}, nil
}

func (c *PhaseControllerImpl) getEmptyPhaseState(
	ctx ctx.GameContext,
	game *state.Game,
	phaseType messaging.PhaseType,
) messaging.GameState[messaging.EmptyState] {
	ctx.Log().Infow("fetching empty phase state")

	return messaging.GameState[messaging.EmptyState]{
		ID:   game.ID,
		Turn: game.Turn,
		Phase: messaging.Phase[messaging.EmptyState]{
			Type:  phaseType,
			State: messaging.EmptyState{},
		},
	}
}
