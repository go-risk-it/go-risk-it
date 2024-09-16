package attack

import (
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/board"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/attack/dice"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/service"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/phase"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/region"
)

type Move struct {
	AttackingRegionID string
	DefendingRegionID string
	TroopsInSource    int64
	TroopsInTarget    int64
	AttackingTroops   int64
}

type MoveResult struct {
	AttackingRegionID string
	DefendingRegionID string
	ConqueringTroops  int64
}

type Service interface {
	service.Service[Move, *MoveResult]

	HasConqueredQ(ctx ctx.GameContext, querier db.Querier) (bool, error)
	CanContinueAttackingQ(ctx ctx.GameContext, querier db.Querier) (bool, error)
}

type ServiceImpl struct {
	boardService  board.Service
	diceService   dice.Service
	phaseService  phase.Service
	regionService region.Service
}

var _ Service = &ServiceImpl{}

func NewService(
	boardService board.Service,
	diceService dice.Service,
	phaseService phase.Service,
	regionService region.Service,
) *ServiceImpl {
	return &ServiceImpl{
		boardService:  boardService,
		diceService:   diceService,
		phaseService:  phaseService,
		regionService: regionService,
	}
}
