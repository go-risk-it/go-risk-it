package middleware

import (
	"net/http"
	"strings"

	"github.com/go-risk-it/go-risk-it/internal/config"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
)

type CorsMiddleware interface {
	Middleware
}

type CorsMiddlewareImpl struct {
	allowedOrigins map[string]bool
	allowOriginStr string
}

var _ CorsMiddleware = (*CorsMiddlewareImpl)(nil)

func NewCorsMiddleware(serverConfig config.ServerConfig) *CorsMiddlewareImpl {
	allowed := make(map[string]bool, len(serverConfig.AllowedOrigins))
	for _, origin := range serverConfig.AllowedOrigins {
		allowed[origin] = true
	}

	return &CorsMiddlewareImpl{
		allowedOrigins: allowed,
		allowOriginStr: strings.Join(serverConfig.AllowedOrigins, ", "),
	}
}

func (m *CorsMiddlewareImpl) Wrap(routeToWrap route.Route) route.Route {
	return route.NewRoute(
		routeToWrap.Pattern(),
		routeToWrap.RequiresAuth(),
		http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			origin := request.Header.Get("Origin")
			if origin != "" && m.allowedOrigins[origin] {
				writer.Header().Set("Access-Control-Allow-Origin", origin)
				writer.Header().Set("Vary", "Origin")
			}

			writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			writer.Header().
				Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if request.Method == http.MethodOptions {
				writer.WriteHeader(http.StatusOK)

				return
			}

			routeToWrap.ServeHTTP(writer, request)
		}))
}
