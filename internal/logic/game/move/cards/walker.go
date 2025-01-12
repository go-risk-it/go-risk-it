package cards

import (
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
)

func (s *ServiceImpl) WalkQ(_ ctx.GameContext, _ db.Querier, _ bool) (sqlc.PhaseType, error) {
	return sqlc.PhaseTypeDEPLOY, nil
}
