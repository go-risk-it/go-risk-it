package reinforce

import (
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/board"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/cards"
	moveservice "github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/service"
)

type Move struct {
	SourceRegionID string `json:"sourceRegionId"`
	TargetRegionID string `json:"targetRegionId"`
	TroopsInSource int64  `json:"troopsInSource"`
	TroopsInTarget int64  `json:"troopsInTarget"`
	MovingTroops   int64  `json:"movingTroops"`
}

type Service interface {
	moveservice.Service[Move, struct{}]
}

type service struct {
	boardService board.Service
	cardsService cards.Service
}

var _ Service = (*service)(nil)

func NewService(
	boardService board.Service,
	cardsService cards.Service,
) (Service, moveservice.Service[Move, struct{}]) {
	svc := &service{
		boardService: boardService,
		cardsService: cardsService,
	}

	return svc, moveservice.NewTracedService[Move, struct{}](svc)
}

func (s *service) PhaseType() sqlc.GamePhaseType {
	return sqlc.GamePhaseTypeREINFORCE
}
