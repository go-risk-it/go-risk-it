// Package response defines the JSON response bodies for lobby REST endpoints.
//
// It contains [CreateLobby] (returned after lobby creation) and [Lobbies]
// (returned when listing lobbies, split into owned, joined, and joinable).
//
// # Layer
//
// API — outbound REST DTOs with no business logic.
package response
