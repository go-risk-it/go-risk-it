// Package metrics registers OpenTelemetry instruments for the game server.
//
// It defines histograms, counters, and up-down counters across four domains:
// HTTP request handling, game lifecycle, WebSocket broadcasting, and database
// transactions. [NewMetrics] creates all instruments from a single OTel Meter
// and returns the consolidated [Metrics] struct that other packages record into.
//
// # Layer
//
// Kernel — observability instruments with no internal dependencies.
package metrics
