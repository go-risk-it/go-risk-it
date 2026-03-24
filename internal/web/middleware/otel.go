package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/metrics"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	restutils "github.com/go-risk-it/go-risk-it/internal/web/rest/utils"
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
	isWebSocket := routeToWrap.Pattern() == "/ws"

	return route.New(
		routeToWrap.Pattern(),
		routeToWrap.RequiresAuth(),
		http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			spanName := fmt.Sprintf("%s %s", request.Method, routeToWrap.Pattern())

			_, span := m.tracer.Start(request.Context(), spanName,
				trace.WithAttributes(
					attribute.String("http.method", request.Method),
					attribute.String("http.route", routeToWrap.Pattern()),
				),
			)
			defer span.End()

			logContext, ok := request.Context().(ctx.LogContext)
			if !ok {
				_ = restutils.WriteError(
					writer,
					errors.New("invalid log context"),
				)

				return
			}

			traceContext := ctx.WithSpan(logContext, span)

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

			m.recordHTTPMetrics(request, routeToWrap.Pattern(), recorder.statusCode, duration)
		}),
	)
}

func (m *OTelMiddleware) recordHTTPMetrics(
	request *http.Request,
	pattern string,
	statusCode int,
	duration float64,
) {
	attrs := otelmetric.WithAttributes(
		attribute.String("http.method", request.Method),
		attribute.String("http.route", pattern),
		attribute.Int("http.status_code", statusCode),
	)

	m.metrics.HTTPRequestDuration.Record(request.Context(), duration, attrs)
	m.metrics.HTTPRequestsTotal.Add(request.Context(), 1, attrs)

	if statusCode >= http.StatusBadRequest {
		errorAttrs := otelmetric.WithAttributes(
			attribute.String("http.method", request.Method),
			attribute.String("http.route", pattern),
			attribute.String("error.category", StatusToCategory(statusCode)),
		)
		m.metrics.HTTPErrorsTotal.Add(request.Context(), 1, errorAttrs)
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
