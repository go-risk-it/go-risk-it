// Package mux assembles all routes and middleware into a single http.Handler.
//
// [NewServeMux] collects routes via fx group injection, applies the middleware
// chain (log, OTel, auth) per route, registers them on a standard
// [http.ServeMux], and wraps the mux with CORS handling and otelhttp
// instrumentation. The resulting handler is the entry point for all HTTP
// traffic.
//
// # Layer
//
// Web — HTTP multiplexer composition and middleware wiring.
package mux
