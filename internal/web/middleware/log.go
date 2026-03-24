package middleware

import (
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	"go.uber.org/zap"
)

type LogMiddleware struct {
	log *zap.SugaredLogger
}

func NewLogMiddleware(log *zap.SugaredLogger) *LogMiddleware {
	return &LogMiddleware{
		log: log,
	}
}

func (m *LogMiddleware) Wrap(routeToWrap route.Route) route.Route {
	return route.NewRoute(
		routeToWrap.Pattern(),
		routeToWrap.RequiresAuth(),
		http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			ctx := ctx.WithLog(request.Context(), m.log)
			ctx.Log().Debug("applying log middleware")

			ctx.Log().Infow("incoming HTTP request", "method", request.Method, "url", request.URL)

			routeToWrap.ServeHTTP(
				writer,
				request.WithContext(ctx),
			)
		}))
}
