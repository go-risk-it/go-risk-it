// Package ws manages per-lobby WebSocket connections.
//
// The concrete [manager] type implements two narrow interfaces — [Writer]
// and [Gateway] — composed into the aggregate [Manager] interface. It tracks
// connections via [ws.ScopeMap] keyed by lobby ID and emits LobbyPlayerConnected
// events on the bus when a player connects.
//
// # Layer
//
// Web — per-lobby WebSocket connection management and message dispatch.
package ws
