package safego_test

import (
	"context"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/kernel/safego"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
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
	safego.SafeOp(context.Background(), "doWork", func(_ context.Context) error {
		actionRan = true

		return nil
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
		safego.SafeOp(context.Background(), "explode", func(_ context.Context) error {
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

	// Span must have recorded the error event.
	require.NotEmpty(t, panicSpan.Events, "span must have at least one event (the error)")
}

// ---------------------------------------------------------------------------
// Test: Creates child span with correct tracer name (go-risk-it via observe)
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

	safego.SafeOp(parentCtx, "childWork", func(_ context.Context) error {
		// no-op
		return nil
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

	// The tracer name (instrumentation scope) must match observe.RawSpan's constant.
	assert.Equal(t, "go-risk-it", childStub.InstrumentationScope.Name,
		"tracer name must match observe.RawSpan's constant (go-risk-it)")
}

// ---------------------------------------------------------------------------
// Test: Span carries handler attribute with operation name
// ---------------------------------------------------------------------------

//nolint:paralleltest // swaps global TracerProvider
func TestSafeOp_HandlerAttribute(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tracerProvider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	t.Cleanup(func() { _ = tracerProvider.Shutdown(context.Background()) })

	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(tracerProvider)
	t.Cleanup(func() { otel.SetTracerProvider(prev) })

	safego.SafeOp(context.Background(), "attrWork", func(_ context.Context) error {
		// no-op
		return nil
	})

	stubs := exporter.GetSpans()
	var stub *tracetest.SpanStub

	for i := range stubs {
		if stubs[i].Name == "consumer.attrWork" {
			stub = &stubs[i]

			break
		}
	}

	require.NotNil(t, stub,
		"expected span named 'consumer.attrWork', got: %v", spanNamesFromStubs(stubs))

	// The span must carry the handler attribute with the operation name.
	assert.Contains(t, stub.Attributes, attribute.String("handler", "attrWork"),
		"span must have handler attribute set to the operation name")
}

// ---------------------------------------------------------------------------
// Test: Error return — action error is recorded on the span
// ---------------------------------------------------------------------------

//nolint:paralleltest // swaps global TracerProvider
func TestSafeOp_ActionError_RecordsOnSpan(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tracerProvider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	t.Cleanup(func() { _ = tracerProvider.Shutdown(context.Background()) })

	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(tracerProvider)
	t.Cleanup(func() { otel.SetTracerProvider(prev) })

	safego.SafeOp(context.Background(), "failWork", func(_ context.Context) error {
		return assert.AnError
	})

	stubs := exporter.GetSpans()
	var stub *tracetest.SpanStub

	for i := range stubs {
		if stubs[i].Name == "consumer.failWork" {
			stub = &stubs[i]

			break
		}
	}

	require.NotNil(t, stub,
		"expected span named 'consumer.failWork', got: %v", spanNamesFromStubs(stubs))

	// Span must record the error status.
	assert.Equal(t, codes.Error, stub.Status.Code, "span must have Error status")

	// Span must have recorded the error event.
	require.NotEmpty(t, stub.Events, "span must have at least one event (the error)")
}

// spanNamesFromStubs extracts span names from stubs for diagnostic messages.
func spanNamesFromStubs(stubs tracetest.SpanStubs) []string {
	names := make([]string, len(stubs))
	for i, s := range stubs {
		names[i] = s.Name
	}

	return names
}
