package assignment

import (
	"github.com/go-risk-it/go-risk-it/internal/config"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/rand"
)

type RegionAssignment map[string]sqlc.GamePlayer

type Assigner interface {
	AssignRegionsToPlayers(players []sqlc.GamePlayer, regions []string) RegionAssignment
}

type Service interface {
	AssignRegionsToPlayers(players []sqlc.GamePlayer, regions []string) RegionAssignment
}

type ServiceImpl struct {
	assigner Assigner
}

var _ Service = (*ServiceImpl)(nil)

func NewAssignmentService(assignConfig config.RegionassignmentConfig, rng rand.RNG) *ServiceImpl {
	assigner := getAssigner(assignConfig, rng)

	return &ServiceImpl{
		assigner: assigner,
	}
}

func getAssigner(assignConfig config.RegionassignmentConfig, rng rand.RNG) Assigner {
	switch assignConfig.AssignmentStrategy {
	case "sequential":
		return NewSequential()
	case "random":
		return NewRandom(rng)
	default:
		panic("unknown assignment strategy: " + assignConfig.AssignmentStrategy)
	}
}

func (s *ServiceImpl) AssignRegionsToPlayers(
	players []sqlc.GamePlayer,
	regions []string,
) RegionAssignment {
	return s.assigner.AssignRegionsToPlayers(players, regions)
}
