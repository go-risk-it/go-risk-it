package route

import (
	"net/http"
)

type Route struct {
	handler      http.Handler
	pattern      string
	requiresAuth bool
	isWebSocket  bool
}

func newRoute(pattern string, requiresAuth bool, handler http.Handler) *Route {
	return &Route{
		pattern:      pattern,
		handler:      handler,
		requiresAuth: requiresAuth,
	}
}

func newWebSocketRoute(pattern string, requiresAuth bool, handler http.Handler) *Route {
	return &Route{
		pattern:      pattern,
		handler:      handler,
		requiresAuth: requiresAuth,
		isWebSocket:  true,
	}
}

// Wrap creates a new Route with the same metadata but a different handler.
func (r *Route) Wrap(handler http.Handler) *Route {
	return &Route{
		pattern:      r.pattern,
		requiresAuth: r.requiresAuth,
		isWebSocket:  r.isWebSocket,
		handler:      handler,
	}
}

func (r *Route) Pattern() string {
	return r.pattern
}

func (r *Route) RequiresAuth() bool {
	return r.requiresAuth
}

func (r *Route) IsWebSocket() bool {
	return r.isWebSocket
}

func (r *Route) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.handler.ServeHTTP(w, req)
}
