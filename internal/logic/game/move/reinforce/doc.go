// Package reinforce implements the reinforce move: a player moves troops
// between two owned regions that are connected through a path of owned
// territory. After reinforcement, the phase transitions to CARDS (if the
// next player has a valid combination) or DEPLOY (otherwise), advancing
// the turn. If the player conquered a region during the turn, a card is
// drawn before advancing.
//
// # Layer
//
// Logic — reinforcement validation, troop transfer, and turn advancement.
//
// # Key Types
//
// [Service] implements [service.Service] for the reinforce move type.
//
// [Move] is the input DTO carrying source/target region IDs, declared
// troop counts, and the number of troops to move.
//
// # Dependencies
//
//   - [board.Service] for reachability checks through player-owned regions
//   - [cards.Service] for card drawing and next-player combination checks
//   - [state.Service] for reading current game state (turn tracking)
//   - [phase.Service] for inserting the next phase on advancement
//   - [region.Service] for region ownership checks and troop updates
package reinforce
