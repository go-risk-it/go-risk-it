package deploy

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
)

func (s *ServiceImpl) Walk(
	ctx ctx.GameContext,
	querier db.Querier,
	_ bool,
) (sqlc.PhaseType, error) {
	deployableTroops, err := s.GetDeployableTroopsQ(ctx, querier)
	if err != nil {
		return sqlc.PhaseTypeDEPLOY, fmt.Errorf("failed to get deployable troops: %w", err)
	}

	if deployableTroops == 0 {
		return sqlc.PhaseTypeATTACK, nil
	}

	return sqlc.PhaseTypeDEPLOY, nil
}
