package metrics

import (
	"fmt"

	"go.opentelemetry.io/otel/metric"
)

// GameDurationBuckets defines histogram bucket boundaries in seconds,
// suitable for game durations (seconds to minutes).
var GameDurationBuckets = []float64{
	1, 5, 10, 30, 60, 120, 300, 600, 1800, 3600,
}

// GameMetrics holds OTel instruments for game-domain observability.
type GameMetrics struct {
	ActiveGames  metric.Int64UpDownCounter
	GameDuration metric.Float64Histogram
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

	if gameMetrics.GameDuration, err = meter.Float64Histogram("game.duration",
		metric.WithDescription("Duration of completed games in seconds"),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(GameDurationBuckets...),
	); err != nil {
		return nil, fmt.Errorf("failed to create game duration histogram: %w", err)
	}

	return gameMetrics, nil
}
