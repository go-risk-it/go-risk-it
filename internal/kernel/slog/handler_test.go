package slog_test

import (
	"bytes"
	"context"
	"encoding/json"
	stdslog "log/slog"
	"testing"

	gamectx "github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	riskslog "github.com/go-risk-it/go-risk-it/internal/kernel/slog"
	lobbyclx "github.com/go-risk-it/go-risk-it/internal/lobby/ctx"
	"github.com/stretchr/testify/require"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// parsed holds the JSON fields from a single slog log line.
type parsed map[string]any

func newTestLogger(buf *bytes.Buffer) *stdslog.Logger {
	inner := stdslog.NewJSONHandler(buf, &stdslog.HandlerOptions{Level: stdslog.LevelDebug})
	handler := riskslog.NewContextHandler(inner, stdslog.LevelDebug)

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
	require.NotContains(t, result, "user_id")
	require.NotContains(t, result, "game_id")
	require.NotContains(t, result, "lobby_id")
}

func TestTraceContext(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := newTestLogger(&buf)

	_, rwSpan := newTestSpan(t)
	traceCtx := ctx.WithSpan(context.Background(), rwSpan)

	logger.InfoContext(traceCtx, "trace message")

	result := parseLine(t, &buf)
	require.Equal(t, "trace message", result["msg"])

	// traceID and spanID are NOT extracted by ContextHandler — the otelslog
	// bridge handles that. With a plain JSONHandler inner (used in tests),
	// they should not appear.
	require.NotContains(t, result, "traceID")
	require.NotContains(t, result, "spanID")
	require.NotContains(t, result, "user_id")
	require.NotContains(t, result, "game_id")
	require.NotContains(t, result, "lobby_id")
}

func TestUserContext(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := newTestLogger(&buf)

	_, rwSpan := newTestSpan(t)
	traceCtx := ctx.WithSpan(context.Background(), rwSpan)
	userCtx := ctx.WithUserID(traceCtx, "user-123")

	logger.InfoContext(userCtx, "user message")

	result := parseLine(t, &buf)
	require.Equal(t, "user message", result["msg"])
	require.Equal(t, "user-123", result["user_id"])
	require.NotContains(t, result, "traceID")
	require.NotContains(t, result, "spanID")
	require.NotContains(t, result, "game_id")
	require.NotContains(t, result, "lobby_id")
}

func TestGameContext(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := newTestLogger(&buf)

	_, rwSpan := newTestSpan(t)
	traceCtx := ctx.WithSpan(context.Background(), rwSpan)
	userCtx := ctx.WithUserID(traceCtx, "user-456")
	gameCtx := gamectx.WithGameID(userCtx, 42)

	logger.InfoContext(gameCtx, "game message")

	result := parseLine(t, &buf)
	require.Equal(t, "game message", result["msg"])
	require.Equal(t, "user-456", result["user_id"])
	require.InDelta(t, float64(42), result["game_id"], 0)
	require.NotContains(t, result, "traceID")
	require.NotContains(t, result, "spanID")
	require.NotContains(t, result, "lobby_id")
}

func TestLobbyContext(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := newTestLogger(&buf)

	_, rwSpan := newTestSpan(t)
	traceCtx := ctx.WithSpan(context.Background(), rwSpan)
	userCtx := ctx.WithUserID(traceCtx, "user-789")
	lobbyCtx := lobbyclx.WithLobbyID(userCtx, 99)

	logger.InfoContext(lobbyCtx, "lobby message")

	result := parseLine(t, &buf)
	require.Equal(t, "lobby message", result["msg"])
	require.Equal(t, "user-789", result["user_id"])
	require.InDelta(t, float64(99), result["lobby_id"], 0)
	require.NotContains(t, result, "traceID")
	require.NotContains(t, result, "spanID")
	require.NotContains(t, result, "game_id")
}

func TestWithAttrsPreserved(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	inner := stdslog.NewJSONHandler(&buf, &stdslog.HandlerOptions{Level: stdslog.LevelDebug})
	handler := riskslog.NewContextHandler(inner, stdslog.LevelDebug)

	// Add explicit attrs via WithAttrs.
	enriched := handler.WithAttrs([]stdslog.Attr{
		stdslog.String("component", "deploy"),
	})

	logger := stdslog.New(enriched)

	_, rwSpan := newTestSpan(t)
	traceCtx := ctx.WithSpan(context.Background(), rwSpan)
	userCtx := ctx.WithUserID(traceCtx, "user-attrs")

	logger.InfoContext(userCtx, "attrs message")

	result := parseLine(t, &buf)
	require.Equal(t, "attrs message", result["msg"])

	// Both explicit attrs and context attrs should be present.
	require.Equal(t, "deploy", result["component"])
	require.Equal(t, "user-attrs", result["user_id"])
	require.NotContains(t, result, "traceID")
	require.NotContains(t, result, "spanID")
}

func TestWithGroupPreserved(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	inner := stdslog.NewJSONHandler(&buf, &stdslog.HandlerOptions{Level: stdslog.LevelDebug})
	handler := riskslog.NewContextHandler(inner, stdslog.LevelDebug)

	// Wrap in a group and add attrs.
	grouped := handler.WithGroup("request").WithAttrs([]stdslog.Attr{
		stdslog.String("method", "POST"),
	})

	logger := stdslog.New(grouped)

	_, rwSpan := newTestSpan(t)
	traceCtx := ctx.WithSpan(context.Background(), rwSpan)

	logger.InfoContext(traceCtx, "grouped message")

	result := parseLine(t, &buf)
	require.Equal(t, "grouped message", result["msg"])

	// When WithGroup is active on the inner handler, all attrs — both explicit
	// and those added via record.AddAttrs() in Handle — are nested under the
	// group key. This is standard slog.JSONHandler behavior.
	requestGroup, ok := result["request"].(map[string]any)
	require.True(t, ok, "expected 'request' group in output")
	require.Equal(t, "POST", requestGroup["method"])
	require.NotContains(t, requestGroup, "traceID")
	require.NotContains(t, requestGroup, "spanID")
}

func TestLevelFiltering(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	inner := stdslog.NewJSONHandler(&buf, &stdslog.HandlerOptions{Level: stdslog.LevelDebug})
	handler := riskslog.NewContextHandler(inner, stdslog.LevelInfo)

	logger := stdslog.New(handler)

	// Debug should be filtered out by ContextHandler's level gate.
	logger.DebugContext(context.Background(), "debug message")
	require.Empty(t, buf.Bytes(), "debug message should be filtered")

	// Info should pass through.
	logger.InfoContext(context.Background(), "info message")

	result := parseLine(t, &buf)
	require.Equal(t, "info message", result["msg"])
}
