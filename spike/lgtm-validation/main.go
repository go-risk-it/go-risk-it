// spike/lgtm-validation validates LGTM stack telemetry assumptions:
//   - Exemplar attachment on histogram recordings with traced context
//   - OTel log records with custom attributes via otelslog bridge
//   - Trace propagation with parent/child spans
//
// Run against a live LGTM stack on localhost:4318.
package main

import (
	"context"
	"log"
	"log/slog"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/metric"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	serviceName    = "spike-lgtm-validation"
	serviceVersion = "0.1.0"
	endpoint       = "localhost:4318"
)

func main() {
	ctx := context.Background()

	res := newResource(ctx)
	tracerProvider := newTracerProvider(ctx, res)
	meterProvider := newMeterProvider(ctx, res)
	loggerProvider := newLoggerProvider(ctx, res)

	defer shutdown(ctx, tracerProvider, meterProvider, loggerProvider)

	// --- Trace: parent span + child span ---
	tracer := tracerProvider.Tracer(serviceName)
	parentCtx, parentSpan := tracer.Start(ctx, "spike.validate",
		trace.WithSpanKind(trace.SpanKindInternal),
	)

	_, childSpan := tracer.Start(parentCtx, "spike.validate.child")
	childSpan.AddEvent("child-work-done")
	childSpan.End()

	// --- Metric: histogram with traced context (exemplar attachment) ---
	meter := meterProvider.Meter(serviceName)
	histogram, err := meter.Float64Histogram("spike.request.duration",
		metric.WithDescription("Test histogram for exemplar validation"),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(0.01, 0.05, 0.1, 0.25, 0.5, 1.0),
	)
	if err != nil {
		log.Fatal("create histogram:", err)
	}

	// Record with the parent span context — exemplar should attach the trace ID.
	histogram.Record(parentCtx, 0.042)
	histogram.Record(parentCtx, 0.137)
	histogram.Record(parentCtx, 0.003)

	parentSpan.End()

	// --- Log: otelslog bridge with custom attributes ---
	logger := slog.New(
		otelslog.NewHandler(serviceName, otelslog.WithLoggerProvider(loggerProvider)),
	)

	// Log within the traced context so traceID propagates to the log record.
	logger.InfoContext(parentCtx, "spike validation log",
		slog.Int("gameID", 42),
		slog.String("userID", "test-user"),
	)
	logger.WarnContext(parentCtx, "spike warning with trace context",
		slog.Int("gameID", 42),
		slog.String("userID", "test-user"),
		slog.String("detail", "this tests structured metadata in Loki"),
	)

	log.Println("telemetry sent — check Grafana (Tempo, Prometheus, Loki)")
}

func newResource(ctx context.Context) *resource.Resource {
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
		),
	)
	if err != nil {
		log.Fatal("create resource:", err)
	}

	return res
}

func newTracerProvider(ctx context.Context, res *resource.Resource) *sdktrace.TracerProvider {
	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		log.Fatal("create trace exporter:", err)
	}

	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
}

func newMeterProvider(ctx context.Context, res *resource.Resource) *sdkmetric.MeterProvider {
	exporter, err := otlpmetrichttp.New(ctx,
		otlpmetrichttp.WithEndpoint(endpoint),
		otlpmetrichttp.WithInsecure(),
	)
	if err != nil {
		log.Fatal("create metric exporter:", err)
	}

	return sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter,
			sdkmetric.WithInterval(time.Second),
		)),
		sdkmetric.WithResource(res),
	)
}

func newLoggerProvider(ctx context.Context, res *resource.Resource) *sdklog.LoggerProvider {
	exporter, err := otlploghttp.New(ctx,
		otlploghttp.WithEndpoint(endpoint),
		otlploghttp.WithInsecure(),
	)
	if err != nil {
		log.Fatal("create log exporter:", err)
	}

	return sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
		sdklog.WithResource(res),
	)
}

func shutdown(
	ctx context.Context,
	tp *sdktrace.TracerProvider,
	mp *sdkmetric.MeterProvider,
	lp *sdklog.LoggerProvider,
) {
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := tp.Shutdown(shutdownCtx); err != nil {
		log.Println("trace provider shutdown:", err)
	}

	if err := mp.Shutdown(shutdownCtx); err != nil {
		log.Println("meter provider shutdown:", err)
	}

	if err := lp.Shutdown(shutdownCtx); err != nil {
		log.Println("logger provider shutdown:", err)
	}
}
