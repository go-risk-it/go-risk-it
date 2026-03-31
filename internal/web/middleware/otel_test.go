package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/web/middleware"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

//nolint:paralleltest // swaps global TracerProvider
func TestOTelMiddleware_SpanAttributes_Success(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tracerProvider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	t.Cleanup(func() { _ = tracerProvider.Shutdown(t.Context()) })

	previous := otel.GetTracerProvider()
	otel.SetTracerProvider(tracerProvider)
	t.Cleanup(func() { otel.SetTracerProvider(previous) })

	otelMiddleware := middleware.NewOTelMiddleware()

	inner := route.NewForTest("/test", false, http.HandlerFunc(
		func(writer http.ResponseWriter, _ *http.Request) {
			writer.WriteHeader(http.StatusOK)
		}))

	wrapped := otelMiddleware.Wrap(inner)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)

	stub := spans[0]
	assert.True(t,
		hasSpanAttr(stub.Attributes, "http_status_code", attribute.IntValue(http.StatusOK)),
		"span must carry http_status_code=200",
	)
	assert.False(t,
		spanHasKey(stub.Attributes, "error_category"),
		"success span must NOT carry error_category",
	)
}

//nolint:paralleltest // swaps global TracerProvider
func TestOTelMiddleware_SpanAttributes_Error(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		category string
	}{
		{"400→VALIDATION_ERROR", http.StatusBadRequest, "VALIDATION_ERROR"},
		{"401→UNAUTHORIZED", http.StatusUnauthorized, "UNAUTHORIZED"},
		{"404→NOT_FOUND", http.StatusNotFound, "NOT_FOUND"},
		{"500→INTERNAL_ERROR", http.StatusInternalServerError, "INTERNAL_ERROR"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			exporter := tracetest.NewInMemoryExporter()
			tracerProvider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
			t.Cleanup(func() { _ = tracerProvider.Shutdown(t.Context()) })

			previous := otel.GetTracerProvider()
			otel.SetTracerProvider(tracerProvider)
			t.Cleanup(func() { otel.SetTracerProvider(previous) })

			otelMiddleware := middleware.NewOTelMiddleware()

			inner := route.NewForTest("/test", false, http.HandlerFunc(
				func(writer http.ResponseWriter, _ *http.Request) {
					writer.WriteHeader(test.status)
				}))

			wrapped := otelMiddleware.Wrap(inner)

			req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/test", nil)
			rec := httptest.NewRecorder()

			wrapped.ServeHTTP(rec, req)

			spans := exporter.GetSpans()
			require.Len(t, spans, 1)

			stub := spans[0]
			assert.True(t,
				hasSpanAttr(stub.Attributes, "http_status_code", attribute.IntValue(test.status)),
				"span must carry http_status_code=%d", test.status,
			)
			assert.True(
				t,
				hasSpanAttr(
					stub.Attributes,
					"error_category",
					attribute.StringValue(test.category),
				),
				"span must carry error_category=%s",
				test.category,
			)
		})
	}
}

//nolint:paralleltest // swaps global TracerProvider
func TestOTelMiddleware_WebSocket_SkipsStatusRecording(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tracerProvider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	t.Cleanup(func() { _ = tracerProvider.Shutdown(t.Context()) })

	previous := otel.GetTracerProvider()
	otel.SetTracerProvider(tracerProvider)
	t.Cleanup(func() { otel.SetTracerProvider(previous) })

	otelMiddleware := middleware.NewOTelMiddleware()

	inner := route.NewWebSocketForTest("GET /api/v1/games/{id}/ws", true, http.HandlerFunc(
		func(writer http.ResponseWriter, _ *http.Request) {
			writer.WriteHeader(http.StatusOK)
		}))

	wrapped := otelMiddleware.Wrap(inner)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/games/1/ws", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)

	// WebSocket route should not carry http_status_code
	assert.False(t,
		spanHasKey(spans[0].Attributes, "http_status_code"),
		"WebSocket span should not carry http_status_code",
	)
}

func hasSpanAttr(attrs []attribute.KeyValue, key string, expected attribute.Value) bool {
	for _, a := range attrs {
		if string(a.Key) == key && a.Value == expected {
			return true
		}
	}

	return false
}

func spanHasKey(attrs []attribute.KeyValue, key string) bool {
	for _, a := range attrs {
		if string(a.Key) == key {
			return true
		}
	}

	return false
}

func TestStatusToCategory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		status   int
		expected string
	}{
		{400, "VALIDATION_ERROR"},
		{401, "UNAUTHORIZED"},
		{403, "FORBIDDEN"},
		{404, "NOT_FOUND"},
		{409, "CONFLICT"},
		{500, "INTERNAL_ERROR"},
		{502, "INTERNAL_ERROR"},
		{503, "INTERNAL_ERROR"},
		{418, "INTERNAL_ERROR"},
	}

	for _, test := range tests {
		assert.Equal(
			t,
			test.expected,
			middleware.StatusToCategory(test.status),
			"status %d", test.status,
		)
	}
}
