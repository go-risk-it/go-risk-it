// Package headlines detects and emits derived game events — player
// eliminations, continent captures, and continent losses — by maintaining
// an in-memory ownership cache that updates incrementally on each
// MoveExecuted event. Derived events are re-emitted through the event bus
// for downstream consumers (logging, WebSocket broadcast).
//
// # Layer
//
// Logic — derived event detection from ownership changes.
package headlines
