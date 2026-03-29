package tracing

import (
	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const GameTracer = "go-risk-it-game"

// StartGameSpan starts a new span and threads it through the GameContext chain.
// The caller is responsible for ending the returned span.
func StartGameSpan(
	gameCtx ctx.GameContext,
	spanName string,
	attrs ...attribute.KeyValue,
) (ctx.GameContext, trace.Span) {
	enrichedCtx, span := otel.GetTracerProvider(). //nolint:spancheck // caller ends span
							Tracer(GameTracer).
							Start(
			gameCtx, spanName,
			trace.WithAttributes(attrs...),
		)

	return gameCtx.WithBase(enrichedCtx), span //nolint:spancheck // caller ends span
}
