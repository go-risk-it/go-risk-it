package tracing

import (
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
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

// SpanStep creates a child span for a pipeline step, passing the enriched context
// to the function. Error recording is handled automatically.
func SpanStep(
	gameCtx ctx.GameContext,
	spanName string,
	phase string,
	stepFn func(ctx.GameContext) error,
) error {
	enrichedCtx, span := otel.GetTracerProvider().Tracer(GameTracer).Start(
		gameCtx, spanName,
		trace.WithAttributes(attribute.String("phase", phase)),
	)
	defer span.End()

	if err := stepFn(gameCtx.WithBase(enrichedCtx)); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return err
	}

	return nil
}
