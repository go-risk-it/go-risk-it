// Package metrics registers manual OpenTelemetry instruments for resource
// state that has no span equivalent. These are the only metrics that should
// be created by hand — all request-scoped metrics (latency, throughput,
// errors) are derived automatically from traces by the spanmetrics connector.
//
// # What lives here
//
// [StateMetrics] holds two instruments:
//   - ws.connections.active (UpDownCounter) — current WebSocket connection count
//   - db.transaction.retries.total (Counter) — serialization-failure retry count
//
// Game-domain state gauges live in internal/game/logic/metrics:
//   - game.active (UpDownCounter) — number of in-progress games
//   - game.duration (Histogram) — elapsed time of completed games
//
// These instruments track resource state that exists between requests. No span
// can represent "how many WebSocket connections are open right now" because that
// is a continuous gauge, not a request with a start and end.
//
// # What NOT to add
//
// Do not create manual OTel instruments for any of the following — they are
// derived from spans by the OTel Collector's spanmetrics connector
// (see grafana/otelcol-extra.yaml):
//
//   - Latency histograms — use [observe.Span] or [observe.RawSpan] instead.
//     The spanmetrics connector
//     produces traces_spanmetrics_duration_seconds_bucket from every span.
//   - Per-operation counters — use [observe.Span] or [observe.RawSpan]
//     instead. The connector produces
//     traces_spanmetrics_calls_total from every span, broken down by span name
//     and configured dimension attributes.
//   - Error rate metrics — use [observe.Error] instead. It sets the span status to
//     error, which the connector surfaces as calls_total{status_code="STATUS_CODE_ERROR"}.
//
// This is the three-signal model: application code emits traces (via observe)
// and state gauges (via this package). The OTel Collector derives RED metrics
// (Rate, Errors, Duration) from traces, eliminating an entire class of manual
// instrumentation.
//
// # Layer
//
// Kernel — infrastructure state instruments with no internal dependencies.
package metrics
