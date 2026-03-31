// Package observe provides unified observability primitives for dual-signal
// emission: every call site emits both an slog log record and an OTel span
// event through a single function call. Business logic uses observe instead
// of importing log/slog or go.opentelemetry.io/otel directly.
//
// # API
//
// Five public functions cover the full observability surface:
//
//   - [Span] starts a child span and returns a done closure for lifecycle
//     management.
//   - [SpanEvent] adds a named event to the current span (trace-only, no log).
//   - [Info], [Warn], [Error] emit both an slog log and a span event.
//     [Error] additionally records the error on the span and sets its status
//     to Error.
//
// # The done(error) Lifecycle Pattern
//
// [Span] returns (ctx, done) where done must be called when the operation
// completes. This is the standard way to instrument a function:
//
//	func (s *service) CreateGame(ctx context.Context) error {
//	    ctx, done := observe.Span(ctx, "game.create")
//	    defer done(nil)
//
//	    // ... business logic ...
//	}
//
// When the function has a named error return, use a deferred closure to
// capture the final error value:
//
//	func (s *service) PerformMove(ctx context.Context) (err error) {
//	    ctx, done := observe.Span(ctx, "move.perform")
//	    defer func() { done(err) }()
//
//	    // ... business logic that may set err ...
//	}
//
// done(nil) simply ends the span with OK status. done(err) calls
// span.RecordError and sets the span status to Error before ending.
//
// # Span Naming Conventions
//
// Span names use noun.verb format that reads naturally in a trace waterfall:
//
//	observe.Span(ctx, "game.create")
//	observe.Span(ctx, "move.perform")
//	observe.Span(ctx, "phase.advance")
//	observe.Span(ctx, "board.loadGraph")
//
// The noun identifies the domain concept; the verb identifies the operation.
// This produces readable trace trees where each span is a self-describing
// operation.
//
// # Three-Signal Decision Framework
//
// Choose the right function based on what you are observing:
//
//   - [Span] — for operations with meaningful duration. Creates a parent-child
//     relationship in the trace tree. Use for service method entry points,
//     transactions, and I/O calls.
//   - [SpanEvent] — for decision points or milestones within an existing span.
//     No log emission, no new span. Use for "phase transitioned to ATTACK" or
//     "card combination validated" — events that matter for trace analysis but
//     do not need log persistence.
//   - [Info] / [Warn] / [Error] — for business events that need both trace
//     visibility and log persistence. The log is written via [slog.Default] for
//     structured indexing; the span event is written to the current span for
//     trace correlation. Use [Error] for failures that should mark the span as
//     failed.
//
// # Automatic Context Extraction
//
// All functions that accept a context automatically extract domain metadata
// via the [ctx.LogEnricher] interface. If the context implements LogEnricher
// (which [ctx.UserContext], [game/ctx.GameContext], and
// [lobby/ctx.LobbyContext] all do), attributes like user_id, game_id, and
// lobby_id are merged into span events without the call site passing them
// explicitly. The slog channel gets these attributes separately via the
// kernel slog ContextHandler, so observe only adds them to span events.
//
// # OTel Wiring
//
// The observe package is stateless — it calls [otel.GetTracerProvider] and
// [slog.Default] at each invocation, requiring no injected dependencies.
// The initialization chain that makes this work:
//
//  1. main → fx.New → otelsetup.Module
//  2. otelsetup.SetupOTelSDK creates TracerProvider and MeterProvider with
//     OTLP/HTTP exporters
//  3. otel.SetTracerProvider and otel.SetMeterProvider install them globally
//  4. observe.Span calls otel.GetTracerProvider().Tracer("go-risk-it").Start
//
// Because observe reads from the global OTel state, it works anywhere after
// the fx lifecycle starts — no constructor injection needed.
//
// # Layer
//
// Kernel — foundational observability infrastructure with no logic dependencies.
package observe
