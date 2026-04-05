// Package dice provides a configurable dice service for attack resolution.
// The roll strategy (random, attacker_always_wins, attacker_always_loses)
// is determined by configuration, enabling deterministic testing alongside
// production randomness.
//
// # Layer
//
// Logic — dice rolling for attack combat resolution.
//
// # Key Types
//
// [Service] is the primary interface providing RollAttackingDices and
// RollDefendingDices methods for combat resolution.
//
// [Roller] is the pluggable single-die roll interface with implementations:
// [Random] for production use (uniform 1-6) and [Sequence] for deterministic
// testing.
//
// # Dependencies
//
//   - [config.DiceConfig] for selecting the roll strategy at startup
//   - [rand.RNG] for the random dice source in production mode
package dice
