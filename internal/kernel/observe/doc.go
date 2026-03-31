// Package observe provides unified observability primitives for dual-signal
// emission: every call site emits both an slog log record and an OTel span
// event through a single function call. Business logic uses observe instead
// of importing log/slog or go.opentelemetry.io/otel directly.
//
// # API
//
// Six public functions cover the full observability surface:
//
//   - [Span] wraps a typed-context function that returns (T, error) with
//     automatic span lifecycle. The closure receives the span-enriched typed
//     context directly — no named returns, no defer-done discipline, no risk
//     of discarding the context. This is the primary span creation function
//     for business logic.
//   - [SpanErr] is the error-only variant of [Span] — for functions that
//     return only an error (no result value).
//   - [RawSpan] starts a child span on a plain context.Context and returns
//     the enriched context plus a done function for manual lifecycle
//     management. Only needed for infrastructure callers (event bus,
//     transaction wrapper) that operate on untyped contexts. When the parent
//     implements [ctx.Rebaseable], domain metadata is automatically preserved
//     on the child context (see Auto-Rebase below).
//   - [SpanEvent] adds a named event to the current span (trace-only, no log).
//   - [Info], [Warn], [Error] emit both an slog log and a span event.
//     [Error] additionally records the error on the span and sets its status
//     to Error.
//
// # Auto-Rebase
//
// OTel's tracer.Start returns a plain context.Context wrapping the new span,
// which discards any domain metadata (GameID, UserID, etc.) carried by typed
// context types. [RawSpan] solves this automatically: when the parent implements
// [ctx.Rebaseable], RawSpan calls Rebase to copy domain metadata onto the
// OTel-enriched child context. Call sites get a context that carries both
// the new span and the original domain fields — no manual rebasing needed.
//
// [Span] and [SpanErr] build on this: the closure receives the same typed
// context as the parent, so downstream code keeps compile-time access to
// domain fields without a type assertion.
//
// # Span Creation: Choosing the Right Function
//
// Use [Span] when the function returns (T, error) on a typed context:
//
//	func (s *service) CreateGame(ctx gamectx.GameContext) (int64, error) {
//	    return observe.Span(ctx, "game.create",
//	        func(ctx gamectx.GameContext) (int64, error) {
//	            // ctx carries the child span — compile-time safe
//	            return gameID, nil
//	        },
//	    )
//	}
//
// Use [SpanErr] for typed-context functions that return only an error:
//
//	func (s *service) Advance(ctx gamectx.GameContext) error {
//	    return observe.SpanErr(ctx, "game.advance",
//	        func(ctx gamectx.GameContext) error {
//	            // ... business logic ...
//	            return nil
//	        },
//	    )
//	}
//
// Use [RawSpan] + defer done(nil) for void functions on plain contexts
// (infrastructure only):
//
//	func (s *service) BroadcastState(ctx context.Context) {
//	    ctx, done := observe.RawSpan(ctx, "state.broadcast")
//	    defer done(nil)
//	    // ...
//	}
//
// WARNING: never use defer done(nil) in a function that returns error — the
// span will always record success even when the function fails. In
// error-returning functions, always use [Span] or [SpanErr] instead.
//
// An arch_test rule (Rule 25) enforces this: defer done(nil) in an
// error-returning function is a build-time failure. Rule O3 additionally
// restricts [RawSpan] to infrastructure packages — business logic must use
// [Span] or [SpanErr].
//
// # Usage Rules
//
//  1. [Error] is for partial failures only — when execution continues despite
//     the error. If you are returning the error, just return it; done(err) on
//     the span handles status marking.
//  2. Use [Span]/[SpanErr] for business logic (closure-based, structurally
//     safe). Use [RawSpan]+done for infrastructure only (manual lifecycle).
//  3. Attrs are optional — context attributes (user_id, game_id, lobby_id)
//     are auto-extracted via [ctx.LogEnricher]. Pass explicit attrs only for
//     operation-specific metadata.
//
// # Span Naming Conventions
//
// Span names use noun.verb format that reads naturally in a trace waterfall:
//
//	observe.Span(ctx, "game.create", func(ctx gamectx.GameContext) (int64, error) { ... })
//	observe.SpanErr(ctx, "game.advance", func(ctx gamectx.GameContext) error { ... })
//	observe.RawSpan(ctx, "db.transaction")
//
// The noun identifies the domain concept; the verb identifies the operation.
// This produces readable trace trees where each span is a self-describing
// operation.
//
// # Three-Signal Decision Framework
//
// Choose the right function based on what you are observing:
//
//   - [Span] / [SpanErr] / [RawSpan] — for operations with meaningful
//     duration. Creates a parent-child relationship in the trace tree. Use
//     for service method entry points, transactions, and I/O calls.
//   - [SpanEvent] — for decision points or milestones within an existing span.
//     No log emission, no new span. Use for "phase transitioned to ATTACK" or
//     "card combination validated" — events that matter for trace analysis but
//     do not need log persistence.
//   - [Info] / [Warn] / [Error] — for business events that need both trace
//     visibility and log persistence. The log is written via [slog.Default] for
//     structured indexing; the span event is written to the current span for
//     trace correlation. Use [Error] for partial failures that should mark the
//     span as failed while execution continues.
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
//  4. observe.RawSpan calls otel.GetTracerProvider().Tracer("go-risk-it").Start
//
// Because observe reads from the global OTel state, it works anywhere after
// the fx lifecycle starts — no constructor injection needed.
//
// # Layer
//
// Kernel — foundational observability infrastructure with no logic dependencies.
package observe
