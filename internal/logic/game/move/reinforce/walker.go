package reinforce

import (
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
)

func (s *ServiceImpl) Walk(ctx.GameContext, db.Querier) (sqlc.PhaseType, error) {
	return sqlc.PhaseTypeCARDS, nil
}
