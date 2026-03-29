package cards

import (
	"github.com/go-risk-it/go-risk-it/internal/game/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
)

func (s *service) Walk(_ ctx.GameContext, _ db.Querier, _ bool) (sqlc.GamePhaseType, error) {
	return sqlc.GamePhaseTypeDEPLOY, nil
}
