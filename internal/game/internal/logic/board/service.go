package board

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/region"
	domainerrors "github.com/go-risk-it/go-risk-it/internal/kernel/errors"
)

type Service interface {
	GetBoardRegions(ctx context.Context) ([]string, error)
	AreNeighbours(ctx context.Context, source string, target string) (bool, error)
	CanPlayerReachWithRegions(
		ctx ctx.GameContext,
		source string,
		target string,
		regions []snapshot.RegionState,
	) (bool, error)
	GetContinents(ctx ctx.GameContext) (Continents, error)
	GetContinentsControlledByPlayer(
		ctx ctx.GameContext,
		querier db.Querier,
		playerID int64,
	) ([]*Continent, error)
}

type service struct {
	continents    Continents
	graph         Graph
	regionService region.Service
}

var _ Service = (*service)(nil)

func NewService(regionService region.Service) Service {
	return &service{graph: nil, regionService: regionService}
}

// NewServiceWithGraph creates a Service with an injected graph for testing.
func NewServiceWithGraph(regionService region.Service, graph Graph) Service {
	return &service{graph: graph, regionService: regionService}
}

func (s *service) AreNeighbours(
	ctx context.Context,
	source string,
	target string,
) (bool, error) {
	graph, err := s.getGraph()
	if err != nil {
		return false, fmt.Errorf("failed to get graph: %w", err)
	}

	return graph.AreNeighbours(source, target), nil
}

func (s *service) CanPlayerReachWithRegions(
	ctx ctx.GameContext,
	source string,
	target string,
	regions []snapshot.RegionState,
) (bool, error) {
	// Verify the source region exists in the cached regions.
	sourceFound := false

	for _, r := range regions {
		if r.ID == source {
			sourceFound = true

			break
		}
	}

	if !sourceFound {
		return false, domainerrors.NewNotFoundError(
			fmt.Sprintf("source region %s not found in cached regions", source),
		)
	}

	usableRegions := make(map[string]struct{})

	for _, r := range regions {
		if r.OwnerID == ctx.UserID() {
			usableRegions[r.ID] = struct{}{}
		}
	}

	graph, err := s.getGraph()
	if err != nil {
		return false, fmt.Errorf("failed to get graph: %w", err)
	}

	return graph.CanReach(source, target, usableRegions), nil
}

func (s *service) GetBoardRegions(ctx context.Context) ([]string, error) {
	graph, err := s.getGraph()
	if err != nil {
		return nil, fmt.Errorf("failed to get board: %w", err)
	}

	result := graph.GetRegions()

	return result, nil
}

func (s *service) getGraph() (Graph, error) {
	if s.graph != nil {
		return s.graph, nil
	}

	boardDto, err := s.fetchFromFile()
	if err != nil {
		return nil, fmt.Errorf("failed to get boardDto: %w", err)
	}

	s.graph, err = NewGraph(boardDto)
	if err != nil {
		return nil, fmt.Errorf("failed to create graph: %w", err)
	}

	return s.graph, nil
}

func (s *service) GetContinents(ctx ctx.GameContext) (Continents, error) {
	if s.continents != nil {
		return s.continents, nil
	}

	boardDto, err := s.fetchFromFile()
	if err != nil {
		return nil, fmt.Errorf("failed to get boardDto: %w", err)
	}

	s.continents, err = NewContinents(boardDto)
	if err != nil {
		return nil, fmt.Errorf("failed to create continents: %w", err)
	}

	return s.continents, nil
}

func (s *service) fetchFromFile() (*BoardDto, error) {
	data, err := os.ReadFile("map.json")
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	board := &BoardDto{}

	err = json.Unmarshal(data, board)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON: %w", err)
	}

	return board, nil
}

func (s *service) GetContinentsControlledByPlayer(
	ctx ctx.GameContext,
	querier db.Querier,
	playerID int64,
) ([]*Continent, error) {
	playerRegions, err := s.regionService.GetRegionsControlledByPlayer(ctx, querier, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get regions: %w", err)
	}

	continents, err := s.GetContinents(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get continents: %w", err)
	}

	regionStrings := make([]string, len(playerRegions))
	for i, region := range playerRegions {
		regionStrings[i] = region.ExternalReference
	}

	return continents.GetContinentsControlledBy(regionStrings), nil
}
