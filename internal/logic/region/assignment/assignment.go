package assignment

import (
	"github.com/go-risk-it/go-risk-it/internal/data/db"
	sqlc "github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/board"
	"go.uber.org/zap"
)

type RegionAssignment map[board.Region]sqlc.Player

type Service interface {
	AssignRegionsToPlayers(players []sqlc.Player, regions []board.Region) RegionAssignment
}

type ServiceImpl struct {
	q   db.Querier
	log *zap.SugaredLogger
}

func NewAssignmentService(querier db.Querier, logger *zap.SugaredLogger) *ServiceImpl {
	return &ServiceImpl{q: querier, log: logger}
}

func (s *ServiceImpl) AssignRegionsToPlayers(
	players []sqlc.Player,
	regions []board.Region,
) RegionAssignment {
	regionsToPlayers := make(map[board.Region]sqlc.Player)
	for i, region := range regions {
		regionsToPlayers[region] = players[i%len(players)]
	}

	return regionsToPlayers
}
