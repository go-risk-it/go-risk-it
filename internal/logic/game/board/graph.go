package board

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"slices"

	domainerrors "github.com/go-risk-it/go-risk-it/internal/logic/errors"
)

type Graph interface {
	GetRegions() []string
	AreNeighbours(source string, target string) bool
	CanReach(
		ctx context.Context,
		source string,
		target string,
		usableRegions map[string]struct{},
	) bool
}

type graphImpl struct {
	Edges map[string]map[string]struct{}
}

var _ Graph = (*graphImpl)(nil)

func (g *graphImpl) GetRegions() []string {
	return slices.Sorted(maps.Keys(g.Edges))
}

func NewGraph(board *BoardDto) (Graph, error) {
	err := validateGraph(board)
	if err != nil {
		return nil, fmt.Errorf("invalid board: %w", err)
	}

	topG := &graphImpl{
		Edges: make(map[string]map[string]struct{}),
	}

	for _, region := range board.Regions {
		topG.Edges[region.ExternalReference] = make(map[string]struct{})
	}

	for _, border := range board.Borders {
		topG.Edges[border.Source][border.Target] = struct{}{}
		topG.Edges[border.Target][border.Source] = struct{}{}
	}

	return topG, nil
}

func validateGraph(board *BoardDto) error {
	if len(board.Regions) == 0 {
		return domainerrors.NewValidationError("no regions")
	}

	if len(board.Borders) == 0 {
		return domainerrors.NewValidationError("no borders")
	}

	regionNames := make(map[string]struct{})
	for _, region := range board.Regions {
		if _, ok := regionNames[region.ExternalReference]; ok {
			return domainerrors.NewValidationError("duplicate region")
		}

		regionNames[region.ExternalReference] = struct{}{}
	}

	err := validateBorders(board, regionNames)
	if err != nil {
		return fmt.Errorf("invalid borders: %w", err)
	}

	return nil
}

func validateBorders(board *BoardDto, regionNames map[string]struct{}) error {
	for _, border := range board.Borders {
		if border.Source == "" {
			return domainerrors.NewValidationError("empty source")
		}

		if border.Target == "" {
			return domainerrors.NewValidationError("empty target")
		}

		if border.Source == border.Target {
			return domainerrors.NewValidationError("self-loop")
		}

		if _, ok := regionNames[border.Source]; !ok {
			return domainerrors.NewValidationErrorf("unknown source %v", border.Source)
		}

		if _, ok := regionNames[border.Target]; !ok {
			return domainerrors.NewValidationErrorf("unknown target %v", border.Target)
		}
	}

	return nil
}

func (g *graphImpl) AreNeighbours(source string, target string) bool {
	_, ok := g.Edges[source][target]

	return ok
}

func (g *graphImpl) CanReach(
	ctx context.Context,
	source string,
	target string,
	usableRegions map[string]struct{},
) bool {
	if _, ok := usableRegions[source]; !ok {
		return false
	}

	if _, ok := usableRegions[target]; !ok {
		return false
	}

	visited := make(map[string]struct{})

	return g.canReachRecursive(ctx, source, target, usableRegions, visited)
}

func (g *graphImpl) canReachRecursive(
	ctx context.Context,
	source string,
	target string,
	usableRegions map[string]struct{},
	visited map[string]struct{},
) bool {
	if source == target {
		slog.DebugContext(ctx, "region is reachable", "source", source, "target", target)

		return true
	}

	if _, ok := visited[source]; ok {
		return false
	}

	visited[source] = struct{}{}

	for neighbour := range g.Edges[source] {
		if _, ok := usableRegions[neighbour]; !ok {
			continue
		}

		if g.canReachRecursive(ctx, neighbour, target, usableRegions, visited) {
			return true
		}
	}

	return false
}
