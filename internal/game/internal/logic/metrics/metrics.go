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

// SummaryMoveBuckets defines bucket boundaries for total moves per game.
var SummaryMoveBuckets = []float64{10, 25, 50, 100, 200, 500, 1000}

// SummaryTurnBuckets defines bucket boundaries for total turns (phase transitions) per game.
var SummaryTurnBuckets = []float64{5, 10, 25, 50, 100, 200, 500}

// SummaryAttackBuckets defines bucket boundaries for total attack moves per game.
var SummaryAttackBuckets = []float64{5, 10, 25, 50, 100, 200, 500}

// SummaryHeadlineBuckets defines bucket boundaries for total headline events per game.
var SummaryHeadlineBuckets = []float64{1, 3, 5, 10, 25, 50, 100}

// GameMetrics holds OTel instruments for game-domain observability.
type GameMetrics struct {
	ActiveGames  metric.Int64UpDownCounter
	GameDuration metric.Float64Histogram

	// Summary histograms record per-game totals at completion.
	SummaryMoves     metric.Float64Histogram
	SummaryAttacks   metric.Float64Histogram
	SummaryTurns     metric.Float64Histogram
	SummaryHeadlines metric.Float64Histogram
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

	if gameMetrics.SummaryMoves, err = meter.Float64Histogram("game.summary.moves",
		metric.WithDescription("Total moves per completed game"),
		metric.WithExplicitBucketBoundaries(SummaryMoveBuckets...),
	); err != nil {
		return nil, fmt.Errorf("failed to create summary moves histogram: %w", err)
	}

	if gameMetrics.SummaryAttacks, err = meter.Float64Histogram("game.summary.attacks",
		metric.WithDescription("Total attack moves per completed game"),
		metric.WithExplicitBucketBoundaries(SummaryAttackBuckets...),
	); err != nil {
		return nil, fmt.Errorf("failed to create summary attacks histogram: %w", err)
	}

	if gameMetrics.SummaryTurns, err = meter.Float64Histogram("game.summary.turns",
		metric.WithDescription("Total phase transitions per completed game"),
		metric.WithExplicitBucketBoundaries(SummaryTurnBuckets...),
	); err != nil {
		return nil, fmt.Errorf("failed to create summary turns histogram: %w", err)
	}

	if gameMetrics.SummaryHeadlines, err = meter.Float64Histogram("game.summary.headlines",
		metric.WithDescription("Total headline events per completed game"),
		metric.WithExplicitBucketBoundaries(SummaryHeadlineBuckets...),
	); err != nil {
		return nil, fmt.Errorf("failed to create summary headlines histogram: %w", err)
	}

	return gameMetrics, nil
}
