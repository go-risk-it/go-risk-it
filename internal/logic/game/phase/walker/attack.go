package walker

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/performer/attack"
)

type AttackPhaseWalker interface {
	PhaseWalker
}

type AttackPhaseWalkerImpl struct {
	attackService attack.Service
}

var _ AttackPhaseWalker = (*AttackPhaseWalkerImpl)(nil)

func NewAttackPhaseWalker(attackService attack.Service) *AttackPhaseWalkerImpl {
	return &AttackPhaseWalkerImpl{
		attackService: attackService,
	}
}

func (w *AttackPhaseWalkerImpl) Walk(
	ctx ctx.MoveContext,
	querier db.Querier,
) (sqlc.PhaseType, error) {
	hasConquered, err := w.attackService.HasConqueredQ(ctx, querier)
	if err != nil {
		return sqlc.PhaseTypeATTACK, fmt.Errorf("failed to check if attack has conquered: %w", err)
	}

	if hasConquered {
		ctx.Log().Infow("must advance phase to CONQUER")

		return sqlc.PhaseTypeCONQUER, nil
	}

	canContinueAttacking, err := w.attackService.CanContinueAttackingQ(ctx, querier)
	if err != nil {
		return sqlc.PhaseTypeATTACK, fmt.Errorf("failed to check if attack can continue: %w", err)
	}

	if !canContinueAttacking {
		ctx.Log().Infow("cannot continue attacking, must advance phase to REINFORCE")

		return sqlc.PhaseTypeREINFORCE, nil
	}

	return sqlc.PhaseTypeATTACK, nil
}
