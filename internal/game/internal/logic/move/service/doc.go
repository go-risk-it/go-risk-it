// Package service defines the generic move-service contract and its associated
// context and effect types. Each concrete move package (deploy, attack, conquer,
// reinforce, cards) implements [Service] parameterized by its move input T and
// result type R.
//
// # Layer
//
// Logic — generic move interfaces and value types.
//
// # Key Types
//
// [Service] is the composite interface combining [Performer], [PhaseWalker],
// and [Advancer] into a single move-type contract.
//
// [Performer] executes a move within a transaction and returns a result R
// plus a [MoveEffect] describing all observable state changes.
//
// [PhaseWalker] determines the next game phase given a [WalkContext].
//
// [Advancer] applies post-transition logic (troop grants, card awards) given
// a target phase, result R, and [AdvanceContext].
//
// [MoveEffect] and [AdvanceEffect] are side-effect bundles capturing region
// updates, card deltas, and mission changes produced by performers and
// advancers respectively.
//
// # Dependencies
//
//   - [snapshot] types for cached game state passed through contexts
//   - [board.Continents] for continent data in advance contexts
//   - [db.Querier] for transactional database access in performer/advancer calls
package service
