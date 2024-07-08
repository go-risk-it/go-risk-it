package controller

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/api/game/messaging"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/performer/deploy"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/state"
)

type PhaseController interface {
	GetDeployPhaseState(
		ctx ctx.GameContext,
		game *state.Game,
	) (messaging.GameState[messaging.DeployPhaseState], error)
}

type PhaseControllerImpl struct {
	deployService deploy.Service
}

var _ PhaseController = (*PhaseControllerImpl)(nil)

func NewPhaseController(
	deployService deploy.Service,
) *PhaseControllerImpl {
	return &PhaseControllerImpl{
		deployService: deployService,
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
