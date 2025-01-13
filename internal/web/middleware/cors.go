package middleware

import (
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/config"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
)

type CorsMiddleware interface {
	Middleware
}

type CorsMiddlewareImpl struct {
	jwtConfig config.JwtConfig
}

var _ CorsMiddleware = (*CorsMiddlewareImpl)(nil)

func NewCorsMiddleware(jwtConfig config.JwtConfig) CorsMiddleware {
	return &CorsMiddlewareImpl{jwtConfig: jwtConfig}
}

func (m *CorsMiddlewareImpl) Wrap(routeToWrap route.Route) route.Route {
	return route.NewRoute(
		routeToWrap.Pattern(),
		routeToWrap.RequiresAuth(),
		http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.Header().Set("Access-Control-Allow-Origin", "*")
			writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if request.Method == http.MethodOptions {
				writer.WriteHeader(http.StatusOK)

				return
			}

			routeToWrap.ServeHTTP(writer, request)
		}))
}
