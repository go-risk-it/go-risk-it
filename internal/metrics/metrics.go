package metrics

import (
	"fmt"

	"go.opentelemetry.io/otel/metric"
)

// LatencyBuckets defines histogram bucket boundaries in seconds,
// suitable for sub-second API and phase latencies.
var LatencyBuckets = []float64{
	0.001, 0.005, 0.01, 0.025, 0.05,
	0.1, 0.25, 0.5, 1, 2.5, 5, 10,
}

// GameDurationBuckets defines histogram bucket boundaries in seconds,
// suitable for game durations (seconds to minutes).
var GameDurationBuckets = []float64{
	1, 5, 10, 30, 60, 120, 300, 600, 1800, 3600,
}

type Metrics struct {
	// HTTP metrics
	HTTPRequestDuration metric.Float64Histogram
	HTTPRequestsTotal   metric.Int64Counter
	HTTPErrorsTotal     metric.Int64Counter

	// Game metrics
	ActiveGames   metric.Int64UpDownCounter
	MovesTotal    metric.Int64Counter
	PhaseDuration metric.Float64Histogram
	GamesCreated  metric.Int64Counter
	GamesFinished metric.Int64Counter
	GameDuration  metric.Float64Histogram

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
}

func NewMetrics(meter metric.Meter) (*Metrics, error) {
	metrics := &Metrics{}

	if err := metrics.initHTTPMetrics(meter); err != nil {
		return nil, err
	}

	if err := metrics.initGameMetrics(meter); err != nil {
		return nil, err
	}

	if err := metrics.initWebSocketMetrics(meter); err != nil {
		return nil, err
	}

	if err := metrics.initDBMetrics(meter); err != nil {
		return nil, err
	}

	return metrics, nil
}

func (metrics *Metrics) initHTTPMetrics(meter metric.Meter) error {
	var err error

	if metrics.HTTPRequestDuration, err = meter.Float64Histogram("http.server.request.duration",
		metric.WithDescription("Duration of HTTP requests in seconds"),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(LatencyBuckets...),
	); err != nil {
		return fmt.Errorf("failed to create http request duration histogram: %w", err)
	}

	if metrics.HTTPRequestsTotal, err = meter.Int64Counter("http.server.requests.total",
		metric.WithDescription("Total number of HTTP requests"),
	); err != nil {
		return fmt.Errorf("failed to create http requests total counter: %w", err)
	}

	if metrics.HTTPErrorsTotal, err = meter.Int64Counter("http.server.errors.total",
		metric.WithDescription("Total number of HTTP error responses"),
	); err != nil {
		return fmt.Errorf("failed to create http errors total counter: %w", err)
	}

	return nil
}

func (metrics *Metrics) initGameMetrics(meter metric.Meter) error {
	var err error

	if metrics.ActiveGames, err = meter.Int64UpDownCounter("game.active",
		metric.WithDescription("Number of currently active games"),
	); err != nil {
		return fmt.Errorf("failed to create active games counter: %w", err)
	}

	if metrics.MovesTotal, err = meter.Int64Counter("game.moves.total",
		metric.WithDescription("Total number of moves performed"),
	); err != nil {
		return fmt.Errorf("failed to create moves total counter: %w", err)
	}

	if metrics.PhaseDuration, err = meter.Float64Histogram("game.phase.duration",
		metric.WithDescription("Duration of phase execution in seconds"),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(LatencyBuckets...),
	); err != nil {
		return fmt.Errorf("failed to create phase duration histogram: %w", err)
	}

	if metrics.GamesCreated, err = meter.Int64Counter("game.created.total",
		metric.WithDescription("Total number of games created"),
	); err != nil {
		return fmt.Errorf("failed to create games created counter: %w", err)
	}

	if metrics.GamesFinished, err = meter.Int64Counter("game.finished.total",
		metric.WithDescription("Total number of games finished"),
	); err != nil {
		return fmt.Errorf("failed to create games finished counter: %w", err)
	}

	if metrics.GameDuration, err = meter.Float64Histogram("game.duration",
		metric.WithDescription("Duration of completed games in seconds"),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(GameDurationBuckets...),
	); err != nil {
		return fmt.Errorf("failed to create game duration histogram: %w", err)
	}

	return nil
}

func (metrics *Metrics) initWebSocketMetrics(meter metric.Meter) error {
	var err error

	if metrics.ActiveConnections, err = meter.Int64UpDownCounter("ws.connections.active",
		metric.WithDescription("Number of active WebSocket connections"),
	); err != nil {
		return fmt.Errorf("failed to create active connections counter: %w", err)
	}

	if metrics.BroadcastDuration, err = meter.Float64Histogram("ws.broadcast.duration",
		metric.WithDescription("Duration of broadcast operations in seconds"),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(LatencyBuckets...),
	); err != nil {
		return fmt.Errorf("failed to create broadcast duration histogram: %w", err)
	}

	if metrics.MessagesSent, err = meter.Int64Counter("ws.messages.sent.total",
		metric.WithDescription("Total number of WebSocket messages sent"),
	); err != nil {
		return fmt.Errorf("failed to create messages sent counter: %w", err)
	}

	if metrics.BroadcastErrors, err = meter.Int64Counter("ws.broadcast.errors.total",
		metric.WithDescription("Total number of broadcast errors"),
	); err != nil {
		return fmt.Errorf("failed to create broadcast errors counter: %w", err)
	}

	if metrics.BroadcastFanOut, err = meter.Int64Histogram("ws.broadcast.fanout",
		metric.WithDescription("Number of connections per broadcast"),
	); err != nil {
		return fmt.Errorf("failed to create broadcast fanout histogram: %w", err)
	}

	return nil
}

func (metrics *Metrics) initDBMetrics(meter metric.Meter) error {
	var err error

	if metrics.TransactionDuration, err = meter.Float64Histogram("db.transaction.duration",
		metric.WithDescription("Duration of database transactions"),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(LatencyBuckets...),
	); err != nil {
		return fmt.Errorf("failed to create transaction duration histogram: %w", err)
	}

	if metrics.TransactionRollbacks, err = meter.Int64Counter("db.transaction.rollbacks.total",
		metric.WithDescription("Total number of transaction rollbacks"),
	); err != nil {
		return fmt.Errorf("failed to create transaction rollbacks counter: %w", err)
	}

	if metrics.TransactionRetries, err = meter.Int64Counter(
		"db.transaction.retries.total",
		metric.WithDescription(
			"Total number of transaction retries due to serialization failures",
		),
	); err != nil {
		return fmt.Errorf("failed to create transaction retries counter: %w", err)
	}

	return nil
}
