package metrics

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

const (
	otelFlushInterval = 5 * time.Second
	serviceName       = "perftest"
)

// OTelExporter wraps OTel instruments that mirror the HDR histogram collector.
// Each Record*() call on the Collector also records to these instruments for live export.
type OTelExporter struct {
	provider *sdkmetric.MeterProvider

	// Counters.
	movesTotal      metric.Int64Counter
	errorsTotal     metric.Int64Counter
	gamesCompleted  metric.Int64Counter
	gamesTimedOut   metric.Int64Counter
	gamesFatal      metric.Int64Counter
	conflictsTotal  metric.Int64Counter
	retriesTotal    metric.Int64Counter
	reconnectsTotal metric.Int64Counter

	// UpDown counters.
	gamesActive metric.Int64UpDownCounter

	// Histograms.
	restDuration metric.Float64Histogram
	e2eDuration  metric.Float64Histogram
	wsDuration   metric.Float64Histogram
}

// NewOTelExporter creates an OTel metric exporter pointing at the given OTLP HTTP endpoint.
func NewOTelExporter(ctx context.Context, endpoint string) (*OTelExporter, error) {
	exporter, err := otlpmetrichttp.New(ctx,
		otlpmetrichttp.WithEndpoint(endpoint),
		otlpmetrichttp.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("create OTLP metric exporter: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion("0.1.0"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("create resource: %w", err)
	}

	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter,
			sdkmetric.WithInterval(otelFlushInterval),
		)),
		sdkmetric.WithResource(res),
	)

	meter := provider.Meter("perftest")

	o := &OTelExporter{provider: provider}
	if err := o.initInstruments(meter); err != nil {
		_ = provider.Shutdown(ctx)

		return nil, err
	}

	return o, nil
}

func (o *OTelExporter) initInstruments(meter metric.Meter) error {
	var err error

	if o.movesTotal, err = meter.Int64Counter("perftest.moves.total",
		metric.WithDescription("Total moves performed by load test"),
	); err != nil {
		return err
	}

	if o.errorsTotal, err = meter.Int64Counter("perftest.errors.total",
		metric.WithDescription("Total errors by type"),
	); err != nil {
		return err
	}

	if o.gamesCompleted, err = meter.Int64Counter("perftest.games.completed",
		metric.WithDescription("Games completed successfully"),
	); err != nil {
		return err
	}

	if o.gamesTimedOut, err = meter.Int64Counter("perftest.games.timed_out",
		metric.WithDescription("Games that timed out"),
	); err != nil {
		return err
	}

	if o.gamesFatal, err = meter.Int64Counter("perftest.games.fatal",
		metric.WithDescription("Games with fatal errors"),
	); err != nil {
		return err
	}

	if o.conflictsTotal, err = meter.Int64Counter("perftest.conflicts.total",
		metric.WithDescription("Total 409 conflict responses"),
	); err != nil {
		return err
	}

	if o.retriesTotal, err = meter.Int64Counter("perftest.retries.total",
		metric.WithDescription("Total REST retries"),
	); err != nil {
		return err
	}

	if o.reconnectsTotal, err = meter.Int64Counter("perftest.reconnects.total",
		metric.WithDescription("Total WebSocket reconnection attempts"),
	); err != nil {
		return err
	}

	if o.gamesActive, err = meter.Int64UpDownCounter("perftest.games.active",
		metric.WithDescription("Currently running games"),
	); err != nil {
		return err
	}

	if o.restDuration, err = meter.Float64Histogram("perftest.rest.duration",
		metric.WithDescription("REST API call duration"),
		metric.WithUnit("s"),
	); err != nil {
		return err
	}

	if o.e2eDuration, err = meter.Float64Histogram("perftest.e2e.duration",
		metric.WithDescription("End-to-end move duration"),
		metric.WithUnit("s"),
	); err != nil {
		return err
	}

	if o.wsDuration, err = meter.Float64Histogram("perftest.ws.delivery.duration",
		metric.WithDescription("WebSocket delivery latency"),
		metric.WithUnit("s"),
	); err != nil {
		return err
	}

	return nil
}

// Shutdown flushes and stops the OTel provider.
func (o *OTelExporter) Shutdown(ctx context.Context) error {
	return o.provider.Shutdown(ctx)
}
