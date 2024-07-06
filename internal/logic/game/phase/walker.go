package phase

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/performer/attack"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/performer/deploy"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/state"
)

type Walker interface {
	WalkToTargetPhase(
		ctx ctx.MoveContext,
		querier db.Querier,
		gameState *state.Game,
	) (sqlc.PhaseType, error)
}

type WalkerImpl struct {
	attackService attack.Service
	deployService deploy.Service
}

var _ Walker = (*WalkerImpl)(nil)

func NewWalker(
	attackService attack.Service,
	deployService deploy.Service,
) *WalkerImpl {
	return &WalkerImpl{
		attackService: attackService,
		deployService: deployService,
	}
}

func (w *WalkerImpl) WalkToTargetPhase(
	ctx ctx.MoveContext,
	querier db.Querier,
	gameState *state.Game,
) (sqlc.PhaseType, error) {
	targetPhase := gameState.CurrentPhase

	mustAdvance := true
	for mustAdvance {
		mustAdvance = false

		switch targetPhase {
		case sqlc.PhaseTypeDEPLOY:
			deployableTroops, err := w.deployService.GetDeployableTroops(ctx, querier)
			if err != nil {
				return targetPhase, fmt.Errorf("failed to get deployable troops: %w", err)
			}

			if deployableTroops == 0 {
				ctx.Log().Infow(
					"deploy must advance",
					"phase",
					gameState.CurrentPhase,
				)

				targetPhase = sqlc.PhaseTypeATTACK
				mustAdvance = true
			}
		case sqlc.PhaseTypeATTACK:
			targetPhase, err := w.getTargetPhaseForAttack(ctx, querier)
			if err != nil {
				return targetPhase, fmt.Errorf("failed to get target phase for attack: %w", err)
			}

			if targetPhase != sqlc.PhaseTypeATTACK {
				mustAdvance = true
			}
		case sqlc.PhaseTypeCONQUER:
		case sqlc.PhaseTypeREINFORCE:
		case sqlc.PhaseTypeCARDS:
		}
	}

	return targetPhase, nil
}

func (w *WalkerImpl) getTargetPhaseForAttack(
	ctx ctx.MoveContext,
	querier db.Querier,
) (sqlc.PhaseType, error) {
	targetPhase := sqlc.PhaseTypeATTACK

	hasConquered, err := w.attackService.HasConqueredQ(ctx, querier)
	if err != nil {
		return targetPhase, fmt.Errorf("failed to check if attack has conquered: %w", err)
	}

	if hasConquered {
		ctx.Log().Infow("must advance phase to CONQUER")

		return sqlc.PhaseTypeCONQUER, nil
	}

	canContinueAttacking, err := w.attackService.CanContinueAttackingQ(ctx, querier)
	if err != nil {
		return targetPhase, fmt.Errorf("failed to check if attack can continue: %w", err)
	}

	if !canContinueAttacking {
		ctx.Log().Infow("cannot continue attacking, must advance phase to REINFORCE")

		return sqlc.PhaseTypeREINFORCE, nil
	}

	return targetPhase, nil
}
