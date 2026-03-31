// Package tracing provides helpers for creating OpenTelemetry spans within
// the game context chain.
//
// [StartGameSpan] begins a new span via [observe.Span] and threads it through
// the [ctx.GameContext] so that downstream code inherits the trace. It returns a
// done function that the caller must invoke to end the span and optionally record
// an error. Services self-instrument by calling StartGameSpan at method entry;
// the [service.TracedService] decorator wraps the [service.Service] interface
// with automatic span creation and error recording.
//
// # Layer
//
// Logic — game-domain tracing utilities depending only on internal/kernel/ctx
// and internal/kernel/observe.
package tracing
