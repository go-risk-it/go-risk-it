// Package phase manages game phase transitions. It defines the phase state
// machine (CARDS -> DEPLOY -> ATTACK -> CONQUER/REINFORCE and back),
// validates transitions, and handles turn advancement when crossing from
// REINFORCE to the next player's turn.
//
// # Layer
//
// Logic — phase state machine and turn management.
//
// # Key Types
//
// [Service] provides [Service.InsertPhase] which creates a new phase record,
// advances the turn when leaving REINFORCE (skipping eliminated players),
// and updates the game's current phase pointer.
//
// [ValidTransitions] is the phase state machine map defining which
// transitions are allowed. [ValidateTransition] checks a proposed
// transition against this map.
//
// # Dependencies
//
//   - [state.Service] for reading current game state (turn, phase)
//   - [player.Service] for player state used in turn advancement
package phase
