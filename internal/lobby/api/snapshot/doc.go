// Package snapshot defines the lobby snapshot types used as the canonical
// representation of lobby state across module boundaries.
//
// [LobbySnapshot] and [Participant] are pure value types with JSON tags matching
// the existing WebSocket wire format (see lobby/api/messaging).
//
// # Layer
//
// API — JSON-serializable domain types. Must not import logic/ or data/.
package snapshot
