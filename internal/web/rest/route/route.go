package route

import (
	"net/http"

	"go.uber.org/fx"
)

type Route struct {
	handler      http.Handler
	pattern      string
	requiresAuth bool
}

func New(pattern string, requiresAuth bool, handler http.Handler) *Route {
	return &Route{
		pattern:      pattern,
		handler:      handler,
		requiresAuth: requiresAuth,
	}
}

func (r *Route) Pattern() string {
	return r.pattern
}

func (r *Route) RequiresAuth() bool {
	return r.requiresAuth
}

func (r *Route) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.handler.ServeHTTP(w, req)
}

func AsRoute(f any) any {
	return fx.Annotate(
		f,
		fx.ResultTags(`group:"routes"`),
	)
}
