package ctx_test

import (
	"context"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	kernelctx "github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"github.com/stretchr/testify/require"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace/noop"
)

func TestRebase_SpanFromBase(t *testing.T) {
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
	traceCtx := kernelctx.WithSpan(context.Background(), originalSpan)
	userCtx := kernelctx.WithUserID(traceCtx, "test-user")
	gameCtx := ctx.WithGameID(userCtx, 42)

	// Verify precondition: GameContext initially has the original span.
	require.Equal(t, originalSpanID, gameCtx.Span().SpanContext().SpanID(),
		"precondition: gameCtx should carry the original span")

	// Call Rebase with the child span's context — this simulates what
	// observe.RawSpan does after otel.Tracer.Start() enriches the context.
	rebased := gameCtx.Rebase(childCtx)

	// The rebased context must carry the child span, not the original.
	rebasedGame, ok := rebased.(ctx.GameContext)
	require.True(t, ok, "Rebase must return a GameContext")
	require.Equal(t, childSpanID, rebasedGame.Span().SpanContext().SpanID(),
		"Rebase must extract the span from base, not from the original context")

	// Domain fields must be preserved.
	require.Equal(t, int64(42), rebasedGame.GameID())
	require.Equal(t, "test-user", rebasedGame.UserID())
}

func TestRebase_PreservesNoopSpanFromBase(t *testing.T) {
	t.Parallel()

	// Set up a real OTel tracer to create a real span for the original context.
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })

	tracer := tp.Tracer("test")

	_, originalSpan := tracer.Start(context.Background(), "original-span")
	t.Cleanup(func() { originalSpan.End() })

	// Build a GameContext with a real span.
	traceCtx := kernelctx.WithSpan(context.Background(), originalSpan)
	userCtx := kernelctx.WithUserID(traceCtx, "test-user")
	gameCtx := ctx.WithGameID(userCtx, 7)

	// Rebase with a plain context (no span embedded) should yield a noop span.
	rebased := gameCtx.Rebase(context.Background())

	rebasedGame, ok := rebased.(ctx.GameContext)
	require.True(t, ok, "Rebase must return a GameContext")

	// trace.SpanFromContext on a plain context returns a noop span with invalid SpanContext.
	require.False(t, rebasedGame.Span().SpanContext().IsValid(),
		"Rebase on a plain context should yield an invalid (noop) span")
}

func TestRebase_PropagatesCancellation(t *testing.T) {
	t.Parallel()

	base, cancel := context.WithCancel(context.Background())

	traceCtx := kernelctx.WithSpan(context.Background(), noop.Span{})
	userCtx := kernelctx.WithUserID(traceCtx, "test-user")
	gameCtx := ctx.WithGameID(userCtx, 42)

	rebased := gameCtx.Rebase(base)

	require.NoError(t, rebased.Err(), "rebased context should not be cancelled yet")

	cancel()

	require.Error(t, rebased.Err(), "cancelling base must propagate to rebased GameContext")
}

func TestRebase_PreservesDomainFields(t *testing.T) {
	t.Parallel()

	traceCtx := kernelctx.WithSpan(context.Background(), noop.Span{})
	userCtx := kernelctx.WithUserID(traceCtx, "user-abc")
	gameCtx := ctx.WithGameID(userCtx, 99)

	rebased := gameCtx.Rebase(context.Background())

	rebasedGame, ok := rebased.(ctx.GameContext)
	require.True(t, ok, "Rebase must return a GameContext")
	require.Equal(t, int64(99), rebasedGame.GameID())
	require.Equal(t, "user-abc", rebasedGame.UserID())
}

func TestGameContext_ScopeID(t *testing.T) {
	t.Parallel()

	traceCtx := kernelctx.WithSpan(context.Background(), noop.Span{})
	userCtx := kernelctx.WithUserID(traceCtx, "test-user")
	gameCtx := ctx.WithGameID(userCtx, 42)

	require.Equal(t, gameCtx.GameID(), gameCtx.ScopeID(),
		"ScopeID must equal GameID")
}
