package otel

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
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
				log.Fatalw("failed to shutdown tracer provider", "error", err)
			}

			return nil
		},
	})
}

func setupOTelSDK(otelConfig config.OtelConfig) (func(context.Context) error, error) {
	var shutdownFuncs []func(context.Context) error

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
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
		return nil, err
	}

	if otelConfig.Enabled {
		return trace.NewTracerProvider(trace.WithBatcher(exporter)), nil
	} else {
		return trace.NewTracerProvider(), nil
	}
}

// func newMeterProvider() (*metric.MeterProvider, error) {
//	metricExporter, err := stdoutmetric.SetupOtelSDK()
//	if err != nil {
//		return nil, err
//	}
//
//	meterProvider := metric.NewMeterProvider(
//		metric.WithReader(metric.NewPeriodicReader(metricExporter,
//			// Default is 1m. Set to 3s for demonstrative purposes.
//			metric.WithInterval(3*time.Second))),
//	)
//	return meterProvider, nil
//}

// func newLoggerProvider(logger *zap.SugaredLogger) (*log.LoggerProvider, error) {
//	logExporter, err := stdoutlog.SetupOtelSDK()
//	if err != nil {
//		return nil, err
//	}
//
//	loggerProvider := log.NewLoggerProvider(
//		log.WithProcessor(log.NewBatchProcessor(logExporter)),
//	)
//	return loggerProvider, nil
//}
