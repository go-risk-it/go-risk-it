package baseline_test

import (
	"testing"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/baseline"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindBreakingPoints_DetectsE2ESaturation(t *testing.T) {
	t.Parallel()

	runs := []baseline.LevelResult{
		{Games: 2, Metrics: baseline.MetricsSnapshot{E2E: baseline.LatencyProfile{P95: 0.2}}},
		{Games: 4, Metrics: baseline.MetricsSnapshot{E2E: baseline.LatencyProfile{P95: 0.35}}},
		{Games: 8, Metrics: baseline.MetricsSnapshot{E2E: baseline.LatencyProfile{P95: 0.65}}},
	}

	slos := baseline.DefaultSLOs()
	breakingPoints := baseline.FindBreakingPoints(runs, slos)

	bp := findBP(breakingPoints, "E2E move latency p95")
	require.NotNil(t, bp, "should find E2E p95 breaking point")
	assert.Equal(t, 8, bp.BreaksAtGames)
	assert.InDelta(t, 0.35, bp.LastGoodValue, 0.001)
	assert.InDelta(t, 0.65, bp.BreakValue, 0.001)
}

func TestFindBreakingPoints_NoViolations(t *testing.T) {
	t.Parallel()

	runs := []baseline.LevelResult{
		{Games: 2, Metrics: baseline.MetricsSnapshot{
			E2E: baseline.LatencyProfile{P95: 0.1, P99: 0.2},
		}},
		{Games: 4, Metrics: baseline.MetricsSnapshot{
			E2E: baseline.LatencyProfile{P95: 0.2, P99: 0.4},
		}},
	}

	slos := baseline.DefaultSLOs()
	breakingPoints := baseline.FindBreakingPoints(runs, slos)
	assert.Empty(t, breakingPoints)
}

func TestFindBreakingPoints_MultipleBreaks(t *testing.T) {
	t.Parallel()

	runs := []baseline.LevelResult{
		{Games: 2, Metrics: baseline.MetricsSnapshot{
			E2E: baseline.LatencyProfile{P95: 0.3, P99: 0.5},
		}},
		{Games: 8, Metrics: baseline.MetricsSnapshot{
			E2E: baseline.LatencyProfile{P95: 0.6, P99: 1.2},
		}},
	}

	slos := baseline.DefaultSLOs()
	breakingPoints := baseline.FindBreakingPoints(runs, slos)

	assert.Len(t, breakingPoints, 2)
	assert.NotNil(t, findBP(breakingPoints, "E2E move latency p95"))
	assert.NotNil(t, findBP(breakingPoints, "E2E move latency p99"))
}

func findBP(breakingPoints []baseline.BreakingPoint, sloName string) *baseline.BreakingPoint {
	for i := range breakingPoints {
		if breakingPoints[i].SLOName == sloName {
			return &breakingPoints[i]
		}
	}

	return nil
}
