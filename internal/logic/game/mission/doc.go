// Package mission manages game missions: creation, assignment to players,
// win-condition checking, and reassignment when a player is eliminated.
// Five mission types are supported — two continents, two continents plus one,
// eighteen territories, twenty-four territories, and eliminate player.
//
// # Layer
//
// Logic — mission lifecycle and win-condition evaluation.
//
// # Key Types
//
// [Service] is the primary interface for mission operations: creating missions
// during game setup, checking if the current player's mission is accomplished,
// and reassigning missions after a player elimination.
//
// [BaseMission] is the polymorphic interface implemented by all five mission
// types. Each type knows how to persist its specific data via the querier.
//
// # Dependencies
//
//   - [checker.Registry] for dispatching mission-type-specific win-condition checks
//   - [rand.RNG] for shuffling missions during assignment
//   - [db.Querier] for mission persistence and lookup queries
package mission
