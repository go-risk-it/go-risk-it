package cards

import (
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
	moveservice "github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/service"
)

func (s *service) Walk(_ moveservice.WalkContext) (sqlc.GamePhaseType, error) {
	return sqlc.GamePhaseTypeDEPLOY, nil
}
