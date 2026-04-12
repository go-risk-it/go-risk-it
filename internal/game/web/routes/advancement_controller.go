package routes

import (
	"fmt"

	game "github.com/go-risk-it/go-risk-it/internal/game/api"
	"github.com/go-risk-it/go-risk-it/internal/game/api/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/orchestration"
	domainerrors "github.com/go-risk-it/go-risk-it/internal/kernel/errors"
)

type AdvancementController struct {
	attackAdvancer    orchestration.AttackPhaseAdvancer
	cardsAdvancer     orchestration.CardsPhaseAdvancer
	reinforceAdvancer orchestration.ReinforcePhaseAdvancer
}

func NewAdvancementController(
	attackAdvancer orchestration.AttackPhaseAdvancer,
	cardsAdvancer orchestration.CardsPhaseAdvancer,
	reinforceAdvancer orchestration.ReinforcePhaseAdvancer,
) *AdvancementController {
	return &AdvancementController{
		attackAdvancer:    attackAdvancer,
		cardsAdvancer:     cardsAdvancer,
		reinforceAdvancer: reinforceAdvancer,
	}
}

func (c *AdvancementController) Advance(
	ctx ctx.GameContext,
	advancement request.Advancement,
) error {
	var err error

	switch advancement.CurrentPhase {
	case game.Deploy:
		err = domainerrors.NewValidationError("cannot advance from deploy phase")
	case game.Attack:
		err = c.attackAdvancer.AdvancePhase(ctx)
	case game.Conquer:
		err = domainerrors.NewValidationError("cannot advance from conquer phase")
	case game.Reinforce:
		err = c.reinforceAdvancer.AdvancePhase(ctx)
	case game.Cards:
		err = c.cardsAdvancer.AdvancePhase(ctx)
	default:
		err = fmt.Errorf("invalid phase type: %s", advancement.CurrentPhase)
	}

	if err != nil {
		return fmt.Errorf("unable to advance phase: %w", err)
	}

	return nil
}
