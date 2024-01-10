package assignment

import (
	"github.com/tomfran/go-risk-it/internal/db"
	"github.com/tomfran/go-risk-it/internal/game/board"
	"go.uber.org/zap"
)

type RegionAssignment map[board.Region]db.Player

type Service struct {
	q   *db.Queries
	log *zap.SugaredLogger
}

func NewAssignmentService(queries *db.Queries, logger *zap.SugaredLogger) *Service {
	return &Service{q: queries, log: logger}
}

func (s *Service) AssignRegionsToPlayers(players []db.Player, regions []board.Region) RegionAssignment {
	regionsToPlayers := make(map[board.Region]db.Player)
	for i, region := range regions {
		regionsToPlayers[region] = players[i%len(players)]
	}
	return regionsToPlayers
}
