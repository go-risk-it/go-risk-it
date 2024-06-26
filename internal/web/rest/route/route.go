package route

import (
	"net/http"

	"go.uber.org/fx"
)

type Route interface {
	http.Handler

	Pattern() string
	RequiresAuth() bool
}

type RouteImpl struct {
	handler      http.Handler
	pattern      string
	requiresAuth bool
}

func NewRoute(pattern string, requiresAuth bool, handler http.Handler) *RouteImpl {
	return &RouteImpl{
		pattern:      pattern,
		handler:      handler,
		requiresAuth: requiresAuth,
	}
}

func (r *RouteImpl) Pattern() string {
	return r.pattern
}

func (r *RouteImpl) RequiresAuth() bool {
	return r.requiresAuth
}

func (r *RouteImpl) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.handler.ServeHTTP(w, req)
}

func AsRoute(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(Route)),
		fx.ResultTags(`group:"routes"`),
	)
}
