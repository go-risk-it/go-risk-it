package metrics

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/loadtest/health"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

const (
	otelFlushInterval = 5 * time.Second
	serviceName       = "perftest"
)

// healthCounter abstracts UpDownCounter for testability.
// Production code uses otelUpDownCounter (wraps metric.Int64UpDownCounter).
// Tests inject fakes via the interface.
type healthCounter interface {
	Add(delta int64)
}

// otelUpDownCounter adapts metric.Int64UpDownCounter to healthCounter.
type otelUpDownCounter struct {
	counter metric.Int64UpDownCounter
}

func (c *otelUpDownCounter) Add(delta int64) {
	c.counter.Add(context.Background(), delta)
}

// OTelExporter wraps OTel instruments for live export.
// Only resource-state gauges remain — all RED metrics (rate, errors, duration)
// are now derived from spans via the spanmetrics connector.
type OTelExporter struct {
	meterProvider  *sdkmetric.MeterProvider
	tracerProvider *sdktrace.TracerProvider

	// UpDown counters — resource state with no span equivalent.
	gamesActive metric.Int64UpDownCounter

	// Health classification gauges.
	healthHealthy healthCounter
	healthSlow    healthCounter
	healthStalled healthCounter
	healthZombie  healthCounter
	prevHealth    health.Distribution
}

// NewOTelExporter creates an OTel metric exporter pointing at the given OTLP HTTP endpoint.
func NewOTelExporter(ctx context.Context, endpoint string) (*OTelExporter, error) {
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion("0.1.0"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("create resource: %w", err)
	}

	// Set up propagator for trace context propagation across HTTP boundaries.
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Set up TracerProvider.
	tracerProvider, err := newTracerProvider(ctx, endpoint, res)
	if err != nil {
		return nil, err
	}

	otel.SetTracerProvider(tracerProvider)

	// Set up MeterProvider.
	meterProvider, err := newMeterProvider(ctx, endpoint, res)
	if err != nil {
		_ = tracerProvider.Shutdown(ctx)

		return nil, err
	}

	meter := meterProvider.Meter("perftest")

	o := &OTelExporter{meterProvider: meterProvider, tracerProvider: tracerProvider}
	if err := o.initInstruments(meter); err != nil {
		_ = tracerProvider.Shutdown(ctx)
		_ = meterProvider.Shutdown(ctx)

		return nil, err
	}

	return o, nil
}

func newTracerProvider(
	ctx context.Context,
	endpoint string,
	res *resource.Resource,
) (*sdktrace.TracerProvider, error) {
	traceExporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("create OTLP trace exporter: %w", err)
	}

	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
	), nil
}

func newMeterProvider(
	ctx context.Context,
	endpoint string,
	res *resource.Resource,
) (*sdkmetric.MeterProvider, error) {
	metricExporter, err := otlpmetrichttp.New(ctx,
		otlpmetrichttp.WithEndpoint(endpoint),
		otlpmetrichttp.WithInsecure(),
		otlpmetrichttp.WithTemporalitySelector(
			func(_ sdkmetric.InstrumentKind) metricdata.Temporality {
				return metricdata.CumulativeTemporality
			},
		),
	)
	if err != nil {
		return nil, fmt.Errorf("create OTLP metric exporter: %w", err)
	}

	return sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter,
			sdkmetric.WithInterval(otelFlushInterval),
		)),
		sdkmetric.WithResource(res),
	), nil
}

func (o *OTelExporter) initInstruments(meter metric.Meter) error {
	var err error

	if o.gamesActive, err = meter.Int64UpDownCounter("perftest.games.active",
		metric.WithDescription("Currently running games"),
	); err != nil {
		return err
	}

	healthHealthy, err := meter.Int64UpDownCounter("perftest.health.healthy",
		metric.WithDescription("Games classified as healthy"),
	)
	if err != nil {
		return err
	}

	o.healthHealthy = &otelUpDownCounter{counter: healthHealthy}

	healthSlow, err := meter.Int64UpDownCounter("perftest.health.slow",
		metric.WithDescription("Games classified as slow"),
	)
	if err != nil {
		return err
	}

	o.healthSlow = &otelUpDownCounter{counter: healthSlow}

	healthStalled, err := meter.Int64UpDownCounter("perftest.health.stalled",
		metric.WithDescription("Games classified as stalled"),
	)
	if err != nil {
		return err
	}

	o.healthStalled = &otelUpDownCounter{counter: healthStalled}

	healthZombie, err := meter.Int64UpDownCounter("perftest.health.zombie",
		metric.WithDescription("Games classified as zombie"),
	)
	if err != nil {
		return err
	}

	o.healthZombie = &otelUpDownCounter{counter: healthZombie}

	return nil
}

// RecordHealthDistribution reports health classification changes to OTel.
// It computes deltas from the previous snapshot and applies them as UpDownCounter
// additions. Not concurrent-safe — intended for use from the single-threaded hold loop.
func (o *OTelExporter) RecordHealthDistribution(dist health.Distribution) {
	o.healthHealthy.Add(int64(dist.Healthy - o.prevHealth.Healthy))
	o.healthSlow.Add(int64(dist.Slow - o.prevHealth.Slow))
	o.healthStalled.Add(int64(dist.Stalled - o.prevHealth.Stalled))
	o.healthZombie.Add(int64(dist.Zombie - o.prevHealth.Zombie))

	o.prevHealth = dist
}

// ResetHealthCounters zeroes the health counters by emitting negative deltas,
// then clears prevHealth. Call between staircase steps so counters start fresh.
func (o *OTelExporter) ResetHealthCounters() {
	o.healthHealthy.Add(int64(-o.prevHealth.Healthy))
	o.healthSlow.Add(int64(-o.prevHealth.Slow))
	o.healthStalled.Add(int64(-o.prevHealth.Stalled))
	o.healthZombie.Add(int64(-o.prevHealth.Zombie))

	o.prevHealth = health.Distribution{}
}

// Shutdown flushes and stops both the meter and tracer providers.
func (o *OTelExporter) Shutdown(ctx context.Context) error {
	return errors.Join(
		o.meterProvider.Shutdown(ctx),
		o.tracerProvider.Shutdown(ctx),
	)
}
