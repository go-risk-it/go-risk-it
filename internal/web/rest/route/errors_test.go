package route_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	domainerrors "github.com/go-risk-it/go-risk-it/internal/kernel/errors"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	restutils "github.com/go-risk-it/go-risk-it/internal/web/rest/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func newTracedContext(t *testing.T) context.Context {
	t.Helper()

	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	tracer := tp.Tracer("test")
	spanCtx, span := tracer.Start(t.Context(), "test-op")

	t.Cleanup(func() { span.End() })

	return spanCtx
}

func TestWrapErrors_NilError_NoOp(t *testing.T) {
	t.Parallel()

	handler := route.WrapErrors(func(w http.ResponseWriter, _ *http.Request) error {
		w.WriteHeader(http.StatusNoContent)

		return nil
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/test", nil)

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestWrapErrors_ValidationError_Returns400WithTraceID(t *testing.T) {
	t.Parallel()

	spanCtx := newTracedContext(t)

	handler := route.WrapErrors(func(_ http.ResponseWriter, _ *http.Request) error {
		return domainerrors.NewValidationError("bad input")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(spanCtx, http.MethodPost, "/test", nil)

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp restutils.ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "bad input", resp.Error)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
	assert.NotEmpty(t, resp.TraceID)
}

func TestWrapErrors_ConflictError_Returns409(t *testing.T) {
	t.Parallel()

	spanCtx := newTracedContext(t)

	handler := route.WrapErrors(func(_ http.ResponseWriter, _ *http.Request) error {
		return domainerrors.NewConflictError("wrong phase")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(spanCtx, http.MethodPost, "/test", nil)

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)

	var resp restutils.ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "wrong phase", resp.Error)
	assert.Equal(t, "CONFLICT", resp.Code)
}

func TestWrapErrors_InternalError_NoMessageLeak(t *testing.T) {
	t.Parallel()

	spanCtx := newTracedContext(t)

	handler := route.WrapErrors(func(_ http.ResponseWriter, _ *http.Request) error {
		return errors.New("secret: p4ssw0rd")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(spanCtx, http.MethodPost, "/test", nil)

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var resp restutils.ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "an internal error occurred", resp.Error)
	assert.NotContains(t, rec.Body.String(), "p4ssw0rd")
}

func TestWrapErrors_RecordsErrorOnSpan(t *testing.T) {
	t.Parallel()

	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	tracer := tp.Tracer("test")
	spanCtx, span := tracer.Start(t.Context(), "test-op")

	handler := route.WrapErrors(func(_ http.ResponseWriter, _ *http.Request) error {
		return domainerrors.NewValidationError("bad input")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(spanCtx, http.MethodPost, "/test", nil)

	handler.ServeHTTP(rec, req)
	span.End()

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)
	assert.Equal(t, codes.Error, spans[0].Status.Code)
	assert.Equal(t, "VALIDATION_ERROR", spans[0].Status.Description)
}

func TestWrapErrors_NoSpan_EmptyTraceID(t *testing.T) {
	t.Parallel()

	handler := route.WrapErrors(func(_ http.ResponseWriter, _ *http.Request) error {
		return domainerrors.NewValidationError("bad input")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/test", nil)

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	_, hasTraceID := resp["traceId"]
	assert.False(t, hasTraceID, "traceId should be omitted when no span context")
}
