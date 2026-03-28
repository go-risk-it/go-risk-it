// Package game defines shared types for the game API surface.
//
// It contains the [PhaseType] enum used by both REST request/response DTOs
// and WebSocket messaging types to identify the current game phase (cards,
// deploy, attack, conquer, reinforce).
//
// # Layer
//
// API — shared game API types with no internal dependencies.
package game
