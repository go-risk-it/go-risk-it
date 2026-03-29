package metrics

import (
	"fmt"

	"go.opentelemetry.io/otel/metric"
)

// LatencyBuckets defines histogram bucket boundaries in seconds,
// suitable for sub-second API latencies.
var LatencyBuckets = []float64{
	0.001, 0.005, 0.01, 0.025, 0.05,
	0.1, 0.25, 0.5, 1, 2.5, 5, 10,
}

// InfraMetrics holds OTel instruments for infrastructure observability:
// HTTP, WebSocket, DB transactions, and event bus.
type InfraMetrics struct {
	// HTTP metrics
	HTTPRequestDuration metric.Float64Histogram
	HTTPRequestsTotal   metric.Int64Counter
	HTTPErrorsTotal     metric.Int64Counter

	// WebSocket metrics
	ActiveConnections metric.Int64UpDownCounter
	BroadcastDuration metric.Float64Histogram
	MessagesSent      metric.Int64Counter
	BroadcastErrors   metric.Int64Counter
	BroadcastFanOut   metric.Int64Histogram

	// DB metrics
	TransactionDuration  metric.Float64Histogram
	TransactionRollbacks metric.Int64Counter
	TransactionRetries   metric.Int64Counter

	// Event bus metrics
	EventBusDispatchDuration metric.Float64Histogram
	EventBusEventsTotal      metric.Int64Counter
	EventHandlerDuration     metric.Float64Histogram
}

// NewInfraMetrics creates all infrastructure OTel instruments on the given meter.
func NewInfraMetrics(meter metric.Meter) (*InfraMetrics, error) {
	infraMetrics := &InfraMetrics{}

	if err := infraMetrics.initHTTPMetrics(meter); err != nil {
		return nil, err
	}

	if err := infraMetrics.initWebSocketMetrics(meter); err != nil {
		return nil, err
	}

	if err := infraMetrics.initDBMetrics(meter); err != nil {
		return nil, err
	}

	if err := infraMetrics.initEventBusMetrics(meter); err != nil {
		return nil, err
	}

	return infraMetrics, nil
}

func (m *InfraMetrics) initHTTPMetrics(meter metric.Meter) error {
	var err error

	if m.HTTPRequestDuration, err = meter.Float64Histogram("http.server.request.duration",
		metric.WithDescription("Duration of HTTP requests in seconds"),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(LatencyBuckets...),
	); err != nil {
		return fmt.Errorf("failed to create http request duration histogram: %w", err)
	}

	if m.HTTPRequestsTotal, err = meter.Int64Counter("http.server.requests.total",
		metric.WithDescription("Total number of HTTP requests"),
	); err != nil {
		return fmt.Errorf("failed to create http requests total counter: %w", err)
	}

	if m.HTTPErrorsTotal, err = meter.Int64Counter("http.server.errors.total",
		metric.WithDescription("Total number of HTTP error responses"),
	); err != nil {
		return fmt.Errorf("failed to create http errors total counter: %w", err)
	}

	return nil
}

func (m *InfraMetrics) initWebSocketMetrics(meter metric.Meter) error {
	var err error

	if m.ActiveConnections, err = meter.Int64UpDownCounter("ws.connections.active",
		metric.WithDescription("Number of active WebSocket connections"),
	); err != nil {
		return fmt.Errorf("failed to create active connections counter: %w", err)
	}

	if m.BroadcastDuration, err = meter.Float64Histogram("ws.broadcast.duration",
		metric.WithDescription("Duration of broadcast operations in seconds"),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(LatencyBuckets...),
	); err != nil {
		return fmt.Errorf("failed to create broadcast duration histogram: %w", err)
	}

	if m.MessagesSent, err = meter.Int64Counter("ws.messages.sent.total",
		metric.WithDescription("Total number of WebSocket messages sent"),
	); err != nil {
		return fmt.Errorf("failed to create messages sent counter: %w", err)
	}

	if m.BroadcastErrors, err = meter.Int64Counter("ws.broadcast.errors.total",
		metric.WithDescription("Total number of broadcast errors"),
	); err != nil {
		return fmt.Errorf("failed to create broadcast errors counter: %w", err)
	}

	if m.BroadcastFanOut, err = meter.Int64Histogram("ws.broadcast.fanout",
		metric.WithDescription("Number of connections per broadcast"),
	); err != nil {
		return fmt.Errorf("failed to create broadcast fanout histogram: %w", err)
	}

	return nil
}

func (m *InfraMetrics) initDBMetrics(meter metric.Meter) error {
	var err error

	if m.TransactionDuration, err = meter.Float64Histogram("db.transaction.duration",
		metric.WithDescription("Duration of database transactions"),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(LatencyBuckets...),
	); err != nil {
		return fmt.Errorf("failed to create transaction duration histogram: %w", err)
	}

	if m.TransactionRollbacks, err = meter.Int64Counter("db.transaction.rollbacks.total",
		metric.WithDescription("Total number of transaction rollbacks"),
	); err != nil {
		return fmt.Errorf("failed to create transaction rollbacks counter: %w", err)
	}

	if m.TransactionRetries, err = meter.Int64Counter(
		"db.transaction.retries.total",
		metric.WithDescription(
			"Total number of transaction retries due to serialization failures",
		),
	); err != nil {
		return fmt.Errorf("failed to create transaction retries counter: %w", err)
	}

	return nil
}

func (m *InfraMetrics) initEventBusMetrics(meter metric.Meter) error {
	var err error

	if m.EventBusDispatchDuration, err = meter.Float64Histogram(
		"event_bus.dispatch.duration",
		metric.WithDescription("Duration of event bus dispatch (all handlers) in seconds"),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(LatencyBuckets...),
	); err != nil {
		return fmt.Errorf("failed to create event bus dispatch duration histogram: %w", err)
	}

	if m.EventBusEventsTotal, err = meter.Int64Counter(
		"event_bus.events.total",
		metric.WithDescription("Total number of events emitted through the event bus"),
	); err != nil {
		return fmt.Errorf("failed to create event bus events total counter: %w", err)
	}

	if m.EventHandlerDuration, err = meter.Float64Histogram(
		"event_handler.duration",
		metric.WithDescription("Duration of individual event handler execution in seconds"),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(LatencyBuckets...),
	); err != nil {
		return fmt.Errorf("failed to create event handler duration histogram: %w", err)
	}

	return nil
}
