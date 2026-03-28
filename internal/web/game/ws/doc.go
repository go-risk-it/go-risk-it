// Package ws manages per-game WebSocket connections.
//
// The concrete [manager] type implements four narrow interfaces — [Writer],
// [Presence], [Lifecycle], and [Gateway] — composed into the aggregate
// [Manager] interface. It tracks connections in a map keyed by game ID,
// validates connection attempts against game state and player participation,
// and emits PlayerConnected events on the bus when a player upgrades to WS.
// Concurrency is handled by an upgradable RWMutex with double-check locking
// on connection creation.
//
// # Layer
//
// Web — per-game WebSocket connection management and message dispatch.
package ws
