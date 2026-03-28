// Package state provides the game state service for reading core game
// metadata: current turn, active phase, winner, and the list of games a
// user has joined.
//
// # Layer
//
// Logic — game state queries.
//
// # Key Types
//
// [Service] is the primary interface for reading game metadata: current
// turn, active phase, winner, and the list of games a user has joined.
//
// [Game] is the domain type returned by GetGameState, containing the
// game's ID, turn, phase, and winner user ID.
//
// # Dependencies
//
//   - [db.Querier] for game database queries
package state
