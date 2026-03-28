// Package card manages the card deck: creation during game setup, querying
// a player's hand, and transferring card ownership when a player is
// eliminated. Each region is mapped to one card (infantry, artillery, or
// cavalry), plus two jolly cards per game.
//
// # Layer
//
// Logic — card deck lifecycle and ownership management.
//
// # Key Types
//
// [Service] is the primary interface for card operations: creation,
// querying a player's hand, and transferring ownership on elimination.
//
// # Dependencies
//
//   - [region.Service] for mapping regions to cards during deck creation
//   - [rand.RNG] for shuffling the deck before card-type assignment
package card
