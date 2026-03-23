package baseline_test

import (
	"testing"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/baseline"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/metrics"
	"github.com/stretchr/testify/assert"
)

func TestSnapshotToMetrics_Conversion(t *testing.T) {
	t.Parallel()

	snap := &metrics.Snapshot{
		E2EMove:                metrics.HistogramSnapshot{P50: 200, P95: 350, P99: 500, Max: 800},
		WSDelivery:             metrics.HistogramSnapshot{P50: 80, P95: 120, P99: 180, Max: 300},
		TotalMoves:             1000,
		TotalErrors:            5,
		GamesCompleted:         8,
		GamesTimedOut:          1,
		GamesFatal:             0,
		TotalRetries:           10,
		TotalConflicts:         3,
		TotalReconnects:        2,
		TotalReconnectFailures: 1,
		PhaseLatency: map[string]metrics.HistogramSnapshot{
			"deploy": {P50: 100, P95: 200, P99: 300, Max: 400},
			"attack": {P50: 150, P95: 350, P99: 500, Max: 700},
		},
		RESTLatency: map[string]metrics.HistogramSnapshot{
			"deploy":    {P50: 50, P95: 100, P99: 150, Max: 200},
			"reinforce": {P50: 60, P95: 110, P99: 160, Max: 210},
		},
		ErrorBreakdown: map[string]int64{
			"timeout":   3,
			"transient": 2,
		},
		PhaseEntries: map[string]int64{"deploy": 20, "attack": 30},
		PhaseMoves:   map[string]int64{"deploy": 100, "attack": 200},
		ThroughputBuckets: []metrics.ThroughputBucket{
			{OffsetSec: 0, Moves: 50},
			{OffsetSec: 5, Moves: 80},
			{OffsetSec: 10, Moves: 60},
		},
	}

	m := baseline.SnapshotToMetrics(snap, 60.0)

	// E2E latency profile.
	assert.InDelta(t, 0.200, m.E2E.P50, 0.0001)
	assert.InDelta(t, 0.350, m.E2E.P95, 0.0001)
	assert.InDelta(t, 0.500, m.E2E.P99, 0.0001)
	assert.InDelta(t, 0.800, m.E2E.Max, 0.0001)

	// WS delivery latency profile.
	assert.InDelta(t, 0.080, m.WSDelivery.P50, 0.0001)
	assert.InDelta(t, 0.120, m.WSDelivery.P95, 0.0001)
	assert.InDelta(t, 0.180, m.WSDelivery.P99, 0.0001)
	assert.InDelta(t, 0.300, m.WSDelivery.Max, 0.0001)

	// Throughput.
	assert.InDelta(t, 16.6667, m.ThroughputMPS, 0.001)
	assert.InDelta(t, 16.0, m.ThroughputPeakMPS, 0.001) // 80 moves / 5s bucket

	// Error rate.
	assert.InDelta(t, 0.005, m.HTTPErrorRate, 0.0001)

	// Counters.
	assert.Equal(t, int64(1000), m.TotalMoves)
	assert.Equal(t, int64(5), m.TotalErrors)
	assert.Equal(t, int64(8), m.GamesCompleted)
	assert.Equal(t, int64(1), m.GamesTimedOut)
	assert.Equal(t, int64(0), m.GamesFatal)

	// Resilience.
	assert.Equal(t, int64(10), m.TotalRetries)
	assert.Equal(t, int64(3), m.TotalConflicts)
	assert.Equal(t, int64(2), m.TotalReconnects)
	assert.Equal(t, int64(1), m.TotalReconnectFailures)

	// Phase latency.
	assert.Len(t, m.PhaseLatency, 2)
	assert.InDelta(t, 0.200, m.PhaseLatency["deploy"].P95, 0.0001)
	assert.InDelta(t, 0.350, m.PhaseLatency["attack"].P95, 0.0001)

	// REST latency.
	assert.Len(t, m.RESTLatency, 2)
	assert.InDelta(t, 0.100, m.RESTLatency["deploy"].P95, 0.0001)
	assert.InDelta(t, 0.110, m.RESTLatency["reinforce"].P95, 0.0001)

	// Error breakdown.
	assert.Equal(t, int64(3), m.ErrorBreakdown["timeout"])
	assert.Equal(t, int64(2), m.ErrorBreakdown["transient"])

	// Phase flow.
	assert.Equal(t, int64(20), m.PhaseEntries["deploy"])
	assert.Equal(t, int64(200), m.PhaseMoves["attack"])

	// Duration.
	assert.InDelta(t, 60.0, m.DurationSec, 0.0001)
}

func TestSnapshotToMetrics_ZeroMoves(t *testing.T) {
	t.Parallel()

	snap := &metrics.Snapshot{
		TotalMoves:  0,
		TotalErrors: 0,
	}

	m := baseline.SnapshotToMetrics(snap, 60.0)

	assert.Equal(t, 0.0, m.HTTPErrorRate, "zero moves should produce zero error rate")
	assert.Equal(t, 0.0, m.ThroughputMPS, "zero moves should produce zero throughput")
}

func TestSnapshotToMetrics_ZeroDuration(t *testing.T) {
	t.Parallel()

	snap := &metrics.Snapshot{
		TotalMoves:  100,
		TotalErrors: 1,
	}

	m := baseline.SnapshotToMetrics(snap, 0)

	assert.Equal(t, 0.0, m.ThroughputMPS, "zero duration should produce zero throughput")
	assert.InDelta(t, 0.01, m.HTTPErrorRate, 0.0001, "error rate should still work")
}

func TestSnapshotToMetrics_OmitsZeroEntries(t *testing.T) {
	t.Parallel()

	snap := &metrics.Snapshot{
		ErrorBreakdown: map[string]int64{"timeout": 0, "transient": 0},
		PhaseEntries:   map[string]int64{"deploy": 0},
		PhaseMoves:     map[string]int64{"deploy": 0},
	}

	m := baseline.SnapshotToMetrics(snap, 10.0)

	assert.Nil(t, m.ErrorBreakdown, "all-zero error breakdown should be nil")
	assert.Nil(t, m.PhaseEntries, "all-zero phase entries should be nil")
	assert.Nil(t, m.PhaseMoves, "all-zero phase moves should be nil")
}
