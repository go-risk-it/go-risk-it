// Package slog configures structured JSON logging with automatic context enrichment.
//
// [NewLogger] creates an slog.Logger that writes JSON to stderr, with the log
// level set to Debug in non-production environments and Info in production.
// The [ContextHandler] wraps the standard JSON handler and extracts trace IDs,
// user IDs, game IDs, and lobby IDs from the context chain, so every log line
// carries the relevant request metadata without manual enrichment.
//
// # Layer
//
// Kernel — logging configuration depending only on internal/ctx.
package slog
