package assignment

import (
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/board"
)

type RegionAssignment map[board.Region]sqlc.Player

type Service interface {
	AssignRegionsToPlayers(players []sqlc.Player, regions []board.Region) RegionAssignment
}

type ServiceImpl struct{}

var _ Service = (*ServiceImpl)(nil)

func NewAssignmentService() *ServiceImpl {
	return &ServiceImpl{}
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