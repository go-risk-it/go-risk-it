// Package metrics registers OpenTelemetry instruments for infrastructure
// observability: HTTP request handling, WebSocket broadcasting, database
// transactions, and event bus dispatch.
//
// [NewInfraMetrics] creates all instruments from a single OTel Meter
// and returns the consolidated [InfraMetrics] struct. Game-domain metrics
// live in internal/game/logic/metrics.
//
// # Layer
//
// Kernel — infrastructure observability instruments with no internal dependencies.
package metrics
