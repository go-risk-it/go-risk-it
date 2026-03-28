// Package middleware provides HTTP middleware for the request pipeline.
//
// Four middleware types wrap routes in the mux assembly order:
//
//   - [LogMiddleware] logs incoming requests at Info level.
//   - [OTelMiddleware] creates trace spans, records HTTP duration and error
//     metrics, and injects a TraceContext into the request.
//   - [AuthMiddleware] verifies JWT Bearer tokens and enriches the context
//     with UserContext. It also extracts tokens smuggled via the
//     Sec-WebSocket-Protocol header for browser WS connections.
//   - [CorsMiddleware] handles CORS headers and OPTIONS preflight at the
//     mux level.
//
// # Layer
//
// Web — cross-cutting HTTP request processing.
package middleware
