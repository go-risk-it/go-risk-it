package assignment

import (
	"github.com/tomfran/go-risk-it/internal/db"
	"github.com/tomfran/go-risk-it/internal/logic/board"
	"go.uber.org/zap"
)

type RegionAssignment map[board.Region]db.Player

type Service interface {
	AssignRegionsToPlayers(players []db.Player, regions []board.Region) RegionAssignment
}

type ServiceImpl struct {
	q   *db.Queries
	log *zap.SugaredLogger
}

func NewAssignmentService(queries *db.Queries, logger *zap.SugaredLogger) *ServiceImpl {
	return &ServiceImpl{q: queries, log: logger}
}

func (s *ServiceImpl) AssignRegionsToPlayers(
	players []db.Player,
	regions []board.Region,
) RegionAssignment {
	regionsToPlayers := make(map[board.Region]db.Player)
	for i, region := range regions {
		regionsToPlayers[region] = players[i%len(players)]
	}

	return regionsToPlayers
}
