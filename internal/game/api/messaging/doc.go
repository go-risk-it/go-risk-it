// Package messaging defines the WebSocket message types sent to game clients.
//
// Each state type represents one slice of the game state that is broadcast
// to connected players: board (regions and troops), cards (hand contents),
// game (phase and turn), missions (objective details), players (connection
// and alive/dead status), and move history. Generic types [GameState] and
// [MissionState] are parameterized over their phase- or mission-specific
// payloads using Go type-set constraints.
//
// # Layer
//
// API — JSON-serializable DTOs for WebSocket delivery.
package messaging
