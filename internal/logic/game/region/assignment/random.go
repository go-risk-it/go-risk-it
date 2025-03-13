package assignment

import (
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/rand"
)

type Random struct {
	rng rand.RNG
}

var _ Assigner = (*Random)(nil)

func NewRandom(src rand.RNG) Assigner {
	return &Random{rng: src}
}

func (s *Random) AssignRegionsToPlayers(
	players []sqlc.GamePlayer,
	regions []string,
) RegionAssignment {
	regionsToPlayers := make(map[string]sqlc.GamePlayer)

	s.rng.Shuffle(len(regions), func(i, j int) {
		regions[i], regions[j] = regions[j], regions[i]
	})

	s.rng.Shuffle(len(players), func(i, j int) {
		players[i], players[j] = players[j], players[i]
	})

	for i, region := range regions {
		regionsToPlayers[region] = players[i%len(players)]
	}

	return regionsToPlayers
}
