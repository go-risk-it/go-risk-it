package controller

import (
	"errors"
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/api/game"
	"github.com/go-risk-it/go-risk-it/internal/api/game/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/advancement"
)

type AdvancementController interface {
	Advance(ctx ctx.GameContext, advancement request.Advancement) error
}

type AdvancementControllerImpl struct {
	attackAdvancer    advancement.AttackAdvancer
	cardsAdvancer     advancement.CardsAdvancer
	reinforceAdvancer advancement.ReinforceAdvancer
}

var _ AdvancementController = (*AdvancementControllerImpl)(nil)

func NewAdvancementController(
	attackAdvancer advancement.AttackAdvancer,
	cardsAdvancer advancement.CardsAdvancer,
	reinforceAdvancer advancement.ReinforceAdvancer,
) *AdvancementControllerImpl {
	return &AdvancementControllerImpl{
		attackAdvancer:    attackAdvancer,
		cardsAdvancer:     cardsAdvancer,
		reinforceAdvancer: reinforceAdvancer,
	}
}

func (c *AdvancementControllerImpl) Advance(
	ctx ctx.GameContext,
	advancement request.Advancement,
) error {
	var err error

	switch advancement.CurrentPhase {
	case game.Deploy:
		err = errors.New("cannot advance from deploy phase")
	case game.Attack:
		err = c.attackAdvancer.Advance(ctx)
	case game.Conquer:
		err = errors.New("cannot advance from conquer phase")
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
