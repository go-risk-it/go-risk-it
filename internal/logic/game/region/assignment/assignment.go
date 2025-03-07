package assignment

import (
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
)

type RegionAssignment map[string]sqlc.GamePlayer

type Service interface {
	AssignRegionsToPlayers(players []sqlc.GamePlayer, regions []string) RegionAssignment
}

type ServiceImpl struct{}

var _ Service = (*ServiceImpl)(nil)

func NewAssignmentService() *ServiceImpl {
	return &ServiceImpl{}
}

func (s *ServiceImpl) AssignRegionsToPlayers(
	players []sqlc.GamePlayer,
	regions []string,
) RegionAssignment {
	regionsToPlayers := make(map[string]sqlc.GamePlayer)
	for i, region := range regions {
		regionsToPlayers[region] = players[i%len(players)]
	}

	return regionsToPlayers
}
