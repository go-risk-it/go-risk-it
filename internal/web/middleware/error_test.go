package middleware_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	domainerrors "github.com/go-risk-it/go-risk-it/internal/logic/errors"
	"github.com/go-risk-it/go-risk-it/internal/web/middleware"
	restutils "github.com/go-risk-it/go-risk-it/internal/web/rest/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func newTestSpan(t *testing.T) context.Context {
	t.Helper()

	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	tracer := tp.Tracer("test")
	ctx, span := tracer.Start(t.Context(), "test-op")

	t.Cleanup(func() { span.End() })

	return ctx
}

func TestHandleErrors_NilError_NoOp(t *testing.T) {
	t.Parallel()

	handler := middleware.HandleErrors(func(writer http.ResponseWriter, _ *http.Request) error {
		writer.WriteHeader(http.StatusNoContent)

		return nil
	})

	recorder := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/test", nil)

	handler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusNoContent, recorder.Code)
}

func TestHandleErrors_ValidationError_Returns400WithTraceID(t *testing.T) {
	t.Parallel()

	ctx := newTestSpan(t)

	handler := middleware.HandleErrors(func(_ http.ResponseWriter, _ *http.Request) error {
		return domainerrors.NewValidationError("bad input")
	})

	recorder := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/test", nil)

	handler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var resp restutils.ErrorResponse
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &resp))
	assert.Equal(t, "bad input", resp.Error)
	assert.Equal(t, "VALIDATION_ERROR", resp.Code)
	assert.NotEmpty(t, resp.TraceID, "traceId should be present")
}

func TestHandleErrors_ConflictError_Returns409(t *testing.T) {
	t.Parallel()

	ctx := newTestSpan(t)

	handler := middleware.HandleErrors(func(_ http.ResponseWriter, _ *http.Request) error {
		return domainerrors.NewConflictError("wrong phase")
	})

	recorder := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/test", nil)

	handler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusConflict, recorder.Code)

	var resp restutils.ErrorResponse
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &resp))
	assert.Equal(t, "wrong phase", resp.Error)
	assert.Equal(t, "CONFLICT", resp.Code)
}

func TestHandleErrors_ForbiddenError_Returns403(t *testing.T) {
	t.Parallel()

	ctx := newTestSpan(t)

	handler := middleware.HandleErrors(func(_ http.ResponseWriter, _ *http.Request) error {
		return domainerrors.NewForbiddenError("not your turn")
	})

	recorder := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/test", nil)

	handler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusForbidden, recorder.Code)

	var resp restutils.ErrorResponse
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &resp))
	assert.Equal(t, "not your turn", resp.Error)
	assert.Equal(t, "FORBIDDEN", resp.Code)
}

func TestHandleErrors_NotFoundError_Returns404(t *testing.T) {
	t.Parallel()

	ctx := newTestSpan(t)

	handler := middleware.HandleErrors(func(_ http.ResponseWriter, _ *http.Request) error {
		return domainerrors.NewNotFoundError("game not found")
	})

	recorder := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/test", nil)

	handler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusNotFound, recorder.Code)

	var resp restutils.ErrorResponse
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &resp))
	assert.Equal(t, "game not found", resp.Error)
	assert.Equal(t, "NOT_FOUND", resp.Code)
}

func TestHandleErrors_GenericError_Returns500_NoMessageLeak(t *testing.T) {
	t.Parallel()

	ctx := newTestSpan(t)

	handler := middleware.HandleErrors(func(_ http.ResponseWriter, _ *http.Request) error {
		return errors.New("secret DB password: p4ssw0rd")
	})

	recorder := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/test", nil)

	handler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)

	var resp restutils.ErrorResponse
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &resp))
	assert.Equal(t, "an internal error occurred", resp.Error)
	assert.Equal(t, "INTERNAL_ERROR", resp.Code)
	assert.NotContains(t, recorder.Body.String(), "p4ssw0rd")
}

func TestHandleErrors_RecordsErrorOnSpan(t *testing.T) {
	t.Parallel()

	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	tracer := tp.Tracer("test")
	ctx, span := tracer.Start(t.Context(), "test-op")

	handler := middleware.HandleErrors(func(_ http.ResponseWriter, _ *http.Request) error {
		return domainerrors.NewValidationError("bad input")
	})

	recorder := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/test", nil)

	handler.ServeHTTP(recorder, req)
	span.End()

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)

	exportedSpan := spans[0]
	assert.Equal(t, codes.Error, exportedSpan.Status.Code)
	assert.Equal(t, "VALIDATION_ERROR", exportedSpan.Status.Description)
	require.NotEmpty(t, exportedSpan.Events)

	// Find the "exception" event (RecordError creates it)
	found := false
	for _, evt := range exportedSpan.Events {
		if evt.Name == "exception" {
			found = true

			break
		}
	}

	assert.True(t, found, "expected an 'exception' event from RecordError")
}

func TestHandleErrors_NoSpan_EmptyTraceID(t *testing.T) {
	t.Parallel()

	handler := middleware.HandleErrors(func(_ http.ResponseWriter, _ *http.Request) error {
		return domainerrors.NewValidationError("bad input")
	})

	recorder := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/test", nil)

	handler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &resp))
	_, hasTraceID := resp["traceId"]
	assert.False(t, hasTraceID, "traceId should be omitted when no span context")
}
