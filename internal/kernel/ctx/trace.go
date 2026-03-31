package ctx

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

type TraceContext interface {
	context.Context
	Span() trace.Span
}

type traceContext struct {
	context.Context //nolint:containedctx // deliberate context enrichment chain

	span trace.Span
}

var _ TraceContext = (*traceContext)(nil)

func (c *traceContext) Span() trace.Span {
	return c.span
}

func WithSpan(ctx context.Context, span trace.Span) TraceContext {
	return &traceContext{
		Context: ctx,
		span:    span,
	}
}

// WithSpanFromContext extracts the current span from ctx and wraps it in a
// TraceContext. Use this in business logic instead of importing otel/trace
// directly.
func WithSpanFromContext(ctx context.Context) TraceContext {
	return WithSpan(ctx, trace.SpanFromContext(ctx))
}
