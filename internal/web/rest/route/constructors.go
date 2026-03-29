package route

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
