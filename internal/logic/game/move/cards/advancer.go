package cards

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
)

func (s *ServiceImpl) AdvanceQ(
	ctx ctx.GameContext,
	querier db.Querier,
	targetPhase sqlc.PhaseType,
	_ *MoveResult,
) error {
	if targetPhase != sqlc.PhaseTypeDEPLOY {
		return fmt.Errorf("cannot advance cards phase to %s", targetPhase)
	}

	return nil
}
