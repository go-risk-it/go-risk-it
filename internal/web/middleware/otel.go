package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/metrics"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelmetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type OTelMiddleware struct {
	tracer  trace.Tracer
	metrics *metrics.Metrics
}

func NewOTelMiddleware(metrics *metrics.Metrics) *OTelMiddleware {
	return &OTelMiddleware{
		tracer:  otel.GetTracerProvider().Tracer("go-risk-it-http"),
		metrics: metrics,
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
		spanName := fmt.Sprintf("%s %s", request.Method, routeToWrap.Pattern())

		tracedCtx, span := m.tracer.Start(request.Context(), spanName,
			trace.WithAttributes(
				attribute.String("http.method", request.Method),
				attribute.String("http.route", routeToWrap.Pattern()),
			),
		)
		defer span.End()

		traceContext := ctx.WithSpan(tracedCtx, span)

		// WebSocket routes need the raw response writer for nbio's upgrade.
		// Skip status recording and HTTP metrics for WS connections.
		if isWebSocket {
			routeToWrap.ServeHTTP(writer, request.WithContext(traceContext))

			return
		}

		recorder := &statusRecorder{ResponseWriter: writer, statusCode: http.StatusOK}
		start := time.Now()

		routeToWrap.ServeHTTP(
			recorder,
			request.WithContext(traceContext),
		)

		duration := time.Since(start).Seconds()

		m.recordHTTPMetrics(
			tracedCtx,
			request.Method,
			routeToWrap.Pattern(),
			recorder.statusCode,
			duration,
		)
	})

	return routeToWrap.Wrap(handler)
}

func (m *OTelMiddleware) recordHTTPMetrics(
	ctx context.Context,
	method string,
	pattern string,
	statusCode int,
	duration float64,
) {
	attrs := otelmetric.WithAttributes(
		attribute.String("http.method", method),
		attribute.String("http.route", pattern),
		attribute.Int("http.status_code", statusCode),
	)

	m.metrics.HTTPRequestDuration.Record(ctx, duration, attrs)
	m.metrics.HTTPRequestsTotal.Add(ctx, 1, attrs)

	if statusCode >= http.StatusBadRequest {
		errorAttrs := otelmetric.WithAttributes(
			attribute.String("http.method", method),
			attribute.String("http.route", pattern),
			attribute.String("error.category", StatusToCategory(statusCode)),
		)
		m.metrics.HTTPErrorsTotal.Add(ctx, 1, errorAttrs)
	}
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
