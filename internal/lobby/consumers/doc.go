// Package consumers contains event handlers that react to lobby bus events
// (LobbyStateChanged, LobbyPlayerConnected) and publish state updates over
// WebSocket connections. Consumer-local interfaces (Writer) decouple this
// package from the concrete WebSocket manager in the web layer — Go duck
// typing handles satisfaction.
//
// # Layer
//
// Web — reactive event consumers bridging domain events to WebSocket delivery.
package consumers
