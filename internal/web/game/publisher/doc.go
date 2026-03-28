// Package publisher consumes game events from the event bus and publishes
// state updates over WebSocket connections.
//
// [GameStatePublisher] subscribes to MoveExecuted, PhaseTransitioned,
// GameCompleted, and PlayerConnected events. Each handler fetches the
// current game snapshot, converts it to WS messages via the converter
// package, and dispatches them through narrow WS interfaces (Writer,
// Presence, Lifecycle). Sub-operations use panic recovery (safeOp) so
// a failure in one delivery does not block the others.
//
// # Layer
//
// Web — reactive event-to-WebSocket bridge.
package publisher
