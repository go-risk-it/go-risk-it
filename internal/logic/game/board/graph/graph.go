package graph

import (
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

	return result
}

var _ Graph = (*GraphImpl)(nil)

func New(board *dto.Board) *GraphImpl {
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

	return topG
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
