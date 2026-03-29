package bus

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/kernel/metrics"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
)

// SafeOp runs action with a child span and duration metric recording. On panic
// it records the error on the span and logs the recovered value and stack trace.
// This is a sequential wrapper (not a goroutine) — the bus already owns
// goroutine lifecycle. Passing nil for met is safe (metric recording is skipped).
func SafeOp(
	tracerName string,
	parent context.Context,
	name string,
	met *metrics.InfraMetrics,
	action func(),
) {
	ctx, span := otel.GetTracerProvider().
		Tracer(tracerName).
		Start(parent, "consumer."+name)
	defer span.End()

	start := time.Now()

	defer func() {
		elapsed := time.Since(start).Seconds()

		if met != nil {
			met.EventHandlerDuration.Record(ctx, elapsed,
				metric.WithAttributes(attribute.String("handler", name)))
		}

		if recovered := recover(); recovered != nil {
			span.RecordError(fmt.Errorf("panic in %s: %v", name, recovered))
			span.SetStatus(codes.Error, "panic")

			slog.ErrorContext(ctx, "panic in consumer operation",
				"operation", name,
				"error", recovered,
				"stack", string(debug.Stack()),
			)
		}
	}()

	action()
}
