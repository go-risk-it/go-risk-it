// Package mux assembles all routes and middleware into a single http.Handler.
//
// [NewServeMux] collects routes via fx group injection, applies the middleware
// chain (log, OTel, auth) per route, registers them on a standard
// [http.ServeMux], and wraps the mux with CORS handling. The resulting handler
// is the entry point for all HTTP traffic.
//
// Note: otelhttp.NewHandler was removed because [middleware.OTelMiddleware]
// already creates per-route spans with accurate http.route attributes. The
// mux-level handler added a redundant root span (route="/") that polluted
// spanmetrics with an uninformative aggregation.
//
// # Layer
//
// Web — HTTP multiplexer composition and middleware wiring.
package mux
