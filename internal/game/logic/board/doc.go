// Package board models the Risk game board as a graph of regions and
// continents. It lazily loads the board topology from a JSON file and
// provides adjacency queries (neighbour checks, reachability through
// player-owned regions) and continent control calculations.
//
// # Layer
//
// Logic — board topology, adjacency, and continent ownership queries.
//
// # Key Types
//
// [Service] is the primary interface for board queries: region listing,
// neighbour checks, player-scoped reachability, and continent control.
//
// [Graph] represents the board's region adjacency graph, constructed from
// [BoardDto]. It supports neighbour and reachability queries.
//
// [Continents] provides continent-level queries, including listing all
// continents and determining which are fully controlled by a set of regions.
//
// [Continent] holds a continent's metadata (external reference, bonus troops)
// and its constituent region identifiers.
//
// # Dependencies
//
//   - [region.Service] for fetching region ownership used in reachability checks
package board
