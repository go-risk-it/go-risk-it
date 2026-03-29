package cards

import (
	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
)

func (s *service) Walk(_ ctx.GameContext, _ db.Querier, _ bool) (sqlc.GamePhaseType, error) {
	return sqlc.GamePhaseTypeDEPLOY, nil
}
