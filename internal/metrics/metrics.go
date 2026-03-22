package metrics

import (
	"go.opentelemetry.io/otel/metric"
)

type Metrics struct {
	// HTTP metrics
	HTTPRequestDuration metric.Float64Histogram
	HTTPRequestsTotal   metric.Int64Counter

	// Game metrics
	ActiveGames   metric.Int64UpDownCounter
	MovesTotal    metric.Int64Counter
	PhaseDuration metric.Float64Histogram
	GamesCreated  metric.Int64Counter
	GamesFinished metric.Int64Counter

	// WebSocket metrics
	ActiveConnections metric.Int64UpDownCounter
	BroadcastDuration metric.Float64Histogram
	MessagesSent      metric.Int64Counter
	BroadcastErrors   metric.Int64Counter
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

	return metrics, nil
}

func (metrics *Metrics) initHTTPMetrics(meter metric.Meter) error {
	var err error

	if metrics.HTTPRequestDuration, err = meter.Float64Histogram("http.server.request.duration",
		metric.WithDescription("Duration of HTTP requests in seconds"),
		metric.WithUnit("s"),
	); err != nil {
		return err
	}

	if metrics.HTTPRequestsTotal, err = meter.Int64Counter("http.server.requests.total",
		metric.WithDescription("Total number of HTTP requests"),
	); err != nil {
		return err
	}

	return nil
}

func (metrics *Metrics) initGameMetrics(meter metric.Meter) error {
	var err error

	if metrics.ActiveGames, err = meter.Int64UpDownCounter("game.active",
		metric.WithDescription("Number of currently active games"),
	); err != nil {
		return err
	}

	if metrics.MovesTotal, err = meter.Int64Counter("game.moves.total",
		metric.WithDescription("Total number of moves performed"),
	); err != nil {
		return err
	}

	if metrics.PhaseDuration, err = meter.Float64Histogram("game.phase.duration",
		metric.WithDescription("Duration of phase execution in seconds"),
		metric.WithUnit("s"),
	); err != nil {
		return err
	}

	if metrics.GamesCreated, err = meter.Int64Counter("game.created.total",
		metric.WithDescription("Total number of games created"),
	); err != nil {
		return err
	}

	if metrics.GamesFinished, err = meter.Int64Counter("game.finished.total",
		metric.WithDescription("Total number of games finished"),
	); err != nil {
		return err
	}

	return nil
}

func (metrics *Metrics) initWebSocketMetrics(meter metric.Meter) error {
	var err error

	if metrics.ActiveConnections, err = meter.Int64UpDownCounter("ws.connections.active",
		metric.WithDescription("Number of active WebSocket connections"),
	); err != nil {
		return err
	}

	if metrics.BroadcastDuration, err = meter.Float64Histogram("ws.broadcast.duration",
		metric.WithDescription("Duration of broadcast operations in seconds"),
		metric.WithUnit("s"),
	); err != nil {
		return err
	}

	if metrics.MessagesSent, err = meter.Int64Counter("ws.messages.sent.total",
		metric.WithDescription("Total number of WebSocket messages sent"),
	); err != nil {
		return err
	}

	if metrics.BroadcastErrors, err = meter.Int64Counter("ws.broadcast.errors.total",
		metric.WithDescription("Total number of broadcast errors"),
	); err != nil {
		return err
	}

	return nil
}
