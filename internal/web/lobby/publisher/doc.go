// Package publisher consumes lobby events from the event bus and publishes
// state updates over WebSocket connections.
//
// [LobbyStatePublisher] subscribes to LobbyStateChanged and
// LobbyPlayerConnected events. On each event it fetches the current lobby
// state via the StateController, builds a WS message, and dispatches it
// through the narrow [ws.Writer] interface — broadcasting to all players
// on state changes, or writing to the single connecting player. Sub-operations
// use panic recovery (safeOp) for resilience.
//
// # Layer
//
// Web — reactive event-to-WebSocket bridge for lobbies.
package publisher
