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
// [Service.Draw] assigns a random available card to the current player, and
// [Service.NextPlayerHasValidCombination] checks if the next player must
// play cards before deploying.
//
// [Move] is the input DTO containing one or more [CardCombination] entries.
//
// [MoveResult] carries the extra deployable troops and per-region troop
// grants awarded by the played combinations.
//
// # Dependencies
//
//   - [board.Service] for continent control bonus calculation
//   - [phase.Service] for inserting the deploy phase on advancement
//   - [player.Service] for looking up the current and next player
//   - [region.Service] for region ownership and troop grants
//   - [rand.RNG] for random card selection during draw
package cards
