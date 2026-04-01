package routes

import (
	"fmt"

	game "github.com/go-risk-it/go-risk-it/internal/game/api"
	"github.com/go-risk-it/go-risk-it/internal/game/api/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/advancement"
	domainerrors "github.com/go-risk-it/go-risk-it/internal/kernel/errors"
)

type AdvancementController struct {
	attackAdvancer    advancement.AttackAdvancer
	cardsAdvancer     advancement.CardsAdvancer
	reinforceAdvancer advancement.ReinforceAdvancer
}

func NewAdvancementController(
	attackAdvancer advancement.AttackAdvancer,
	cardsAdvancer advancement.CardsAdvancer,
	reinforceAdvancer advancement.ReinforceAdvancer,
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
		err = c.attackAdvancer.Advance(ctx)
	case game.Conquer:
		err = domainerrors.NewValidationError("cannot advance from conquer phase")
	case game.Reinforce:
		err = c.reinforceAdvancer.Advance(ctx)
	case game.Cards:
		err = c.cardsAdvancer.Advance(ctx)
	default:
		err = fmt.Errorf("invalid phase type: %s", advancement.CurrentPhase)
	}

	if err != nil {
		return fmt.Errorf("unable to advance phase: %w", err)
	}

	return nil
}
