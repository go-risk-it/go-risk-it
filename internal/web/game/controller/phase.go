package controller

import (
	"fmt"
	"log/slog"

	"github.com/go-risk-it/go-risk-it/internal/api/game"
	"github.com/go-risk-it/go-risk-it/internal/api/game/messaging"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/conquer"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/deploy"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/state"
)

type PhaseController struct {
	conquerService conquer.Service
	deployService  deploy.Service
}

func NewPhaseController(
	conquerService conquer.Service,
	deployService deploy.Service,
) *PhaseController {
	return &PhaseController{
		conquerService: conquerService,
		deployService:  deployService,
	}
}

func (c *PhaseController) GetDeployPhaseState(
	ctx ctx.GameContext,
	gameState *state.Game,
) (messaging.GameState[messaging.DeployPhaseState], error) {
	slog.InfoContext(ctx, "fetching deploy phase state")

	deployableTroops, err := c.deployService.GetDeployableTroops(ctx)
	if err != nil {
		return messaging.GameState[messaging.DeployPhaseState]{}, fmt.Errorf(
			"failed to get deployable troops: %w",
			err,
		)
	}

	return messaging.GameState[messaging.DeployPhaseState]{
		ID:   gameState.ID,
		Turn: gameState.Turn,
		Phase: messaging.Phase[messaging.DeployPhaseState]{
			Type: game.Deploy,
			State: messaging.DeployPhaseState{
				DeployableTroops: deployableTroops,
			},
		},
		WinnerUserID: gameState.WinnerUserID,
	}, nil
}

func (c *PhaseController) GetAttackPhaseState(
	ctx ctx.GameContext,
	gameState *state.Game,
) (messaging.GameState[messaging.EmptyState], error) {
	return c.getEmptyPhaseState(ctx, gameState, game.Attack), nil
}

func (c *PhaseController) GetConquerPhaseState(
	ctx ctx.GameContext,
	gameState *state.Game,
) (messaging.GameState[messaging.ConquerPhaseState], error) {
	slog.InfoContext(ctx, "fetching conquer phase state")

	conquerPhase, err := c.conquerService.GetPhaseState(ctx)
	if err != nil {
		return messaging.GameState[messaging.ConquerPhaseState]{}, fmt.Errorf(
			"failed to get conquer phase state: %w",
			err,
		)
	}

	return messaging.GameState[messaging.ConquerPhaseState]{
		ID:   gameState.ID,
		Turn: gameState.Turn,
		Phase: messaging.Phase[messaging.ConquerPhaseState]{
			Type: game.Conquer,
			State: messaging.ConquerPhaseState{
				AttackingRegionID: conquerPhase.SourceRegion,
				DefendingRegionID: conquerPhase.TargetRegion,
				MinTroopsToMove:   conquerPhase.MinimumTroops,
			},
		},
		WinnerUserID: gameState.WinnerUserID,
	}, nil
}

func (c *PhaseController) GetReinforcePhaseState(
	ctx ctx.GameContext,
	gameState *state.Game,
) (messaging.GameState[messaging.EmptyState], error) {
	return c.getEmptyPhaseState(ctx, gameState, game.Reinforce), nil
}

func (c *PhaseController) GetCardsPhaseState(
	ctx ctx.GameContext,
	gameState *state.Game,
) (messaging.GameState[messaging.EmptyState], error) {
	return c.getEmptyPhaseState(ctx, gameState, game.Cards), nil
}

func (c *PhaseController) getEmptyPhaseState(
	ctx ctx.GameContext,
	game *state.Game,
	phaseType game.PhaseType,
) messaging.GameState[messaging.EmptyState] {
	slog.InfoContext(ctx, "fetching phase state", "phaseType", phaseType)

	return messaging.GameState[messaging.EmptyState]{
		ID:   game.ID,
		Turn: game.Turn,
		Phase: messaging.Phase[messaging.EmptyState]{
			Type:  phaseType,
			State: messaging.EmptyState{},
		},
		WinnerUserID: game.WinnerUserID,
	}
}
