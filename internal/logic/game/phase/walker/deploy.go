package walker

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/performer/deploy"
)

type DeployPhaseWalker interface {
	PhaseWalker
}

type DeployPhaseWalkerImpl struct {
	deployService deploy.Service
}

var _ DeployPhaseWalker = (*DeployPhaseWalkerImpl)(nil)

func NewDeployPhaseWalker(deployService deploy.Service) *DeployPhaseWalkerImpl {
	return &DeployPhaseWalkerImpl{
		deployService: deployService,
	}
}

func (w *DeployPhaseWalkerImpl) Walk(
	ctx ctx.MoveContext,
	querier db.Querier,
) (sqlc.PhaseType, error) {
	deployableTroops, err := w.deployService.GetDeployableTroops(ctx, querier)
	if err != nil {
		return sqlc.PhaseTypeDEPLOY, fmt.Errorf("failed to get deployable troops: %w", err)
	}

	if deployableTroops == 0 {
		return sqlc.PhaseTypeATTACK, nil
	}

	return sqlc.PhaseTypeDEPLOY, nil
}
