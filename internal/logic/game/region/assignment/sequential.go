package assignment

import "github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"

type Sequential struct{}

var _ Assigner = (*Sequential)(nil)

func NewSequential() Assigner {
	return &Sequential{}
}

func (s *Sequential) AssignRegionsToPlayers(
	players []sqlc.GamePlayer,
	regions []string,
) RegionAssignment {
	regionsToPlayers := make(map[string]sqlc.GamePlayer)
	for i, region := range regions {
		regionsToPlayers[region] = players[i%len(players)]
	}

	return regionsToPlayers
}
