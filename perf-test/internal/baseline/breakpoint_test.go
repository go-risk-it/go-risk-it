package baseline_test

import (
	"testing"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/baseline"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindBreakingPoints_DetectsDBSaturation(t *testing.T) {
	t.Parallel()

	runs := []baseline.LevelResult{
		{Games: 2, Metrics: baseline.MetricsSnapshot{DBTxnP95: 0.02}},
		{Games: 4, Metrics: baseline.MetricsSnapshot{DBTxnP95: 0.035}},
		{Games: 8, Metrics: baseline.MetricsSnapshot{DBTxnP95: 0.065}},
	}

	slos := baseline.DefaultSLOs()
	breakingPoints := baseline.FindBreakingPoints(runs, slos)

	dbBP := findBP(breakingPoints, "DB transaction latency")
	require.NotNil(t, dbBP, "should find DB breaking point")
	assert.Equal(t, 8, dbBP.BreaksAtGames)
	assert.InDelta(t, 0.035, dbBP.LastGoodValue, 0.001)
	assert.InDelta(t, 0.065, dbBP.BreakValue, 0.001)
}

func TestFindBreakingPoints_NoViolations(t *testing.T) {
	t.Parallel()

	runs := []baseline.LevelResult{
		{Games: 2, Metrics: baseline.MetricsSnapshot{DBTxnP95: 0.01, E2EP95: 0.1}},
		{Games: 4, Metrics: baseline.MetricsSnapshot{DBTxnP95: 0.02, E2EP95: 0.2}},
	}

	slos := baseline.DefaultSLOs()
	breakingPoints := baseline.FindBreakingPoints(runs, slos)
	assert.Empty(t, breakingPoints)
}

func TestFindBreakingPoints_MultipleBreaks(t *testing.T) {
	t.Parallel()

	runs := []baseline.LevelResult{
		{Games: 2, Metrics: baseline.MetricsSnapshot{DBTxnP95: 0.02, E2EP95: 0.3}},
		{Games: 8, Metrics: baseline.MetricsSnapshot{DBTxnP95: 0.08, E2EP95: 0.6}},
	}

	slos := baseline.DefaultSLOs()
	breakingPoints := baseline.FindBreakingPoints(runs, slos)

	assert.Len(t, breakingPoints, 2)
	assert.NotNil(t, findBP(breakingPoints, "DB transaction latency"))
	assert.NotNil(t, findBP(breakingPoints, "E2E move latency"))
}

func findBP(breakingPoints []baseline.BreakingPoint, sloName string) *baseline.BreakingPoint {
	for i := range breakingPoints {
		if breakingPoints[i].SLOName == sloName {
			return &breakingPoints[i]
		}
	}

	return nil
}
