package observe_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"

	gamectx "github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
	lobbyctx "github.com/go-risk-it/go-risk-it/internal/lobby/ctx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// setupTracing installs a real TracerProvider backed by an InMemoryExporter.
// Returns the exporter for span inspection and registers cleanup.
func setupTracing(t *testing.T) *tracetest.InMemoryExporter {
	t.Helper()

	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })

	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	t.Cleanup(func() { otel.SetTracerProvider(prev) })

	return exporter
}

// setupSlog replaces the default slog logger with a JSON handler writing to buf.
// Returns the buffer for log inspection and registers cleanup.
func setupSlog(t *testing.T) *bytes.Buffer {
	t.Helper()

	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := slog.New(handler)

	prev := slog.Default()
	slog.SetDefault(logger)
	t.Cleanup(func() { slog.SetDefault(prev) })

	return &buf
}

// parsedLog holds unmarshaled JSON from a single slog log line.
type parsedLog map[string]any

func parseLogLine(t *testing.T, buf *bytes.Buffer) parsedLog {
	t.Helper()

	var result parsedLog
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))

	return result
}

// startParentSpan creates a span on the current global TracerProvider and
// returns the context carrying it.
func startParentSpan(t *testing.T) (context.Context, trace.Span) {
	t.Helper()

	tracer := otel.GetTracerProvider().Tracer("test")
	spanCtx, span := tracer.Start(context.Background(), "parent-op")
	t.Cleanup(func() { span.End() })

	return spanCtx, span
}

// findSpan locates a span stub by name, or returns nil.
func findSpan(stubs tracetest.SpanStubs, name string) *tracetest.SpanStub {
	for i := range stubs {
		if stubs[i].Name == name {
			return &stubs[i]
		}
	}

	return nil
}

// findEvent locates an event by name within a span stub, or returns nil.
func findEvent(stub *tracetest.SpanStub, name string) *sdktrace.Event {
	for i := range stub.Events {
		if stub.Events[i].Name == name {
			return &stub.Events[i]
		}
	}

	return nil
}

// hasAttr checks if an event's attributes contain a key with the expected value.
func hasAttr(attrs []attribute.KeyValue, key string, expected attribute.Value) bool {
	for _, a := range attrs {
		if string(a.Key) == key && a.Value == expected {
			return true
		}
	}

	return false
}

// buildGameContext creates a full GameContext chain for testing context attribute extraction.
func buildGameContext(t *testing.T) gamectx.GameContext {
	t.Helper()

	parentCtx, parentSpan := startParentSpan(t)
	traceCtx := ctx.WithSpan(parentCtx, parentSpan)
	userCtx := ctx.WithUserID(traceCtx, "user-42")

	return gamectx.WithGameID(userCtx, 99)
}

// ---------------------------------------------------------------------------
// Tests: RawSpan
// ---------------------------------------------------------------------------

//nolint:paralleltest // swaps global TracerProvider
func TestRawSpan_CreatesAndEndsSpan(t *testing.T) {
	exporter := setupTracing(t)

	spanCtx, done := observe.RawSpan(context.Background(), "test-operation",
		attribute.String("key", "value"),
	)
	require.NotNil(t, spanCtx)

	// Span is active while not yet done.
	span := trace.SpanFromContext(spanCtx)
	require.True(t, span.SpanContext().IsValid(), "span must be valid")

	done(nil)

	stubs := exporter.GetSpans()
	stub := findSpan(stubs, "test-operation")
	require.NotNil(t, stub, "span must be recorded")

	// Verify the explicit attribute is on the span.
	require.True(t, hasAttr(stub.Attributes, "key", attribute.StringValue("value")),
		"span must carry the explicit attribute")

	// done(nil) should NOT set error status.
	assert.Equal(t, codes.Unset, stub.Status.Code, "span must not have error status on nil")
}

//nolint:paralleltest // swaps global TracerProvider
func TestRawSpan_DoneWithError_RecordsError(t *testing.T) {
	exporter := setupTracing(t)

	_, done := observe.RawSpan(context.Background(), "failing-op")

	testErr := errors.New("something went wrong")
	done(testErr)

	stubs := exporter.GetSpans()
	stub := findSpan(stubs, "failing-op")
	require.NotNil(t, stub, "span must be recorded")

	// Verify error status.
	assert.Equal(t, codes.Error, stub.Status.Code, "span must have Error status")
	assert.Equal(t, "something went wrong", stub.Status.Description)

	// Verify error event was recorded.
	errEvt := findEvent(stub, "exception")
	require.NotNil(t, errEvt, "span must have an exception event from RecordError")
}

