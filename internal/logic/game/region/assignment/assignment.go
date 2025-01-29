package assignment

import (
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
)

type RegionAssignment map[string]sqlc.Player

type Service interface {
	AssignRegionsToPlayers(players []sqlc.Player, regions []string) RegionAssignment
}

type ServiceImpl struct{}

var _ Service = (*ServiceImpl)(nil)

func NewAssignmentService() *ServiceImpl {
	return &ServiceImpl{}
}

func (s *ServiceImpl) AssignRegionsToPlayers(
	players []sqlc.Player,
	regions []string,
) RegionAssignment {
	regionsToPlayers := make(map[string]sqlc.Player)
	for i, region := range regions {
		regionsToPlayers[region] = players[i%len(players)]
	}

	return regionsToPlayers
}
