package slog_test

import (
	"bytes"
	"context"
	"encoding/json"
	stdslog "log/slog"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	riskslog "github.com/go-risk-it/go-risk-it/internal/slog"
	"github.com/stretchr/testify/require"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.uber.org/zap"
)

// parsed holds the JSON fields from a single slog log line.
type parsed map[string]any

func newTestLogger(buf *bytes.Buffer) *stdslog.Logger {
	inner := stdslog.NewJSONHandler(buf, &stdslog.HandlerOptions{Level: stdslog.LevelDebug})
	handler := riskslog.NewContextHandler(inner)

	return stdslog.New(handler)
}

func parseLine(t *testing.T, buf *bytes.Buffer) parsed {
	t.Helper()

	var result parsed
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))

	return result
}

func newTestSpan(t *testing.T) (context.Context, sdktrace.ReadWriteSpan) {
	t.Helper()

	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })

	tracer := tp.Tracer("test")
	spanCtx, span := tracer.Start(context.Background(), "test-span")

	t.Cleanup(func() { span.End() })

	rwSpan, ok := span.(sdktrace.ReadWriteSpan)
	require.True(t, ok)

	return spanCtx, rwSpan
}

func newLogContext() ctx.LogContext {
	return ctx.WithLog(context.Background(), zap.NewNop().Sugar())
}

func TestPlainContext(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := newTestLogger(&buf)

	logger.InfoContext(context.Background(), "plain message")

	result := parseLine(t, &buf)
	require.Equal(t, "plain message", result["msg"])

	// No context fields should be present.
	require.NotContains(t, result, "traceID")
	require.NotContains(t, result, "spanID")
	require.NotContains(t, result, "userID")
	require.NotContains(t, result, "gameID")
	require.NotContains(t, result, "lobbyID")
}

func TestTraceContext(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := newTestLogger(&buf)

	_, rwSpan := newTestSpan(t)
	logCtx := newLogContext()
	traceCtx := ctx.WithSpan(logCtx, rwSpan)

	logger.InfoContext(traceCtx, "trace message")

	result := parseLine(t, &buf)
	require.Equal(t, "trace message", result["msg"])
	require.Equal(t, rwSpan.SpanContext().TraceID().String(), result["traceID"])
	require.Equal(t, rwSpan.SpanContext().SpanID().String(), result["spanID"])
	require.NotContains(t, result, "userID")
	require.NotContains(t, result, "gameID")
	require.NotContains(t, result, "lobbyID")
}

func TestUserContext(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := newTestLogger(&buf)

	_, rwSpan := newTestSpan(t)
	logCtx := newLogContext()
	traceCtx := ctx.WithSpan(logCtx, rwSpan)
	userCtx := ctx.WithUserID(traceCtx, "user-123")

	logger.InfoContext(userCtx, "user message")

	result := parseLine(t, &buf)
	require.Equal(t, "user message", result["msg"])
	require.Equal(t, rwSpan.SpanContext().TraceID().String(), result["traceID"])
	require.Equal(t, rwSpan.SpanContext().SpanID().String(), result["spanID"])
	require.Equal(t, "user-123", result["userID"])
	require.NotContains(t, result, "gameID")
	require.NotContains(t, result, "lobbyID")
}

func TestGameContext(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := newTestLogger(&buf)

	_, rwSpan := newTestSpan(t)
	logCtx := newLogContext()
	traceCtx := ctx.WithSpan(logCtx, rwSpan)
	userCtx := ctx.WithUserID(traceCtx, "user-456")
	gameCtx := ctx.WithGameID(userCtx, 42)

	logger.InfoContext(gameCtx, "game message")

	result := parseLine(t, &buf)
	require.Equal(t, "game message", result["msg"])
	require.Equal(t, rwSpan.SpanContext().TraceID().String(), result["traceID"])
	require.Equal(t, rwSpan.SpanContext().SpanID().String(), result["spanID"])
	require.Equal(t, "user-456", result["userID"])
	require.InDelta(t, float64(42), result["gameID"], 0)
	require.NotContains(t, result, "lobbyID")
}

func TestLobbyContext(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := newTestLogger(&buf)

	_, rwSpan := newTestSpan(t)
	logCtx := newLogContext()
	traceCtx := ctx.WithSpan(logCtx, rwSpan)
	userCtx := ctx.WithUserID(traceCtx, "user-789")
	lobbyCtx := ctx.WithLobbyID(userCtx, 99)

	logger.InfoContext(lobbyCtx, "lobby message")

	result := parseLine(t, &buf)
	require.Equal(t, "lobby message", result["msg"])
	require.Equal(t, rwSpan.SpanContext().TraceID().String(), result["traceID"])
	require.Equal(t, rwSpan.SpanContext().SpanID().String(), result["spanID"])
	require.Equal(t, "user-789", result["userID"])
	require.InDelta(t, float64(99), result["lobbyID"], 0)
	require.NotContains(t, result, "gameID")
}

func TestWithAttrsPreserved(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	inner := stdslog.NewJSONHandler(&buf, &stdslog.HandlerOptions{Level: stdslog.LevelDebug})
	handler := riskslog.NewContextHandler(inner)

	// Add explicit attrs via WithAttrs.
	enriched := handler.WithAttrs([]stdslog.Attr{
		stdslog.String("component", "deploy"),
	})

	logger := stdslog.New(enriched)

	_, rwSpan := newTestSpan(t)
	logCtx := newLogContext()
	traceCtx := ctx.WithSpan(logCtx, rwSpan)
	userCtx := ctx.WithUserID(traceCtx, "user-attrs")

	logger.InfoContext(userCtx, "attrs message")

	result := parseLine(t, &buf)
	require.Equal(t, "attrs message", result["msg"])

	// Both explicit attrs and context attrs should be present.
	require.Equal(t, "deploy", result["component"])
	require.Equal(t, rwSpan.SpanContext().TraceID().String(), result["traceID"])
	require.Equal(t, "user-attrs", result["userID"])
}

func TestWithGroupPreserved(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	inner := stdslog.NewJSONHandler(&buf, &stdslog.HandlerOptions{Level: stdslog.LevelDebug})
	handler := riskslog.NewContextHandler(inner)

	// Wrap in a group and add attrs.
	grouped := handler.WithGroup("request").WithAttrs([]stdslog.Attr{
		stdslog.String("method", "POST"),
	})

	logger := stdslog.New(grouped)

	_, rwSpan := newTestSpan(t)
	logCtx := newLogContext()
	traceCtx := ctx.WithSpan(logCtx, rwSpan)

	logger.InfoContext(traceCtx, "grouped message")

	result := parseLine(t, &buf)
	require.Equal(t, "grouped message", result["msg"])

	// When WithGroup is active on the inner handler, all attrs — both explicit
	// and those added via record.AddAttrs() in Handle — are nested under the
	// group key. This is standard slog.JSONHandler behavior.
	requestGroup, ok := result["request"].(map[string]any)
	require.True(t, ok, "expected 'request' group in output")
	require.Equal(t, "POST", requestGroup["method"])
	require.Equal(t, rwSpan.SpanContext().TraceID().String(), requestGroup["traceID"])
	require.Equal(t, rwSpan.SpanContext().SpanID().String(), requestGroup["spanID"])
}
