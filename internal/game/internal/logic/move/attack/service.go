package attack

import (
	attackapi "github.com/go-risk-it/go-risk-it/internal/game/api/moves/attack"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/board"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/attack/dice"
	moveservice "github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/service"
)

type Move struct {
	AttackingRegionID string `json:"attackingRegionId"`
	DefendingRegionID string `json:"defendingRegionId"`
	TroopsInSource    int64  `json:"troopsInSource"`
	TroopsInTarget    int64  `json:"troopsInTarget"`
	AttackingTroops   int64  `json:"attackingTroops"`
}

// MoveResult is a type alias to the canonical definition in game/api/moves/attack.
// This preserves backward compatibility for all existing consumers.
type MoveResult = attackapi.MoveResult

type Service interface {
	moveservice.Service[Move, *MoveResult]
}

type service struct {
	boardService board.Service
	diceService  dice.Service
}

var _ Service = (*service)(nil)

func NewService(
	boardService board.Service,
	diceService dice.Service,
) (Service, moveservice.Service[Move, *MoveResult]) {
	svc := &service{
		boardService: boardService,
		diceService:  diceService,
	}

	return svc, moveservice.NewTracedService[Move, *MoveResult](svc)
}

func (s *service) PhaseType() sqlc.GamePhaseType {
	return sqlc.GamePhaseTypeATTACK
}