//nolint:paralleltest // swaps global TracerProvider
func TestRawSpan_InheritsParentTrace(t *testing.T) {
	exporter := setupTracing(t)

	parentCtx, parentSpan := startParentSpan(t)
	parentTraceID := parentSpan.SpanContext().TraceID()
	parentSpanID := parentSpan.SpanContext().SpanID()

	_, done := observe.RawSpan(parentCtx, "child-op")
	done(nil)

	stubs := exporter.GetSpans()
	stub := findSpan(stubs, "child-op")
	require.NotNil(t, stub)

	assert.Equal(t, parentTraceID, stub.SpanContext.TraceID(),
		"child must be in same trace as parent")
	assert.Equal(t, parentSpanID, stub.Parent.SpanID(),
		"child's parent must be the parent span")
}

// ---------------------------------------------------------------------------
// Tests: SpanEvent
// ---------------------------------------------------------------------------

//nolint:paralleltest // swaps global TracerProvider
func TestSpanEvent_AddsEventToCurrentSpan(t *testing.T) {
	exporter := setupTracing(t)

	parentCtx, parentSpan := startParentSpan(t)

	observe.SpanEvent(parentCtx, "something-happened",
		attribute.Int64("count", 7),
	)

	parentSpan.End()

	stubs := exporter.GetSpans()
	stub := findSpan(stubs, "parent-op")
	require.NotNil(t, stub)

	evt := findEvent(stub, "something-happened")
	require.NotNil(t, evt, "event must be added to the current span")

	require.True(t, hasAttr(evt.Attributes, "count", attribute.Int64Value(7)),
		"explicit attr must be on the event")
}

//nolint:paralleltest // swaps global TracerProvider
func TestSpanEvent_MergesContextAttrs(t *testing.T) {
	exporter := setupTracing(t)

	gameCtx := buildGameContext(t)

	observe.SpanEvent(gameCtx, "ctx-event",
		attribute.String("extra", "val"),
	)

	// End the parent span that's on the game context.
	gameCtx.Span().End()

	stubs := exporter.GetSpans()
	stub := findSpan(stubs, "parent-op")
	require.NotNil(t, stub)

	evt := findEvent(stub, "ctx-event")
	require.NotNil(t, evt)

	// Context attrs (user_id, game_id) should be merged.
	require.True(t, hasAttr(evt.Attributes, "user_id", attribute.StringValue("user-42")),
		"context user_id must be present")
	require.True(t, hasAttr(evt.Attributes, "game_id", attribute.Int64Value(99)),
		"context game_id must be present")
	require.True(t, hasAttr(evt.Attributes, "extra", attribute.StringValue("val")),
		"explicit attr must be present")
}

// ---------------------------------------------------------------------------
// Tests: Info / Warn / Error — dual-signal emission
// ---------------------------------------------------------------------------

//nolint:paralleltest // swaps global TracerProvider + default slog
func TestInfo_EmitsSlogAndSpanEvent(t *testing.T) {
	exporter := setupTracing(t)
	buf := setupSlog(t)

	parentCtx, parentSpan := startParentSpan(t)

	observe.Info(parentCtx, "info message",
		attribute.String("service", "test"),
	)

	parentSpan.End()

	// Verify slog output.
	result := parseLogLine(t, buf)
	assert.Equal(t, "INFO", result["level"])
	assert.Equal(t, "info message", result["msg"])
	assert.Equal(t, "test", result["service"])

	// Verify span event.
	stubs := exporter.GetSpans()
	stub := findSpan(stubs, "parent-op")
	require.NotNil(t, stub)

	evt := findEvent(stub, "info message")
	require.NotNil(t, evt, "Info must add a span event")
	require.True(t, hasAttr(evt.Attributes, "service", attribute.StringValue("test")))
}

