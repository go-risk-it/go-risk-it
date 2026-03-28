// Package tracing provides helpers for creating OpenTelemetry spans within
// the game context chain.
//
// [StartGameSpan] begins a new span and threads it through the [ctx.GameContext]
// so that downstream code inherits the trace. [SpanStep] wraps a function call
// in a child span with automatic error recording, suitable for pipeline steps
// in the move execution path.
//
// # Layer
//
// Infrastructure — tracing utilities depending only on internal/ctx.
package tracing
