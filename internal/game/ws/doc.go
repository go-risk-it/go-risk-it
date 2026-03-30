// Package ws manages per-game WebSocket connections.
//
// The concrete [manager] type implements four narrow interfaces — [Writer],
// [Presence], [Lifecycle], and [Gateway] — composed into the aggregate
// [Manager] interface. It tracks connections via [ws.ScopeMap] keyed by game ID
// and emits PlayerConnected events on the bus when a player upgrades to WS.
// Connection validation (game existence, player participation) is handled by
// the route handler before WebSocket upgrade — the manager is pure infrastructure.
//
// # Layer
//
// Web — per-game WebSocket connection management and message dispatch.
package ws