//nolint:paralleltest // swaps global TracerProvider + default slog
func TestWarn_EmitsSlogAndSpanEvent(t *testing.T) {
	exporter := setupTracing(t)
	buf := setupSlog(t)

	parentCtx, parentSpan := startParentSpan(t)

	observe.Warn(parentCtx, "warn message",
		attribute.Int64("threshold", 100),
	)

	parentSpan.End()

	// Verify slog output.
	result := parseLogLine(t, buf)
	assert.Equal(t, "WARN", result["level"])
	assert.Equal(t, "warn message", result["msg"])
	assert.InDelta(t, float64(100), result["threshold"], 0)

	// Verify span event.
	stubs := exporter.GetSpans()
	stub := findSpan(stubs, "parent-op")
	require.NotNil(t, stub)

	evt := findEvent(stub, "warn message")
	require.NotNil(t, evt, "Warn must add a span event")
	require.True(t, hasAttr(evt.Attributes, "threshold", attribute.Int64Value(100)))
}

//nolint:paralleltest // swaps global TracerProvider + default slog
func TestError_EmitsSlogAndSpanEventAndRecordsError(t *testing.T) {
	exporter := setupTracing(t)
	buf := setupSlog(t)

	parentCtx, parentSpan := startParentSpan(t)

	testErr := errors.New("db connection lost")
	observe.Error(parentCtx, testErr, "error message",
		attribute.String("db", "primary"),
	)

	parentSpan.End()

	// Verify slog output.
	result := parseLogLine(t, buf)
	assert.Equal(t, "ERROR", result["level"])
	assert.Equal(t, "error message", result["msg"])
	assert.Equal(t, "primary", result["db"])
	assert.Equal(t, "db connection lost", result["error"])

	// Verify span event (the message event, not the exception event).
	stubs := exporter.GetSpans()
	stub := findSpan(stubs, "parent-op")
	require.NotNil(t, stub)

	evt := findEvent(stub, "error message")
	require.NotNil(t, evt, "Error must add a span event for the message")

	// Verify RecordError was called (creates "exception" event).
	errEvt := findEvent(stub, "exception")
	require.NotNil(t, errEvt, "Error must call RecordError (exception event)")

	// Verify SetStatus.
	assert.Equal(t, codes.Error, stub.Status.Code, "span must have Error status")
	assert.Equal(t, "db connection lost", stub.Status.Description)
}

//nolint:paralleltest // swaps global TracerProvider + default slog
func TestInfo_WithGameContext_MergesContextAttrs(t *testing.T) {
	exporter := setupTracing(t)
	buf := setupSlog(t)

	gameCtx := buildGameContext(t)

	observe.Info(gameCtx, "game info",
		attribute.String("phase", "deploy"),
	)

	gameCtx.Span().End()

	// Slog should have the explicit attr (context attrs come from ContextHandler,
	// not from observe — observe adds them to span events only).
	result := parseLogLine(t, buf)
	assert.Equal(t, "game info", result["msg"])
	assert.Equal(t, "deploy", result["phase"])

	// Span event should have context attrs merged.
	stubs := exporter.GetSpans()
	stub := findSpan(stubs, "parent-op")
	require.NotNil(t, stub)

	evt := findEvent(stub, "game info")
	require.NotNil(t, evt)
	require.True(t, hasAttr(evt.Attributes, "user_id", attribute.StringValue("user-42")),
		"context user_id must be on span event")
	require.True(t, hasAttr(evt.Attributes, "game_id", attribute.Int64Value(99)),
		"context game_id must be on span event")
}

// ---------------------------------------------------------------------------
// Tests: Context attrs on span attributes (via RawSpan context merging)
// ---------------------------------------------------------------------------

//nolint:paralleltest // swaps global TracerProvider
func TestRawSpan_MergesContextAttrsOntoSpan(t *testing.T) {
	exporter := setupTracing(t)

	gameCtx := buildGameContext(t)

	childCtx, done := observe.RawSpan(gameCtx, "test-ctx-merge",
		attribute.String("phase", "ATTACK"),
	)
	_ = childCtx
	done(nil)

	// End the parent span too.
	gameCtx.Span().End()

	stubs := exporter.GetSpans()
	stub := findSpan(stubs, "test-ctx-merge")
	require.NotNil(t, stub)

	// Context attrs (user_id, game_id) should be span-level attributes.
	require.True(t, hasAttr(stub.Attributes, "user_id", attribute.StringValue("user-42")),
		"context user_id must be on span attributes")
	require.True(t, hasAttr(stub.Attributes, "game_id", attribute.Int64Value(99)),
		"context game_id must be on span attributes")
	// Explicit attr should also be present.
	require.True(t, hasAttr(stub.Attributes, "phase", attribute.StringValue("ATTACK")),
		"explicit attr must be on span attributes")
}

