package bus_test

import (
	"context"
	"testing"

	eventbus "github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	"github.com/go-risk-it/go-risk-it/internal/kernel/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// ---------------------------------------------------------------------------
// Test: Normal execution — action runs and span is created
// ---------------------------------------------------------------------------

//nolint:paralleltest // swaps global TracerProvider
func TestSafeOp_NormalExecution(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tracerProvider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	t.Cleanup(func() { _ = tracerProvider.Shutdown(context.Background()) })

	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(tracerProvider)
	t.Cleanup(func() { otel.SetTracerProvider(prev) })

	actionRan := false
	eventbus.SafeOp("test-tracer", context.Background(), "doWork", nil, func() {
		actionRan = true
	})

	require.True(t, actionRan, "action must run")

	stubs := exporter.GetSpans()
	var found bool

	for _, stub := range stubs {
		if stub.Name == "consumer.doWork" {
			found = true

			break
		}
	}

	require.True(t, found,
		"expected span named 'consumer.doWork', got: %v", spanNamesFromStubs(stubs))
}

// ---------------------------------------------------------------------------
// Test: Panic recovery — panic does not propagate, span records error
// ---------------------------------------------------------------------------

//nolint:paralleltest // swaps global TracerProvider
func TestSafeOp_PanicRecovery(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tracerProvider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	t.Cleanup(func() { _ = tracerProvider.Shutdown(context.Background()) })

	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(tracerProvider)
	t.Cleanup(func() { otel.SetTracerProvider(prev) })

	require.NotPanics(t, func() {
		eventbus.SafeOp("test-tracer", context.Background(), "explode", nil, func() {
			panic("boom")
		})
	})

	stubs := exporter.GetSpans()
	var panicSpan *tracetest.SpanStub

	for i := range stubs {
		if stubs[i].Name == "consumer.explode" {
			panicSpan = &stubs[i]

			break
		}
	}

	require.NotNil(t, panicSpan,
		"expected span named 'consumer.explode', got: %v", spanNamesFromStubs(stubs))

	// Span must record the error status.
	assert.Equal(t, codes.Error, panicSpan.Status.Code, "span must have Error status")
	assert.Equal(t, "panic", panicSpan.Status.Description, "span description must be 'panic'")

	// Span must have recorded the error event.
	require.NotEmpty(t, panicSpan.Events, "span must have at least one event (the error)")
}

// ---------------------------------------------------------------------------
// Test: Records duration metric
// ---------------------------------------------------------------------------

//nolint:paralleltest // swaps global TracerProvider
func TestSafeOp_RecordsDuration(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tracerProvider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	t.Cleanup(func() { _ = tracerProvider.Shutdown(context.Background()) })

	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(tracerProvider)
	t.Cleanup(func() { otel.SetTracerProvider(prev) })

	// Create a real metric reader so we can inspect recorded values.
	metricReader := sdkmetric.NewManualReader()
	meterProvider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(metricReader))
	t.Cleanup(func() { _ = meterProvider.Shutdown(context.Background()) })

	meter := meterProvider.Meter("test")

	met, err := metrics.NewInfraMetrics(meter)
	require.NoError(t, err)

	eventbus.SafeOp("test-tracer", context.Background(), "measuredOp", met, func() {
		// no-op — just records duration
	})

	// Collect and verify metrics.
	var resourceMetrics metricdata.ResourceMetrics
	require.NoError(t, metricReader.Collect(context.Background(), &resourceMetrics))

	// Find the event_handler.duration histogram.
	var foundMetric bool

	for _, scope := range resourceMetrics.ScopeMetrics {
		for _, m := range scope.Metrics {
			if m.Name == "event_handler.duration" {
				foundMetric = true

				break
			}
		}
	}

	require.True(t, foundMetric, "expected event_handler.duration metric to be recorded")
}

// ---------------------------------------------------------------------------
// Test: Nil metrics — no panic when InfraMetrics is nil
// ---------------------------------------------------------------------------

//nolint:paralleltest // swaps global TracerProvider
func TestSafeOp_NilMetrics(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tracerProvider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	t.Cleanup(func() { _ = tracerProvider.Shutdown(context.Background()) })

	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(tracerProvider)
	t.Cleanup(func() { otel.SetTracerProvider(prev) })

	require.NotPanics(t, func() {
		eventbus.SafeOp("test-tracer", context.Background(), "nilMetrics", nil, func() {
			// no-op
		})
	})

	stubs := exporter.GetSpans()
	var found bool

	for _, stub := range stubs {
		if stub.Name == "consumer.nilMetrics" {
			found = true

			break
		}
	}

	require.True(t, found,
		"expected span named 'consumer.nilMetrics', got: %v", spanNamesFromStubs(stubs))
}

// ---------------------------------------------------------------------------
// Test: Creates child span with correct tracer name
// ---------------------------------------------------------------------------

//nolint:paralleltest // swaps global TracerProvider
func TestSafeOp_CreatesChildSpan(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tracerProvider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	t.Cleanup(func() { _ = tracerProvider.Shutdown(context.Background()) })

	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(tracerProvider)
	t.Cleanup(func() { otel.SetTracerProvider(prev) })

	// Create a parent span.
	tracer := tracerProvider.Tracer("test-parent")
	parentCtx, parentSpan := tracer.Start(context.Background(), "parent-op")
	parentSpanCtx := parentSpan.SpanContext()

	eventbus.SafeOp("my-publisher-tracer", parentCtx, "childWork", nil, func() {
		// no-op
	})

	parentSpan.End()

	stubs := exporter.GetSpans()
	var childStub *tracetest.SpanStub

	for i := range stubs {
		if stubs[i].Name == "consumer.childWork" {
			childStub = &stubs[i]

			break
		}
	}

	require.NotNil(t, childStub,
		"expected span named 'consumer.childWork', got: %v", spanNamesFromStubs(stubs))

	// The child span must be in the same trace as the parent.
	assert.Equal(t, parentSpanCtx.TraceID(), childStub.SpanContext.TraceID(),
		"child span must be in the same trace as parent")

	// The child span's parent must be the parent span.
	assert.Equal(t, parentSpanCtx.SpanID(), childStub.Parent.SpanID(),
		"child span must have parent as its parent")

	// The tracer name (instrumentation scope) must match what was passed.
	assert.Equal(t, "my-publisher-tracer", childStub.InstrumentationScope.Name,
		"tracer name must match the tracerName parameter")
}

// spanNamesFromStubs extracts span names from stubs for diagnostic messages.
func spanNamesFromStubs(stubs tracetest.SpanStubs) []string {
	names := make([]string, len(stubs))
	for i, s := range stubs {
		names[i] = s.Name
	}

	return names
}
