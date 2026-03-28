package otel

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/config"
	"go.opentelemetry.io/contrib/instrumentation/host"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(NewLoggerProvider),
	fx.Invoke(SetupOTelSDK),
)

func NewLoggerProvider(otelConfig config.OtelConfig) (*sdklog.LoggerProvider, error) {
	return newLoggerProvider(otelConfig)
}

func SetupOTelSDK(
	lifecycle fx.Lifecycle,
	otelConfig config.OtelConfig,
	loggerProvider *sdklog.LoggerProvider,
) {
	// Set up OpenTelemetry.
	otelShutdown, err := setupOTelSDK(otelConfig)
	if err != nil {
		slog.Error("could not set up OpenTelemetry", "error", err)
		panic("could not set up OpenTelemetry: " + err.Error())
	}

	lifecycle.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			shutdownErr := errors.Join(
				otelShutdown(ctx),
				loggerProvider.Shutdown(ctx),
			)
			if shutdownErr != nil {
				slog.Error("failed to shutdown OTel providers", "error", shutdownErr)
				panic("failed to shutdown OTel providers: " + shutdownErr.Error())
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

	// runtime.NewProducer provides precomputed histogram metrics (go.schedule.duration)
	// that runtime.Start() alone does not emit.
	producer := runtime.NewProducer()

	return sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(
			sdkmetric.NewPeriodicReader(
				exporter,
				sdkmetric.WithInterval(10*time.Second),
				sdkmetric.WithProducer(producer),
			),
		),
	), nil
}

func newLoggerProvider(otelConfig config.OtelConfig) (*sdklog.LoggerProvider, error) {
	if !otelConfig.Enabled {
		// Dev fallback: stdout-only so logs are visible without LGTM stack.
		stdoutExporter, err := stdoutlog.New()
		if err != nil {
			return nil, fmt.Errorf("failed to create stdout log exporter: %w", err)
		}

		return sdklog.NewLoggerProvider(
			sdklog.WithProcessor(sdklog.NewSimpleProcessor(stdoutExporter)),
		), nil
	}

	// OTLP batch processor for Loki shipping.
	otlpExporter, err := otlploghttp.New(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP log exporter: %w", err)
	}

	// Stdout simple processor for dev visibility.
	stdoutExporter, err := stdoutlog.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout log exporter: %w", err)
	}

	return sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(otlpExporter)),
		sdklog.WithProcessor(sdklog.NewSimpleProcessor(stdoutExporter)),
	), nil
}
