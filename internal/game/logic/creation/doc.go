// Package creation orchestrates full game setup: inserting the game record,
// creating players, assigning missions and regions, generating the card deck,
// and creating the initial deploy phase. The entire operation runs inside a
// single transaction, with post-commit events and metric recording.
//
// # Layer
//
// Logic — game creation orchestration.
//
// # Key Types
//
// [Service] is the primary interface providing CreateGame and
// CreateGameWithQuerier methods for full game setup.
//
// # Dependencies
//
//   - [card.Service] for generating the card deck
//   - [mission.Service] for assigning missions to players
//   - [player.Service] for creating player records
//   - [region.Service] for creating and assigning regions
//   - [eventbus.Publisher] for emitting GameCreated events after commit
//   - [metrics.Metrics] for recording game creation counts
//   - [timing.GameTiming] for tracking game start times
package creation
