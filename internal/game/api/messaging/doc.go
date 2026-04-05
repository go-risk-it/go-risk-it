// Package messaging defines the WebSocket message envelope and type constants
// for real-time communication with game clients.
//
// The protocol uses two message types: [PlayerViewType] delivers the full
// per-player game snapshot (board, cards, mission, phase, players, move log),
// and [PlayerConnectionType] notifies other players of connect/disconnect events.
//
// # Layer
//
// API — JSON-serializable DTOs for WebSocket delivery.
package messaging
