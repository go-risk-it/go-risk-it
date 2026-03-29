// Package logger provides a structured event logging consumer for the event bus.
//
// [Register] subscribes an OnAll handler that logs every emitted event as a
// single slog line at LevelInfo. Each log entry includes the event type,
// timestamp, scope ID (game or lobby, via type-switch), and the full
// ToRecord payload as a nested "payload" group. It is wired as an fx.Invoke
// consumer through [Module].
//
// # Layer
//
// Kernel — observability consumer with no domain logic dependencies.
package logger
