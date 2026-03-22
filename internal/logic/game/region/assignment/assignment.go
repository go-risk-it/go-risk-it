package assignment

import (
	"fmt"

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

func NewAssignmentService(
	assignConfig config.RegionassignmentConfig,
	rng rand.RNG,
) (*ServiceImpl, error) {
	assigner, err := getAssigner(assignConfig, rng)
	if err != nil {
		return nil, err
	}

	return &ServiceImpl{
		assigner: assigner,
	}, nil
}

func getAssigner(assignConfig config.RegionassignmentConfig, rng rand.RNG) (Assigner, error) {
	switch assignConfig.AssignmentStrategy {
	case "sequential":
		return NewSequential(), nil
	case "random":
		return NewRandom(rng), nil
	default:
		return nil, fmt.Errorf("unknown assignment strategy: %s", assignConfig.AssignmentStrategy)
	}
}

func (s *ServiceImpl) AssignRegionsToPlayers(
	players []sqlc.GamePlayer,
	regions []string,
) RegionAssignment {
	return s.assigner.AssignRegionsToPlayers(players, regions)
}
