package graph

import (
	"errors"
	"fmt"
	"slices"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/board/dto"
)

type Graph interface {
	GetRegions() []string
	AreNeighbours(source string, target string) bool
	CanReach(
		context ctx.LogContext,
		source string,
		target string,
		usableRegions map[string]struct{},
	) bool
}

type GraphImpl struct {
	Edges map[string]map[string]struct{}
}

func (g *GraphImpl) GetRegions() []string {
	result := make([]string, 0, len(g.Edges))
	for region := range g.Edges {
		result = append(result, region)
	}

	slices.Sort(result)

	return result
}

var _ Graph = (*GraphImpl)(nil)

func New(board *dto.Board) (*GraphImpl, error) {
	err := validate(board)
	if err != nil {
		return nil, fmt.Errorf("invalid board: %w", err)
	}

	topG := &GraphImpl{
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

func validate(board *dto.Board) error {
	if len(board.Regions) == 0 {
		return errors.New("no regions")
	}

	if len(board.Borders) == 0 {
		return errors.New("no borders")
	}

	regionNames := make(map[string]struct{})
	for _, region := range board.Regions {
		if _, ok := regionNames[region.ExternalReference]; ok {
			return errors.New("duplicate region")
		}

		regionNames[region.ExternalReference] = struct{}{}
	}

	err := validateBorders(board, regionNames)
	if err != nil {
		return fmt.Errorf("invalid borders: %w", err)
	}

	return nil
}

func validateBorders(board *dto.Board, regionNames map[string]struct{}) error {
	for _, border := range board.Borders {
		if border.Source == "" {
			return errors.New("empty source")
		}

		if border.Target == "" {
			return errors.New("empty target")
		}

		if border.Source == border.Target {
			return errors.New("self-loop")
		}

		if _, ok := regionNames[border.Source]; !ok {
			return fmt.Errorf("unknown source %v", border.Source)
		}

		if _, ok := regionNames[border.Target]; !ok {
			return fmt.Errorf("unknown target %v", border.Target)
		}
	}

	return nil
}

func (g *GraphImpl) AreNeighbours(source string, target string) bool {
	_, ok := g.Edges[source][target]

	return ok
}

func (g *GraphImpl) CanReach(
	context ctx.LogContext,
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

	return g.canReachRecursive(context, source, target, usableRegions, visited)
}

func (g *GraphImpl) canReachRecursive(
	context ctx.LogContext,
	source string,
	target string,
	usableRegions map[string]struct{},
	visited map[string]struct{},
) bool {
	if source == target {
		context.Log().Debugw("region is reachable", "source", source, "target", target)

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

		if g.canReachRecursive(context, neighbour, target, usableRegions, visited) {
			return true
		}
	}

	context.Log().Debugw("target unreachable", "source", source, "target", target)

	return false
}
