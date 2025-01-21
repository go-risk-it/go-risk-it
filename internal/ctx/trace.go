package ctx

import "go.opentelemetry.io/otel/trace"

type TraceContext interface {
	LogContext
	Span() trace.Span
}

type traceContext struct {
	LogContext
	span trace.Span
}

var _ TraceContext = (*traceContext)(nil)

func (c *traceContext) Span() trace.Span {
	return c.span
}

func WithSpan(ctx LogContext, span trace.Span) TraceContext {
	ctx.SetLog(ctx.Log().With("traceID", span.SpanContext().TraceID()))

	return &traceContext{
		LogContext: ctx,
		span:       span,
	}
}
