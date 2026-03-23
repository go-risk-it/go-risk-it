package middleware

import (
	"errors"
	"net/http"

	domainerrors "github.com/go-risk-it/go-risk-it/internal/logic/errors"
	restutils "github.com/go-risk-it/go-risk-it/internal/web/rest/utils"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// ErrorHandlerFunc is an HTTP handler that returns an error instead of writing it directly.
// When wrapped with HandleErrors, the error is mapped to an appropriate HTTP response.
type ErrorHandlerFunc func(w http.ResponseWriter, r *http.Request) error

// HandleErrors wraps an ErrorHandlerFunc into a standard http.HandlerFunc.
// On error it: extracts the trace ID from the span context, records the error on the span,
// sets the span status, and writes a JSON error response with the trace ID.
// On nil error it is a no-op (the handler already wrote the response).
func HandleErrors(handler ErrorHandlerFunc) http.HandlerFunc {
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
			logFromContext(request, logErr)
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

func logFromContext(r *http.Request, err error) {
	type logger interface {
		Errorw(msg string, keysAndValues ...any)
	}

	if l, ok := r.Context().(logger); ok {
		l.Errorw("request failed", "error", err)
	}
}
