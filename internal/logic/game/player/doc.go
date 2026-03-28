// Package player manages game players: creation during game setup, querying
// player state (region counts, turn order), and determining the current and
// next active player. Eliminated players (zero regions) are skipped during
// turn rotation.
//
// # Layer
//
// Logic — player lifecycle and turn-order queries.
//
// # Key Types
//
// [Service] is the primary interface for player operations: creating players,
// fetching player state, and resolving current/next player by turn index.
//
// [Player] is the domain input type for a player joining a game, carrying
// the user ID and display name.
//
// # Dependencies
//
//   - [state.Service] for reading the current game turn
//   - [db.Querier] for player persistence and lookup queries
package player
