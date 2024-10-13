package cards

import (
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
)

func (s *ServiceImpl) Walk(ctx ctx.GameContext, querier db.Querier) (sqlc.PhaseType, error) {
	panic("implement me")
}
