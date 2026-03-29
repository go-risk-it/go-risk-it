package ctx_test

import (
	"context"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"github.com/stretchr/testify/require"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace/noop"
)

func TestWithBase_SpanFromBase(t *testing.T) {
	t.Parallel()

	// Set up a real OTel tracer provider with in-memory exporter.
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })

	tracer := tp.Tracer("test")

	// Create two distinct spans: an "original" and a "child".
	_, originalSpan := tracer.Start(context.Background(), "original-span")
	t.Cleanup(func() { originalSpan.End() })

	childCtx, childSpan := tracer.Start(context.Background(), "child-span")
	t.Cleanup(func() { childSpan.End() })

	originalSpanID := originalSpan.SpanContext().SpanID()
	childSpanID := childSpan.SpanContext().SpanID()

	// Sanity: the two spans must have different IDs.
	require.NotEqual(t, originalSpanID, childSpanID,
		"test setup: original and child spans must differ")

	// Build a GameContext carrying the original span.
	traceCtx := ctx.WithSpan(context.Background(), originalSpan)
	userCtx := ctx.WithUserID(traceCtx, "test-user")
	gameCtx := ctx.WithGameID(userCtx, 42)

	// Verify precondition: GameContext initially has the original span.
	require.Equal(t, originalSpanID, gameCtx.Span().SpanContext().SpanID(),
		"precondition: gameCtx should carry the original span")

	// Call WithBase with the child span's context — this simulates what
	// StartGameSpan does after otel.Tracer.Start() enriches the context.
	rebased := gameCtx.WithBase(childCtx)

	// The rebased context must carry the child span, not the original.
	require.Equal(t, childSpanID, rebased.Span().SpanContext().SpanID(),
		"WithBase must extract the span from base, not from the original context")

	// Domain fields must be preserved.
	require.Equal(t, int64(42), rebased.GameID())
	require.Equal(t, "test-user", rebased.UserID())
}

func TestWithBase_PreservesNoopSpanFromBase(t *testing.T) {
	t.Parallel()

	// Set up a real OTel tracer to create a real span for the original context.
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })

	tracer := tp.Tracer("test")

	_, originalSpan := tracer.Start(context.Background(), "original-span")
	t.Cleanup(func() { originalSpan.End() })

	// Build a GameContext with a real span.
	traceCtx := ctx.WithSpan(context.Background(), originalSpan)
	userCtx := ctx.WithUserID(traceCtx, "test-user")
	gameCtx := ctx.WithGameID(userCtx, 7)

	// WithBase with a plain context (no span embedded) should yield a noop span.
	rebased := gameCtx.WithBase(context.Background())

	// trace.SpanFromContext on a plain context returns a noop span with invalid SpanContext.
	require.False(t, rebased.Span().SpanContext().IsValid(),
		"WithBase on a plain context should yield an invalid (noop) span")
}

func TestWithBase_PropagatesCancellation(t *testing.T) {
	t.Parallel()

	base, cancel := context.WithCancel(context.Background())

	traceCtx := ctx.WithSpan(context.Background(), noop.Span{})
	userCtx := ctx.WithUserID(traceCtx, "test-user")
	gameCtx := ctx.WithGameID(userCtx, 42)

	rebased := gameCtx.WithBase(base)

	require.NoError(t, rebased.Err(), "rebased context should not be cancelled yet")

	cancel()

	require.Error(t, rebased.Err(), "cancelling base must propagate to rebased GameContext")
}

func TestWithBase_PreservesDomainFields(t *testing.T) {
	t.Parallel()

	traceCtx := ctx.WithSpan(context.Background(), noop.Span{})
	userCtx := ctx.WithUserID(traceCtx, "user-abc")
	gameCtx := ctx.WithGameID(userCtx, 99)

	rebased := gameCtx.WithBase(context.Background())

	require.Equal(t, int64(99), rebased.GameID())
	require.Equal(t, "user-abc", rebased.UserID())
}
