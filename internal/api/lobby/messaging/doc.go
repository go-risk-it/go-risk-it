// Package messaging defines the WebSocket message types sent to lobby clients.
//
// It contains [LobbyState] with its list of [Participant] entries, broadcast
// to connected players whenever the lobby membership changes.
//
// # Layer
//
// API — JSON-serializable DTOs for WebSocket delivery.
package messaging
