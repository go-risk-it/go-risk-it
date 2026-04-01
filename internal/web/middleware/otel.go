package middleware

import (
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type OTelMiddleware struct {
	tracer trace.Tracer
}

func NewOTelMiddleware() *OTelMiddleware {
	return &OTelMiddleware{
		tracer: otel.GetTracerProvider().Tracer("go-risk-it-http"),
	}
}

type statusRecorder struct {
	http.ResponseWriter

	statusCode int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func (m *OTelMiddleware) Wrap(routeToWrap *route.Route) *route.Route {
	isWebSocket := routeToWrap.IsWebSocket()

	handler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		spanName := routeToWrap.Pattern()

		tracedCtx, span := m.tracer.Start(request.Context(), spanName,
			trace.WithAttributes(
				attribute.String("http_method", request.Method),
				attribute.String("http_route", routeToWrap.Pattern()),
			),
		)
		defer span.End()

		traceContext := ctx.WithSpan(tracedCtx, span)

		// WebSocket routes need the raw response writer for nbio's upgrade.
		// Skip status recording for WS connections.
		if isWebSocket {
			routeToWrap.ServeHTTP(writer, request.WithContext(traceContext))

			return
		}

		recorder := &statusRecorder{ResponseWriter: writer, statusCode: http.StatusOK}

		routeToWrap.ServeHTTP(
			recorder,
			request.WithContext(traceContext),
		)

		span.SetAttributes(attribute.Int("http_status_code", recorder.statusCode))
		if recorder.statusCode >= http.StatusBadRequest {
			span.SetAttributes(
				attribute.String("error_category", StatusToCategory(recorder.statusCode)),
			)
		}
	})

	return routeToWrap.Wrap(handler)
}

func StatusToCategory(code int) string {
	switch code {
	case http.StatusBadRequest:
		return "VALIDATION_ERROR"
	case http.StatusUnauthorized:
		return "UNAUTHORIZED"
	case http.StatusForbidden:
		return "FORBIDDEN"
	case http.StatusNotFound:
		return "NOT_FOUND"
	case http.StatusConflict:
		return "CONFLICT"
	default:
		return "INTERNAL_ERROR"
	}
}