//nolint:paralleltest // swaps global TracerProvider
func TestRawSpan_PlainContext_NoContextAttrs(t *testing.T) {
	exporter := setupTracing(t)

	_, done := observe.RawSpan(context.Background(), "test-plain",
		attribute.String("key", "val"),
	)
	done(nil)

	stubs := exporter.GetSpans()
	stub := findSpan(stubs, "test-plain")
	require.NotNil(t, stub)

	// Only the explicit attr should be present.
	require.Len(t, stub.Attributes, 1)
	require.True(t, hasAttr(stub.Attributes, "key", attribute.StringValue("val")))
}

// ---------------------------------------------------------------------------
// Tests: LinkedSpan
// ---------------------------------------------------------------------------

//nolint:paralleltest // swaps global TracerProvider
func TestLinkedSpan_CreatesRootSpanWithLink(t *testing.T) {
	exporter := setupTracing(t)

	parentCtx, parentSpan := startParentSpan(t)

	childCtx, done := observe.LinkedSpan(parentCtx, "linked-op",
		attribute.String("event_type", "move_executed"),
	)
	_ = childCtx
	done(nil)
	parentSpan.End()

	stubs := exporter.GetSpans()
	stub := findSpan(stubs, "linked-op")
	require.NotNil(t, stub)

	// Should have the explicit attr.
	require.True(t, hasAttr(stub.Attributes, "event_type", attribute.StringValue("move_executed")))

	// Should be a root span (different trace ID from parent).
	parent := findSpan(stubs, "parent-op")
	require.NotNil(t, parent)
	assert.NotEqual(t, parent.SpanContext.TraceID(), stub.SpanContext.TraceID(),
		"linked span must be a separate trace")

	// Should have a link back to the parent span.
	require.Len(t, stub.Links, 1)
	assert.Equal(t, parent.SpanContext.SpanID(), stub.Links[0].SpanContext.SpanID(),
		"link must point to the trigger span")
}

//nolint:paralleltest // swaps global TracerProvider
func TestLinkedSpan_MergesContextAttrs(t *testing.T) {
	exporter := setupTracing(t)

	gameCtx := buildGameContext(t)

	_, done := observe.LinkedSpan(gameCtx, "linked-game",
		attribute.String("event_type", "game_completed"),
	)
	done(nil)
	gameCtx.Span().End()

	stubs := exporter.GetSpans()
	stub := findSpan(stubs, "linked-game")
	require.NotNil(t, stub)

	// Context attrs from the parent GameContext must be on the linked span.
	require.True(t, hasAttr(stub.Attributes, "user_id", attribute.StringValue("user-42")),
		"context user_id must be on linked span")
	require.True(t, hasAttr(stub.Attributes, "game_id", attribute.Int64Value(99)),
		"context game_id must be on linked span")
	require.True(
		t,
		hasAttr(stub.Attributes, "event_type", attribute.StringValue("game_completed")),
		"explicit attr must be on linked span",
	)
}

//nolint:paralleltest // swaps global TracerProvider
func TestLinkedSpan_NoParentSpan_DegracesGracefully(t *testing.T) {
	exporter := setupTracing(t)

	// No parent span — just a plain context.
	_, done := observe.LinkedSpan(context.Background(), "linked-no-parent")
	done(nil)

	stubs := exporter.GetSpans()
	stub := findSpan(stubs, "linked-no-parent")
	require.NotNil(t, stub, "span should still be created")
	assert.Empty(t, stub.Links, "no links when parent has no span")
}

//nolint:paralleltest // swaps global TracerProvider
func TestLinkedSpan_RecordsError(t *testing.T) {
	exporter := setupTracing(t)

	_, done := observe.LinkedSpan(context.Background(), "linked-error")
	done(errors.New("handler failed"))

	stubs := exporter.GetSpans()
	stub := findSpan(stubs, "linked-error")
	require.NotNil(t, stub)

	assert.Equal(t, codes.Error, stub.Status.Code)
	assert.Equal(t, "handler failed", stub.Status.Description)
}

