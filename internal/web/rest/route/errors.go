package route

import (
	"errors"
	"log/slog"
	"net/http"

	domainerrors "github.com/go-risk-it/go-risk-it/internal/kernel/errors"
	restutils "github.com/go-risk-it/go-risk-it/internal/web/rest/utils"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// WrapErrors wraps a PlainHandler into a standard http.HandlerFunc.
// On error it: extracts the trace ID from the span context, records the error on the span,
// sets the span status, and writes a JSON error response with the trace ID.
func WrapErrors(handler PlainHandler) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		err := handler(writer, request)
		if err == nil {
			return
		}

		span := trace.SpanFromContext(request.Context())
		span.RecordError(err)
		span.SetStatus(codes.Error, errorDescription(err))

		traceID := ""
		if span.SpanContext().HasTraceID() {
			traceID = span.SpanContext().TraceID().String()
		}

		if logErr := restutils.WriteErrorWithTrace(writer, err, traceID); logErr != nil {
			slog.ErrorContext(request.Context(), "request failed", "error", logErr)
		}
	}
}

func errorDescription(err error) string {
	var categorizable domainerrors.Categorizable
	if errors.As(err, &categorizable) {
		return categorizable.Category().String()
	}

	return "INTERNAL_ERROR"
}
