package attack

import (
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/board"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/attack/dice"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/service"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/phase"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/player"
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
	service.Service[Move, *MoveResult]

	HasConqueredQ(ctx ctx.GameContext, querier db.Querier) (bool, error)
	CanContinueAttackingQ(ctx ctx.GameContext, querier db.Querier) (bool, error)
}

type ServiceImpl struct {
	boardService  board.Service
	diceService   dice.Service
	phaseService  phase.Service
	playerService player.Service
	regionService region.Service
}

var _ Service = &ServiceImpl{}

func NewService(
	boardService board.Service,
	diceService dice.Service,
	phaseService phase.Service,
	playerService player.Service,
	regionService region.Service,
) *ServiceImpl {
	return &ServiceImpl{
		boardService:  boardService,
		diceService:   diceService,
		phaseService:  phaseService,
		playerService: playerService,
		regionService: regionService,
	}
}

func (s *ServiceImpl) PhaseType() sqlc.PhaseType {
	return sqlc.PhaseTypeATTACK
}
