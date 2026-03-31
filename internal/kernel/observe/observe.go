package observe

import (
	"context"
	"log/slog"

	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "go-risk-it"

// RawSpan starts a new child span and returns the enriched context plus a done
// function. Callers must call done when the operation completes:
//
//	ctx, done := observe.RawSpan(ctx, "operation", attrs...)
//	defer done(nil)
//
// If the operation produces an error via named returns, use:
//
//	defer func() { done(err) }()
//
// done(err) records the error on the span and sets the span status to Error.
// done(nil) simply ends the span.
//
// Prefer [Span] (typed generic) for business logic with typed contexts
// (GameContext, LobbyContext). Use RawSpan only when the parent is a plain
// context.Context without domain metadata — typically infrastructure callers
// like the event bus or transaction wrapper.

func RawSpan(
	parent context.Context,
	name string,
	attrs ...attribute.KeyValue,
) (context.Context, func(error)) {
	childCtx, span := otel.GetTracerProvider().
		Tracer(tracerName).
		Start(parent, name, trace.WithAttributes(attrs...))

	// Auto-preserve domain type across span boundary.
	if r, ok := parent.(ctx.Rebaseable); ok {
		childCtx = r.Rebase(childCtx)
	}

	done := func(err error) {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}

		span.End()
	}

	return childCtx, done
}

// SpanEvent adds a named event to the current span in ctx. Context attributes
// (from the LogEnricher chain) are automatically merged with the explicit attrs.
func SpanEvent(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	allAttrs := mergeContextAttrs(ctx, attrs)
	span.AddEvent(name, trace.WithAttributes(allAttrs...))
}

// Info emits a dual-signal observation: an slog log at INFO level and a span
// event on the current span. Context attributes are merged into the span event
// automatically.
func Info(ctx context.Context, msg string, attrs ...attribute.KeyValue) {
	emitLog(ctx, slog.LevelInfo, msg, nil, attrs)
	emitSpanEvent(ctx, msg, attrs)
}

// Warn emits a dual-signal observation: an slog log at WARN level and a span
// event on the current span.
func Warn(ctx context.Context, msg string, attrs ...attribute.KeyValue) {
	emitLog(ctx, slog.LevelWarn, msg, nil, attrs)
	emitSpanEvent(ctx, msg, attrs)
}

// Error emits a dual-signal observation: an slog log at ERROR level and a span
// event on the current span. Additionally, it calls span.RecordError and
// span.SetStatus to mark the span as failed.
func Error(ctx context.Context, err error, msg string, attrs ...attribute.KeyValue) {
	emitLog(ctx, slog.LevelError, msg, err, attrs)
	emitSpanEvent(ctx, msg, attrs)

	span := trace.SpanFromContext(ctx)
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}

// ExtractContextAttrs type-asserts ctx as a [ctx.LogEnricher] and converts the
// resulting slog.Attr slice to OTel attribute.KeyValue. Returns nil for contexts
// that do not implement LogEnricher.
//
// Exported for testing — production code should not call this directly.
func ExtractContextAttrs(c context.Context) []attribute.KeyValue {
	return extractContextAttrs(c)
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

// extractContextAttrs performs the LogEnricher type assertion and conversion.
func extractContextAttrs(c context.Context) []attribute.KeyValue {
	enricher, ok := c.(ctx.LogEnricher)
	if !ok {
		return nil
	}

	slogAttrs := enricher.SlogAttrs()
	if len(slogAttrs) == 0 {
		return nil
	}

	result := make([]attribute.KeyValue, 0, len(slogAttrs))

	for _, a := range slogAttrs {
		result = append(result, slogAttrToOtel(a))
	}

	return result
}

// slogAttrToOtel converts a single slog.Attr to an OTel attribute.KeyValue.
func slogAttrToOtel(attr slog.Attr) attribute.KeyValue {
	switch attr.Value.Kind() {
	case slog.KindString:
		return attribute.String(attr.Key, attr.Value.String())
	case slog.KindInt64:
		return attribute.Int64(attr.Key, attr.Value.Int64())
	case slog.KindFloat64:
		return attribute.Float64(attr.Key, attr.Value.Float64())
	case slog.KindBool:
		return attribute.Bool(attr.Key, attr.Value.Bool())
	default:
		// Fall back to the string representation for unsupported types.
		return attribute.String(attr.Key, attr.Value.String())
	}
}

// otelToSlog converts OTel attributes to slog attributes for the log channel.
func otelToSlog(attrs []attribute.KeyValue) []slog.Attr {
	if len(attrs) == 0 {
		return nil
	}

	result := make([]slog.Attr, 0, len(attrs))

	for _, attr := range attrs {
		result = append(result, otelKVToSlog(attr))
	}

	return result
}

// otelKVToSlog converts a single OTel attribute.KeyValue to an slog.Attr.
func otelKVToSlog(otelAttr attribute.KeyValue) slog.Attr {
	key := string(otelAttr.Key)

	switch otelAttr.Value.Type() {
	case attribute.STRING:
		return slog.String(key, otelAttr.Value.AsString())
	case attribute.INT64:
		return slog.Int64(key, otelAttr.Value.AsInt64())
	case attribute.FLOAT64:
		return slog.Float64(key, otelAttr.Value.AsFloat64())
	case attribute.BOOL:
		return slog.Bool(key, otelAttr.Value.AsBool())
	default:
		return slog.String(key, otelAttr.Value.Emit())
	}
}

// mergeContextAttrs prepends context-extracted attributes to the explicit attrs.
func mergeContextAttrs(ctx context.Context, explicit []attribute.KeyValue) []attribute.KeyValue {
	ctxAttrs := extractContextAttrs(ctx)
	if len(ctxAttrs) == 0 {
		return explicit
	}

	merged := make([]attribute.KeyValue, 0, len(ctxAttrs)+len(explicit))
	merged = append(merged, ctxAttrs...)
	merged = append(merged, explicit...)

	return merged
}

// emitLog writes a log line via slog.Default() at the given level.
func emitLog(
	ctx context.Context,
	level slog.Level,
	msg string,
	err error,
	attrs []attribute.KeyValue,
) {
	slogAttrs := otelToSlog(attrs)

	// For Error level, prepend the error as the first attribute.
	if err != nil {
		errAttr := slog.String("error", err.Error())
		slogAttrs = append([]slog.Attr{errAttr}, slogAttrs...)
	}

	args := make([]any, 0, len(slogAttrs))
	for _, a := range slogAttrs {
		args = append(args, a)
	}

	switch level {
	case slog.LevelInfo:
		slog.InfoContext(ctx, msg, args...)
	case slog.LevelWarn:
		slog.WarnContext(ctx, msg, args...)
	case slog.LevelError:
		slog.ErrorContext(ctx, msg, args...)
	default:
		slog.DebugContext(ctx, msg, args...)
	}
}

// emitSpanEvent adds a named event to the current span with merged context attrs.
func emitSpanEvent(ctx context.Context, msg string, attrs []attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	allAttrs := mergeContextAttrs(ctx, attrs)
	span.AddEvent(msg, trace.WithAttributes(allAttrs...))
}
