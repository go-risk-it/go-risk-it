package middleware

import (
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

type OTelMiddleware interface {
	Middleware
}

type OTelMiddlewareImpl struct {
	tracer  trace.Tracer
	metrics *metrics.Metrics
}

var _ OTelMiddleware = (*OTelMiddlewareImpl)(nil)

func NewOTelMiddleware(metrics *metrics.Metrics) OTelMiddleware {
	return &OTelMiddlewareImpl{
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

func (m *OTelMiddlewareImpl) Wrap(routeToWrap route.Route) route.Route {
	isWebSocket := routeToWrap.Pattern() == "/ws"

	return route.NewRoute(
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
				http.Error(writer, "invalid log context", http.StatusInternalServerError)

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
			attrs := otelmetric.WithAttributes(
				attribute.String("http.method", request.Method),
				attribute.String("http.route", routeToWrap.Pattern()),
				attribute.Int("http.status_code", recorder.statusCode),
			)

			m.metrics.HTTPRequestDuration.Record(request.Context(), duration, attrs)
			m.metrics.HTTPRequestsTotal.Add(request.Context(), 1, attrs)
		}),
	)
}
