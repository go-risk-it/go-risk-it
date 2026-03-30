package routes_test

import (
	"context"
	"testing"

	gamectx "github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace/noop"
)

func newTracedContext(t *testing.T) context.Context {
	t.Helper()

	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	tracer := tp.Tracer("test")
	spanCtx, span := tracer.Start(t.Context(), "test-op")

	t.Cleanup(func() { span.End() })

	return spanCtx
}

func userContext(t *testing.T) ctx.UserContext {
	t.Helper()

	tc := ctx.WithSpan(t.Context(), noop.Span{})

	return ctx.WithUserID(tc, "user-123")
}

func testGameContext() gamectx.GameContext {
	uc := ctx.WithUserID(
		ctx.WithSpan(context.Background(), noop.Span{}),
		"user-123",
	)

	return gamectx.WithGameID(uc, 42)
}
