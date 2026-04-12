// Package cards implements the cards move: a player plays card combinations
// to earn extra deployable troops and optional region troop grants. The
// phase always transitions to DEPLOY after cards are played, and the
// advancer calculates the total deployable troops (region ownership +
// continent bonuses + card combination bonuses).
//
// # Layer
//
// Logic — card combination validation, troop calculation, and card draw.
//
// # Key Types
//
// [Service] extends [service.Service] with card-specific operations:
// [Service.Draw] selects a random card from the available deck using RNG
// (pure computation — no DB queries or writes).
//
// [Move] is the input DTO containing one or more [CardCombination] entries.
//
// [MoveResult] carries the extra deployable troops and per-region troop
// grants awarded by the played combinations.
//
// # Dependencies
//
//   - [phase.Service] for inserting the deploy phase on advancement
//   - [rand.RNG] for random card selection during draw
package cards
