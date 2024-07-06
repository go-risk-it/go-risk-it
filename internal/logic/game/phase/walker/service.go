package walker

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
)

type Service interface {
	WalkToTargetPhase(
		ctx ctx.MoveContext,
		querier db.Querier,
		currentPhase sqlc.PhaseType,
	) (sqlc.PhaseType, error)
}

type ServiceImpl struct {
	deployPhaseWalker DeployPhaseWalker
	attackPhaseWalker AttackPhaseWalker
}

var _ Service = (*ServiceImpl)(nil)

func NewService(
	deployPhaseWalker DeployPhaseWalker,
	attackPhaseWalker AttackPhaseWalker,
) *ServiceImpl {
	return &ServiceImpl{
		deployPhaseWalker: deployPhaseWalker,
		attackPhaseWalker: attackPhaseWalker,
	}
}

func (w *ServiceImpl) WalkToTargetPhase(
	ctx ctx.MoveContext,
	querier db.Querier,
	currentPhase sqlc.PhaseType,
) (sqlc.PhaseType, error) {
	ctx.Log().Infow("walking", "from", currentPhase)

	walker, err := w.getWalker(currentPhase)
	if err != nil {
		return "", fmt.Errorf("failed to get phase walker: %w", err)
	}

	targetPhase, err := walker.Walk(ctx, querier)
	if err != nil {
		return "", fmt.Errorf("failed to walk to target phase: %w", err)
	}

	if targetPhase == currentPhase {
		ctx.Log().Infow("no need to walk further")

		return targetPhase, nil
	}

	ctx.Log().Infow(
		"walking further",
		"from",
		currentPhase,
		"to",
		targetPhase,
	)

	return w.WalkToTargetPhase(ctx, querier, targetPhase)
}

func (w *ServiceImpl) getWalker(phaseType sqlc.PhaseType) (PhaseWalker, error) {
	switch phaseType {
	case sqlc.PhaseTypeDEPLOY:
		return w.deployPhaseWalker, nil
	case sqlc.PhaseTypeATTACK:
		return w.attackPhaseWalker, nil
	default:
		return nil, fmt.Errorf("unknown phase type: %s", phaseType)
	}
}
