package middleware

import (
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
)

type LogMiddleware struct{}

func NewLogMiddleware() *LogMiddleware {
	return &LogMiddleware{}
}

func (m *LogMiddleware) Wrap(routeToWrap *route.Route) *route.Route {
	return routeToWrap.Wrap(http.HandlerFunc(
		func(writer http.ResponseWriter, request *http.Request) {
			routeToWrap.ServeHTTP(
				writer,
				request,
			)
		},
	))
}
