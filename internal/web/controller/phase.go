package controller

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/api/game/message"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/performer/deploy"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/state"
)

type PhaseController interface {
	GetDeployPhaseState(
		ctx ctx.GameContext,
		game *state.Game,
	) (message.GameState[message.DeployPhaseState], error)
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
) (message.GameState[message.DeployPhaseState], error) {
	ctx.Log().Infow("fetching deploy phase state")

	deployableTroops, err := c.deployService.GetDeployableTroops(ctx)
	if err != nil {
		return message.GameState[message.DeployPhaseState]{}, fmt.Errorf(
			"failed to get deployable troops: %w",
			err,
		)
	}

	return message.GameState[message.DeployPhaseState]{
		ID:   game.ID,
		Turn: game.Turn,
		Phase: message.Phase[message.DeployPhaseState]{
			Type: message.Deploy,
			State: message.DeployPhaseState{
				DeployableTroops: deployableTroops,
			},
		},
	}, nil
}
