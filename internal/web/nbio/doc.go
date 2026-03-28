// Package nbio configures and manages the nbio HTTP/WebSocket engine.
//
// [NewEngine] creates an [nbhttp.Engine] from the provided configuration and
// registers fx lifecycle hooks to start and gracefully shut down the engine.
// The engine serves both REST and WebSocket traffic on a single TCP listener
// (port 8080) using nbio's non-blocking I/O model.
//
// # Layer
//
// Web — network server lifecycle management.
package nbio
