package route

import "net/http"

// Public creates an unauthenticated route with error handling.
func Public(pattern string, handler PlainHandler) *Route {
	return &Route{
		pattern:      pattern,
		requiresAuth: false,
		handler:      WrapErrors(handler),
	}
}

// Authed creates an authenticated route with error handling.
func Authed(pattern string, handler PlainHandler) *Route {
	return &Route{
		pattern:      pattern,
		requiresAuth: true,
		handler:      WrapErrors(handler),
	}
}

// Domain creates an authenticated route that builds a domain context C from the
// request before invoking the handler. Both context-building errors and handler
// errors are translated to HTTP responses via [WrapErrors].
func Domain[C any](
	pattern string,
	buildCtx func(*http.Request) (C, error),
	handler func(http.ResponseWriter, *http.Request, C) error,
) *Route {
	return New(
		pattern,
		true,
		WrapErrors(func(writer http.ResponseWriter, request *http.Request) error {
			ctx, err := buildCtx(request)
			if err != nil {
				return err
			}

			return handler(writer, request, ctx)
		}),
	)
}

// DomainWS creates an authenticated WebSocket route that builds a domain context
// C from the request before invoking the handler. Behaves like [Domain] but sets
// the WebSocket flag on the resulting [Route].
func DomainWS[C any](
	pattern string,
	buildCtx func(*http.Request) (C, error),
	handler func(http.ResponseWriter, *http.Request, C) error,
) *Route {
	return NewWS(
		pattern,
		true,
		WrapErrors(func(writer http.ResponseWriter, request *http.Request) error {
			ctx, err := buildCtx(request)
			if err != nil {
				return err
			}

			return handler(writer, request, ctx)
		}),
	)
}
