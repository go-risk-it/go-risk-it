package reinforce

import (
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/board"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/cards"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/service"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/phase"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/region"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/state"
)

type Move struct {
	SourceRegionID string `json:"sourceRegionId"`
	TargetRegionID string `json:"targetRegionId"`
	TroopsInSource int64  `json:"troopsInSource"`
	TroopsInTarget int64  `json:"troopsInTarget"`
	MovingTroops   int64  `json:"movingTroops"`
}

type MoveResult struct{}

type Service interface {
	service.Service[Move, *MoveResult]
}

type ServiceImpl struct {
	boardService  board.Service
	cardsService  cards.Service
	gameService   state.Service
	phaseService  phase.Service
	regionService region.Service
}

var _ Service = (*ServiceImpl)(nil)

func NewService(
	boardService board.Service,
	cardsService cards.Service,
	gameService state.Service,
	phaseService phase.Service,
	regionService region.Service,
) *ServiceImpl {
	return &ServiceImpl{
		boardService:  boardService,
		cardsService:  cardsService,
		gameService:   gameService,
		phaseService:  phaseService,
		regionService: regionService,
	}
}

func (s *ServiceImpl) PhaseType() sqlc.GamePhaseType {
	return sqlc.GamePhaseTypeREINFORCE
}
