package attack

import (
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/game/db"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/board"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/attack/dice"
	moveservice "github.com/go-risk-it/go-risk-it/internal/logic/game/move/service"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/phase"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/region"
)

type Move struct {
	AttackingRegionID string `json:"attackingRegionId"`
	DefendingRegionID string `json:"defendingRegionId"`
	TroopsInSource    int64  `json:"troopsInSource"`
	TroopsInTarget    int64  `json:"troopsInTarget"`
	AttackingTroops   int64  `json:"attackingTroops"`
}

type MoveResult struct {
	AttackingRegionID string `json:"attackingRegionId"`
	DefendingRegionID string `json:"defendingRegionId"`
	ConqueringTroops  int64  `json:"conqueringTroops"`
}

type Service interface {
	moveservice.Service[Move, *MoveResult]

	HasConquered(ctx ctx.GameContext, querier db.Querier) (bool, error)
	CanContinueAttacking(ctx ctx.GameContext, querier db.Querier) (bool, error)
}

type service struct {
	boardService  board.Service
	diceService   dice.Service
	phaseService  phase.Service
	regionService region.Service
}

var _ Service = (*service)(nil)

func NewService(
	boardService board.Service,
	diceService dice.Service,
	phaseService phase.Service,
	regionService region.Service,
) (Service, moveservice.Service[Move, *MoveResult]) {
	svc := &service{
		boardService:  boardService,
		diceService:   diceService,
		phaseService:  phaseService,
		regionService: regionService,
	}

	return svc, svc
}

func (s *service) PhaseType() sqlc.GamePhaseType {
	return sqlc.GamePhaseTypeATTACK
}
