package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/metrics"
	"github.com/go-risk-it/go-risk-it/internal/web/middleware"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func setupOTelTest(t *testing.T) (*metrics.Metrics, *sdkmetric.ManualReader) {
	t.Helper()

	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	meter := provider.Meter("test")

	m, err := metrics.NewMetrics(meter)
	require.NoError(t, err)

	return m, reader
}

func collectErrorsTotal(
	t *testing.T,
	reader *sdkmetric.ManualReader,
) []metricdata.DataPoint[int64] {
	t.Helper()

	var rm metricdata.ResourceMetrics

	require.NoError(t, reader.Collect(t.Context(), &rm))

	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == "http.server.errors.total" {
				//nolint:forcetypeassert // test helper, we know the type
				return m.Data.(metricdata.Sum[int64]).DataPoints
			}
		}
	}

	return nil
}

func findDataPointWithCategory(
	points []metricdata.DataPoint[int64],
	category string,
) (metricdata.DataPoint[int64], bool) {
	for _, dp := range points {
		for _, attr := range dp.Attributes.ToSlice() {
			if attr.Key == "error.category" && attr.Value.AsString() == category {
				return dp, true
			}
		}
	}

	return metricdata.DataPoint[int64]{}, false
}

func TestOTelMiddleware_ErrorMetrics_400_ValidationError(t *testing.T) {
	t.Parallel()

	m, reader := setupOTelTest(t)
	otelMiddleware := middleware.NewOTelMiddleware(m)

	inner := route.New("/test", false, http.HandlerFunc(
		func(writer http.ResponseWriter, _ *http.Request) {
			writer.WriteHeader(http.StatusBadRequest)
		}))

	wrapped := otelMiddleware.Wrap(inner)

	logCtx := t.Context()
	req := httptest.NewRequestWithContext(logCtx, http.MethodPost, "/test", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	points := collectErrorsTotal(t, reader)
	require.NotEmpty(t, points)

	dp, found := findDataPointWithCategory(points, "VALIDATION_ERROR")
	require.True(t, found, "expected data point with error.category=VALIDATION_ERROR")
	assert.Equal(t, int64(1), dp.Value)
}

func TestOTelMiddleware_ErrorMetrics_500_InternalError(t *testing.T) {
	t.Parallel()

	m, reader := setupOTelTest(t)
	otelMiddleware := middleware.NewOTelMiddleware(m)

	inner := route.New("/test", false, http.HandlerFunc(
		func(writer http.ResponseWriter, _ *http.Request) {
			writer.WriteHeader(http.StatusInternalServerError)
		}))

	wrapped := otelMiddleware.Wrap(inner)

	logCtx := t.Context()
	req := httptest.NewRequestWithContext(logCtx, http.MethodPost, "/test", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	points := collectErrorsTotal(t, reader)
	require.NotEmpty(t, points)

	dp, found := findDataPointWithCategory(points, "INTERNAL_ERROR")
	require.True(t, found, "expected data point with error.category=INTERNAL_ERROR")
	assert.Equal(t, int64(1), dp.Value)
}

func TestOTelMiddleware_ErrorMetrics_200_NoErrorCounted(t *testing.T) {
	t.Parallel()

	m, reader := setupOTelTest(t)
	otelMiddleware := middleware.NewOTelMiddleware(m)

	inner := route.New("/test", false, http.HandlerFunc(
		func(writer http.ResponseWriter, _ *http.Request) {
			writer.WriteHeader(http.StatusOK)
		}))

	wrapped := otelMiddleware.Wrap(inner)

	logCtx := t.Context()
	req := httptest.NewRequestWithContext(logCtx, http.MethodPost, "/test", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	points := collectErrorsTotal(t, reader)
	assert.Empty(t, points, "no error metric should be recorded for 200")
}

func TestOTelMiddleware_ErrorMetrics_AllCategories(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		status   int
		category string
	}{
		{"400→VALIDATION_ERROR", http.StatusBadRequest, "VALIDATION_ERROR"},
		{"401→UNAUTHORIZED", http.StatusUnauthorized, "UNAUTHORIZED"},
		{"403→FORBIDDEN", http.StatusForbidden, "FORBIDDEN"},
		{"404→NOT_FOUND", http.StatusNotFound, "NOT_FOUND"},
		{"409→CONFLICT", http.StatusConflict, "CONFLICT"},
		{"500→INTERNAL_ERROR", http.StatusInternalServerError, "INTERNAL_ERROR"},
		{"503→INTERNAL_ERROR", http.StatusServiceUnavailable, "INTERNAL_ERROR"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			m, reader := setupOTelTest(t)
			otelMiddleware := middleware.NewOTelMiddleware(m)

			inner := route.New("/games", false, http.HandlerFunc(
				func(writer http.ResponseWriter, _ *http.Request) {
					writer.WriteHeader(test.status)
				}))

			wrapped := otelMiddleware.Wrap(inner)

			logCtx := t.Context()
			req := httptest.NewRequestWithContext(
				logCtx,
				http.MethodPost,
				"/games",
				nil,
			)
			rec := httptest.NewRecorder()

			wrapped.ServeHTTP(rec, req)

			points := collectErrorsTotal(t, reader)
			require.NotEmpty(t, points)

			dp, found := findDataPointWithCategory(points, test.category)
			require.True(t, found, "expected error.category=%s", test.category)
			assert.Equal(t, int64(1), dp.Value)

			// Verify route attribute
			for _, attr := range dp.Attributes.ToSlice() {
				if attr.Key == "http.route" {
					assert.Equal(t, "/games", attr.Value.AsString())
				}

				if attr.Key == "http.method" {
					assert.Equal(t, "POST", attr.Value.AsString())
				}
			}
		})
	}
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
