package middleware

import (
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type OTelMiddleware interface {
	Middleware
}

type OTelMiddlewareImpl struct {
	tracer trace.Tracer
}

var _ OTelMiddleware = (*OTelMiddlewareImpl)(nil)

func NewOTelMiddleware() OTelMiddleware {
	return &OTelMiddlewareImpl{
		tracer: otel.GetTracerProvider().Tracer("go-risk-it-http"),
	}
}

func (m *OTelMiddlewareImpl) Wrap(routeToWrap route.Route) route.Route {
	return route.NewRoute(
		routeToWrap.Pattern(),
		routeToWrap.RequiresAuth(),
		http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			_, span := m.tracer.Start(request.Context(), "roll")
			defer span.End()

			logContext, ok := request.Context().(ctx.LogContext)
			if !ok {
				http.Error(writer, "invalid log context", http.StatusInternalServerError)

				return
			}

			traceContext := ctx.WithSpan(logContext, span)
			routeToWrap.ServeHTTP(
				writer,
				request.WithContext(traceContext),
			)
		}),
	)
}
