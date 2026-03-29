// Package tracing provides helpers for creating OpenTelemetry spans within
// the game context chain.
//
// [StartGameSpan] begins a new span and threads it through the [ctx.GameContext]
// so that downstream code inherits the trace. Services self-instrument by calling
// StartGameSpan at method entry; the [service.TracedService] decorator wraps the
// [service.Service] interface with automatic span creation and error recording.
//
// # Layer
//
// Logic — game-domain tracing utilities depending only on internal/kernel/ctx.
package tracing
