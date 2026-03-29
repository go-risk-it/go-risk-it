// Package consumers contains event handlers that react to game bus events
// (MoveExecuted, PhaseTransitioned, GameCompleted, PlayerConnected) and
// publish state updates over WebSocket connections. Consumer-local interfaces
// (Writer, Presence, Lifecycle) decouple this package from the concrete
// WebSocket manager in the web layer — Go duck typing handles satisfaction.
//
// # Layer
//
// Web — reactive event consumers bridging domain events to WebSocket delivery.
package consumers
