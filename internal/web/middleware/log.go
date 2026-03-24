package middleware

import (
	"log/slog"
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
)

type LogMiddleware struct{}

func NewLogMiddleware() *LogMiddleware {
	return &LogMiddleware{}
}

func (m *LogMiddleware) Wrap(routeToWrap *route.Route) *route.Route {
	return route.New(
		routeToWrap.Pattern(),
		routeToWrap.RequiresAuth(),
		http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			slog.DebugContext(request.Context(), "applying log middleware")

			slog.InfoContext(request.Context(), "incoming HTTP request",
				"method", request.Method, "url", request.URL)

			routeToWrap.ServeHTTP(
				writer,
				request,
			)
		}))
}
