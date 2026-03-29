// Package advancement provides voluntary phase advancement for attack,
// cards, and reinforce moves. It wraps a move service's Walk and Advance
// methods in a RepeatableRead transaction and emits a PhaseTransitioned
// event after commit.
//
// # Layer
//
// Logic — voluntary phase advancement with transactional safety.
//
// # Key Types
//
// [Service] is the generic interface parameterized by move type T and
// result type R, providing Advance and AdvanceWithQuerier methods.
//
// [AttackAdvancer], [CardsAdvancer], and [ReinforceAdvancer] are type
// aliases that bind Service to concrete move and result types.
//
// # Dependencies
//
//   - [state.Service] for reading current game state within the transaction
//   - [service.Service] for move-specific walk and advance logic
//   - [validation.Service] for pre-advance game state validation
//   - [eventbus.Bus] for emitting PhaseTransitioned events after commit
//   - [metrics.Metrics] for recording advance operation metrics
package advancement
