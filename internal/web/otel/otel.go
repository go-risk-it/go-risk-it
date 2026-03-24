package otel

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/config"
	"go.opentelemetry.io/contrib/instrumentation/host"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Options(
	fx.Invoke(SetupOTelSDK),
)

func SetupOTelSDK(
	lifecycle fx.Lifecycle,
	log *zap.SugaredLogger,
	otelConfig config.OtelConfig,
) {
	// Set up OpenTelemetry.
	otelShutdown, err := setupOTelSDK(otelConfig)
	if err != nil {
		log.Fatalw("could not set up OpenTelemetry", "error", err)
	}

	lifecycle.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			if err := otelShutdown(ctx); err != nil {
				log.Fatalw("failed to shutdown OTel providers", "error", err)
			}

			return nil
		},
	})
}

func setupOTelSDK(otelConfig config.OtelConfig) (func(context.Context) error, error) {
	shutdownFuncs := make([]func(context.Context) error, 0, 2)

	shutdown := func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}

		shutdownFuncs = nil

		return err
	}

	// Set up propagator.
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	// Set up trace provider.
	tracerProvider, err := newTraceProvider(otelConfig)
	if err != nil {
		return shutdown, fmt.Errorf("failed to setup trace provider: %w", err)
	}

	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	// Set up meter provider.
	meterProvider, err := newMeterProvider(otelConfig)
	if err != nil {
		return shutdown, fmt.Errorf("failed to setup meter provider: %w", err)
	}

	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)

	// Start Go runtime metrics (goroutines, GC, heap, allocations).
	if otelConfig.Enabled {
		if err := runtime.Start(runtime.WithMinimumReadMemStatsInterval(time.Second)); err != nil {
			return shutdown, fmt.Errorf("failed to start runtime instrumentation: %w", err)
		}

		if err := host.Start(); err != nil {
			return shutdown, fmt.Errorf("failed to start host instrumentation: %w", err)
		}
	}

	return shutdown, nil
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newTraceProvider(otelConfig config.OtelConfig) (*trace.TracerProvider, error) {
	exporter, err := otlptracehttp.New(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	if otelConfig.Enabled {
		return trace.NewTracerProvider(trace.WithBatcher(exporter)), nil
	}

	return trace.NewTracerProvider(), nil
}

func newMeterProvider(otelConfig config.OtelConfig) (*sdkmetric.MeterProvider, error) {
	if !otelConfig.Enabled {
		return sdkmetric.NewMeterProvider(), nil
	}

	exporter, err := otlpmetrichttp.New(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to create metric exporter: %w", err)
	}

	return sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(
			sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(10*time.Second)),
		),
	), nil
}
