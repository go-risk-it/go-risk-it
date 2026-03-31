package metrics

import (
	"fmt"

	"go.opentelemetry.io/otel/metric"
)

// StateMetrics holds OTel instruments for infrastructure state observability:
// WebSocket connection tracking and DB transaction retries.
type StateMetrics struct {
	// WebSocket metrics
	ActiveConnections metric.Int64UpDownCounter

	// DB metrics
	TransactionRetries metric.Int64Counter
}

// NewStateMetrics creates all infrastructure state OTel instruments on the given meter.
func NewStateMetrics(meter metric.Meter) (*StateMetrics, error) {
	stateMetrics := &StateMetrics{}

	if err := stateMetrics.initWebSocketMetrics(meter); err != nil {
		return nil, err
	}

	if err := stateMetrics.initDBMetrics(meter); err != nil {
		return nil, err
	}

	return stateMetrics, nil
}

func (m *StateMetrics) initWebSocketMetrics(meter metric.Meter) error {
	var err error

	if m.ActiveConnections, err = meter.Int64UpDownCounter("ws.connections.active",
		metric.WithDescription("Number of active WebSocket connections"),
	); err != nil {
		return fmt.Errorf("failed to create active connections counter: %w", err)
	}

	return nil
}

func (m *StateMetrics) initDBMetrics(meter metric.Meter) error {
	var err error

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
