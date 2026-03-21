package mapgraph

import (
	"encoding/json"
	"fmt"
	"os"
)

// BoardDTO matches the map.json structure (same as server's board/dto.go).
type BoardDTO struct {
	Regions    []RegionDTO    `json:"layers"`
	Continents []ContinentDTO `json:"continents"`
	Borders    []BorderDTO    `json:"links"`
}

type RegionDTO struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Continent string `json:"continent"`
}

type ContinentDTO struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	BonusTroops int    `json:"bonusTroops"`
}

type BorderDTO struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

// Graph provides adjacency queries over the Risk board.
type Graph struct {
	Regions    []string
	Continents map[string]ContinentDTO
	RegionTo   map[string]string // region → continent
	Neighbours map[string]map[string]struct{}
}

// LoadFromFile parses a map.json file and builds the graph.
func LoadFromFile(path string) (*Graph, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read map file: %w", err)
	}

	var dto BoardDTO
	if err := json.Unmarshal(data, &dto); err != nil {
		return nil, fmt.Errorf("parse map JSON: %w", err)
	}

	g := &Graph{
		Continents: make(map[string]ContinentDTO),
		RegionTo:   make(map[string]string),
		Neighbours: make(map[string]map[string]struct{}),
	}

	for _, c := range dto.Continents {
		g.Continents[c.ID] = c
	}

	for _, r := range dto.Regions {
		g.Regions = append(g.Regions, r.ID)
		g.RegionTo[r.ID] = r.Continent
		g.Neighbours[r.ID] = make(map[string]struct{})
	}

	for _, b := range dto.Borders {
		g.Neighbours[b.Source][b.Target] = struct{}{}
		g.Neighbours[b.Target][b.Source] = struct{}{}
	}

	return g, nil
}

// AreNeighbours returns true if source and target share a border.
func (g *Graph) AreNeighbours(source, target string) bool {
	if n, ok := g.Neighbours[source]; ok {
		_, exists := n[target]
		return exists
	}

	return false
}

// NeighboursOf returns all neighbours of the given region.
func (g *Graph) NeighboursOf(region string) []string {
	n, ok := g.Neighbours[region]
	if !ok {
		return nil
	}

	result := make([]string, 0, len(n))
	for neighbour := range n {
		result = append(result, neighbour)
	}

	return result
}
