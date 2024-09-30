package reinforce

import (
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/board"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/service"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/phase"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/region"
)

type Move struct {
	SourceRegionID string
	TargetRegionID string
	TroopsInSource int64
	TroopsInTarget int64
	MovingTroops   int64
}

type MoveResult struct{}

type Service interface {
	service.Service[Move, *MoveResult]
}

type ServiceImpl struct {
	boardService  board.Service
	phaseService  phase.Service
	regionService region.Service
}

var _ Service = (*ServiceImpl)(nil)

func NewService(
	boardService board.Service,
	phaseService phase.Service,
	regionService region.Service,
) *ServiceImpl {
	return &ServiceImpl{
		boardService:  boardService,
		phaseService:  phaseService,
		regionService: regionService,
	}
}

func (s *ServiceImpl) PhaseType() sqlc.PhaseType {
	return sqlc.PhaseTypeREINFORCE
}

func (s *ServiceImpl) ForcedAdvancementPhase() sqlc.PhaseType {
	return sqlc.PhaseTypeREINFORCE
}
