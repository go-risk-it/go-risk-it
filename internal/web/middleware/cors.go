package middleware

import (
	"net/http"
	"strings"

	"github.com/go-risk-it/go-risk-it/internal/kernel/config"
)

type CorsMiddleware struct {
	allowedOrigins map[string]bool
	allowOriginStr string
}

func NewCorsMiddleware(serverConfig config.ServerConfig) *CorsMiddleware {
	allowed := make(map[string]bool, len(serverConfig.AllowedOrigins))
	for _, origin := range serverConfig.AllowedOrigins {
		allowed[origin] = true
	}

	return &CorsMiddleware{
		allowedOrigins: allowed,
		allowOriginStr: strings.Join(serverConfig.AllowedOrigins, ", "),
	}
}

// WrapHandler wraps an http.Handler with CORS headers and OPTIONS preflight
// handling. This must be applied at the mux level (not per-route) because Go's
// ServeMux rejects OPTIONS requests with 405 before per-route handlers run.
func (m *CorsMiddleware) WrapHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		origin := request.Header.Get("Origin")
		if origin != "" && m.allowedOrigins[origin] {
			writer.Header().Set("Access-Control-Allow-Origin", origin)
			writer.Header().Set("Vary", "Origin")
		}

		writer.Header().Set(
			"Access-Control-Allow-Methods",
			"GET, POST, PUT, DELETE, OPTIONS",
		)
		writer.Header().Set(
			"Access-Control-Allow-Headers",
			"Content-Type, Authorization",
		)

		if request.Method == http.MethodOptions {
			writer.WriteHeader(http.StatusOK)

			return
		}

		next.ServeHTTP(writer, request)
	})
}