// ---------------------------------------------------------------------------
// Tests: Graceful degradation — no span in context
// ---------------------------------------------------------------------------

//nolint:paralleltest // swaps global default slog
func TestInfo_NoSpanInContext_StillLogsSlog(t *testing.T) {
	buf := setupSlog(t)

	// No tracing setup — SpanFromContext will return a no-op span.
	observe.Info(context.Background(), "no-span info",
		attribute.String("key", "val"),
	)

	result := parseLogLine(t, buf)
	assert.Equal(t, "no-span info", result["msg"])
	assert.Equal(t, "val", result["key"])
}

// ---------------------------------------------------------------------------
// Tests: Auto-rebase — RawSpan preserves domain context type
// ---------------------------------------------------------------------------

// buildUserContext creates a bare UserContext chain for testing.
func buildUserContext(t *testing.T) ctx.UserContext {
	t.Helper()

	parentCtx, parentSpan := startParentSpan(t)
	traceCtx := ctx.WithSpan(parentCtx, parentSpan)

	return ctx.WithUserID(traceCtx, "user-bare")
}

//nolint:paralleltest // swaps global TracerProvider
func TestRawSpan_AutoRebase_GameContext_ReturnsGameContext(t *testing.T) {
	setupTracing(t)

	gameCtx := buildGameContext(t)

	childCtx, done := observe.RawSpan(gameCtx, "game-op")
	defer done(nil)

	gc, ok := childCtx.(gamectx.GameContext)
	require.True(t, ok, "RawSpan must return a GameContext when given a GameContext")
	assert.Equal(t, int64(99), gc.GameID())
	assert.Equal(t, "user-42", gc.UserID())
}

//nolint:paralleltest // swaps global TracerProvider
func TestRawSpan_AutoRebase_LobbyContext_ReturnsLobbyContext(t *testing.T) {
	setupTracing(t)

	lobbyCtx := buildLobbyContext(t)

	childCtx, done := observe.RawSpan(lobbyCtx, "lobby-op")
	defer done(nil)

	lc, ok := childCtx.(lobbyctx.LobbyContext)
	require.True(t, ok, "RawSpan must return a LobbyContext when given a LobbyContext")
	assert.Equal(t, int64(77), lc.LobbyID())
	assert.Equal(t, "user-42", lc.UserID())
}

//nolint:paralleltest // swaps global TracerProvider
func TestRawSpan_AutoRebase_UserContext_ReturnsUserContext(t *testing.T) {
	setupTracing(t)

	userCtx := buildUserContext(t)

	childCtx, done := observe.RawSpan(userCtx, "user-op")
	defer done(nil)

	uc, ok := childCtx.(ctx.UserContext)
	require.True(t, ok, "RawSpan must return a UserContext when given a UserContext")
	assert.Equal(t, "user-bare", uc.UserID())
}

//nolint:paralleltest // swaps global TracerProvider
func TestRawSpan_AutoRebase_PlainContext_ReturnsPlainContext(t *testing.T) {
	setupTracing(t)

	childCtx, done := observe.RawSpan(context.Background(), "plain-op")
	defer done(nil)

	// A plain context should NOT be a Rebaseable or any domain context.
	_, isRebaseable := childCtx.(ctx.Rebaseable)
	assert.False(t, isRebaseable, "plain context must not become Rebaseable after RawSpan")

	_, isGame := childCtx.(gamectx.GameContext)
	assert.False(t, isGame, "plain context must not become GameContext after RawSpan")
}

//nolint:paralleltest // swaps global TracerProvider
func TestRawSpan_AutoRebase_PreservesTraceParentChain(t *testing.T) {
	exporter := setupTracing(t)

	gameCtx := buildGameContext(t)
	parentSpan := trace.SpanFromContext(gameCtx)
	parentTraceID := parentSpan.SpanContext().TraceID()
	parentSpanID := parentSpan.SpanContext().SpanID()

	_, done := observe.RawSpan(gameCtx, "child-trace-op")
	done(nil)

	stubs := exporter.GetSpans()
	stub := findSpan(stubs, "child-trace-op")
	require.NotNil(t, stub, "child span must be recorded")

	assert.Equal(t, parentTraceID, stub.SpanContext.TraceID(),
		"child must be in same trace as parent")
	assert.Equal(t, parentSpanID, stub.Parent.SpanID(),
		"child's parent must be the parent span")
}
