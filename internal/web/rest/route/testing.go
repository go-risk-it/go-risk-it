package route

import "net/http"

// NewForTest creates a plain route for use in tests outside the route package.
// It bypasses WrapErrors so tests can use raw http.Handler.
func NewForTest(pattern string, requiresAuth bool, handler http.Handler) *Route {
	return newRoute(pattern, requiresAuth, handler)
}

// NewWebSocketForTest creates a WebSocket route for use in tests outside the route package.
// It bypasses WrapErrors so tests can use raw http.Handler.
func NewWebSocketForTest(pattern string, requiresAuth bool, handler http.Handler) *Route {
	return newWebSocketRoute(pattern, requiresAuth, handler)
}
