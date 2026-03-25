package board

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/game/db"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/region"
)

type Service interface {
	GetBoardRegions(ctx context.Context) ([]string, error)
	AreNeighbours(ctx context.Context, source string, target string) (bool, error)
	CanPlayerReach(
		ctx ctx.GameContext,
		querier db.Querier,
		source string,
		target string,
	) (bool, error)
	GetContinentsControlledByPlayer(
		ctx ctx.GameContext,
		querier db.Querier,
		playerID int64,
	) ([]*Continent, error)
}

type service struct {
	continents    *ContinentsImpl
	graph         Graph
	regionService region.Service
}

var _ Service = (*service)(nil)

func NewService(regionService region.Service) Service {
	return &service{graph: nil, regionService: regionService}
}

func (s *service) AreNeighbours(
	ctx context.Context,
	source string,
	target string,
) (bool, error) {
	graph, err := s.getGraph(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get graph: %w", err)
	}

	return graph.AreNeighbours(source, target), nil
}

func (s *service) CanPlayerReach(
	ctx ctx.GameContext,
	querier db.Querier,
	source string,
	target string,
) (bool, error) {
	slog.InfoContext(ctx, "checking if player can reach target",
		"source", source, "target", target)

	regions, err := s.regionService.GetRegionsWithQuerier(ctx, querier)
	if err != nil {
		return false, fmt.Errorf("failed to get regions: %w", err)
	}

	usableRegions := make(map[string]struct{})

	for _, region := range regions {
		if region.UserID == ctx.UserID() {
			usableRegions[region.ExternalReference] = struct{}{}
		}
	}

	graph, err := s.getGraph(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get graph: %w", err)
	}

	return graph.CanReach(ctx, source, target, usableRegions), nil
}

func (s *service) GetBoardRegions(ctx context.Context) ([]string, error) {
	slog.InfoContext(ctx, "getting board regions")

	graph, err := s.getGraph(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get board: %w", err)
	}

	result := graph.GetRegions()

	slog.InfoContext(ctx, "got board regions", "regions", result)

	return result, nil
}

func (s *service) getGraph(ctx context.Context) (Graph, error) {
	slog.InfoContext(ctx, "getting graph")

	if s.graph != nil {
		slog.InfoContext(ctx, "graph cache hit")

		return s.graph, nil
	}

	slog.InfoContext(ctx, "graph cache miss, fetching board from file")

	boardDto, err := s.fetchFromFile(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get boardDto: %w", err)
	}

	s.graph, err = NewGraph(boardDto)
	if err != nil {
		return nil, fmt.Errorf("failed to create graph: %w", err)
	}

	slog.InfoContext(ctx, "graph cache updated")

	return s.graph, nil
}

func (s *service) getContinents(ctx ctx.GameContext) (*ContinentsImpl, error) {
	slog.InfoContext(ctx, "getting continents")

	if s.continents != nil {
		slog.InfoContext(ctx, "continents cache hit")

		return s.continents, nil
	}

	slog.InfoContext(ctx, "continents cache miss, fetching board from file")

	boardDto, err := s.fetchFromFile(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get boardDto: %w", err)
	}

	s.continents, err = NewContinents(boardDto)
	if err != nil {
		return nil, fmt.Errorf("failed to create continents: %w", err)
	}

	slog.InfoContext(ctx, "continents cache updated")

	return s.continents, nil
}

func (s *service) fetchFromFile(ctx context.Context) (*BoardDto, error) {
	data, err := os.ReadFile("map.json")
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	board := &BoardDto{}

	err = json.Unmarshal(data, board)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON: %w", err)
	}

	slog.DebugContext(ctx, "Read board from file", "board", board)

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

	continents, err := s.getContinents(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get continents: %w", err)
	}

	regionStrings := make([]string, len(playerRegions))
	for i, region := range playerRegions {
		regionStrings[i] = region.ExternalReference
	}

	return continents.GetContinentsControlledBy(regionStrings), nil
}
