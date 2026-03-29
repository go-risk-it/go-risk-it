package metrics

import (
	"fmt"

	"go.opentelemetry.io/otel/metric"
)

// LatencyBuckets defines histogram bucket boundaries in seconds,
// suitable for sub-second phase latencies.
var LatencyBuckets = []float64{
	0.001, 0.005, 0.01, 0.025, 0.05,
	0.1, 0.25, 0.5, 1, 2.5, 5, 10,
}

// GameDurationBuckets defines histogram bucket boundaries in seconds,
// suitable for game durations (seconds to minutes).
var GameDurationBuckets = []float64{
	1, 5, 10, 30, 60, 120, 300, 600, 1800, 3600,
}

// GameMetrics holds OTel instruments for game-domain observability.
type GameMetrics struct {
	ActiveGames   metric.Int64UpDownCounter
	MovesTotal    metric.Int64Counter
	PhaseDuration metric.Float64Histogram
	GamesCreated  metric.Int64Counter
	GamesFinished metric.Int64Counter
	GameDuration  metric.Float64Histogram
}

// NewGameMetrics creates all game-domain OTel instruments on the given meter.
func NewGameMetrics(meter metric.Meter) (*GameMetrics, error) {
	gameMetrics := &GameMetrics{}

	var err error

	if gameMetrics.ActiveGames, err = meter.Int64UpDownCounter("game.active",
		metric.WithDescription("Number of currently active games"),
	); err != nil {
		return nil, fmt.Errorf("failed to create active games counter: %w", err)
	}

	if gameMetrics.MovesTotal, err = meter.Int64Counter("game.moves.total",
		metric.WithDescription("Total number of moves performed"),
	); err != nil {
		return nil, fmt.Errorf("failed to create moves total counter: %w", err)
	}

	if gameMetrics.PhaseDuration, err = meter.Float64Histogram("game.phase.duration",
		metric.WithDescription("Duration of phase execution in seconds"),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(LatencyBuckets...),
	); err != nil {
		return nil, fmt.Errorf("failed to create phase duration histogram: %w", err)
	}

	if gameMetrics.GamesCreated, err = meter.Int64Counter("game.created.total",
		metric.WithDescription("Total number of games created"),
	); err != nil {
		return nil, fmt.Errorf("failed to create games created counter: %w", err)
	}

	if gameMetrics.GamesFinished, err = meter.Int64Counter("game.finished.total",
		metric.WithDescription("Total number of games finished"),
	); err != nil {
		return nil, fmt.Errorf("failed to create games finished counter: %w", err)
	}

	if gameMetrics.GameDuration, err = meter.Float64Histogram("game.duration",
		metric.WithDescription("Duration of completed games in seconds"),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(GameDurationBuckets...),
	); err != nil {
		return nil, fmt.Errorf("failed to create game duration histogram: %w", err)
	}

	return gameMetrics, nil
}
