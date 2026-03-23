package baseline_test

import (
	"bytes"
	"testing"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/baseline"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalyze_NoInsights(t *testing.T) {
	t.Parallel()

	snap := baseline.MetricsSnapshot{
		E2E:           baseline.LatencyProfile{P50: 0.1, P95: 0.2, P99: 0.3},
		WSDelivery:    baseline.LatencyProfile{P50: 0.05, P95: 0.1, P99: 0.15},
		ThroughputMPS: 10.0,
		TotalMoves:    100,
	}

	insights := baseline.Analyze(snap)
	assert.Empty(t, insights)
}

func TestAnalyze_FatalGames(t *testing.T) {
	t.Parallel()

	snap := baseline.MetricsSnapshot{GamesFatal: 2}

	insights := baseline.Analyze(snap)
	require.Len(t, insights, 1)
	assert.Equal(t, "critical", insights[0].Severity)
	assert.Equal(t, "Fatal games", insights[0].Title)
}

func TestAnalyze_HighTimeoutRate(t *testing.T) {
	t.Parallel()

	snap := baseline.MetricsSnapshot{
		GamesCompleted: 5,
		GamesTimedOut:  3,
	}

	insights := baseline.Analyze(snap)
	found := findInsight(insights, "High timeout rate")
	require.NotNil(t, found)
	assert.Equal(t, "warning", found.Severity)
}

func TestAnalyze_TimeoutRateBelowThreshold(t *testing.T) {
	t.Parallel()

	snap := baseline.MetricsSnapshot{
		GamesCompleted: 10,
		GamesTimedOut:  1,
	}

	insights := baseline.Analyze(snap)
	assert.Nil(t, findInsight(insights, "High timeout rate"))
}

func TestAnalyze_TailLatencyBlowup(t *testing.T) {
	t.Parallel()

	snap := baseline.MetricsSnapshot{
		E2E:        baseline.LatencyProfile{P95: 0.1, P99: 0.5},
		WSDelivery: baseline.LatencyProfile{P95: 0.05, P99: 0.2},
	}

	insights := baseline.Analyze(snap)
	assert.NotNil(t, findInsight(insights, "E2E tail latency blow-up"))
	assert.NotNil(t, findInsight(insights, "WS delivery tail latency blow-up"))
}

func TestAnalyze_NoTailLatencyBlowup(t *testing.T) {
	t.Parallel()

	snap := baseline.MetricsSnapshot{
		E2E:        baseline.LatencyProfile{P95: 0.1, P99: 0.15},
		WSDelivery: baseline.LatencyProfile{P95: 0.05, P99: 0.07},
	}

	insights := baseline.Analyze(snap)
	assert.Nil(t, findInsight(insights, "E2E tail latency blow-up"))
	assert.Nil(t, findInsight(insights, "WS delivery tail latency blow-up"))
}

func TestAnalyze_HighRetryRate(t *testing.T) {
	t.Parallel()

	snap := baseline.MetricsSnapshot{
		TotalMoves:   100,
		TotalRetries: 10,
	}

	insights := baseline.Analyze(snap)
	found := findInsight(insights, "High retry rate")
	require.NotNil(t, found)
	assert.Equal(t, "warning", found.Severity)
}

func TestAnalyze_ConflictStorm(t *testing.T) {
	t.Parallel()

	snap := baseline.MetricsSnapshot{
		TotalMoves:     100,
		TotalConflicts: 5,
	}

	insights := baseline.Analyze(snap)
	found := findInsight(insights, "Conflict storm")
	require.NotNil(t, found)
	assert.Equal(t, "bottleneck", found.Category)
}

func TestAnalyze_ErrorDominance(t *testing.T) {
	t.Parallel()

	snap := baseline.MetricsSnapshot{
		TotalErrors:    10,
		ErrorBreakdown: map[string]int64{"timeout": 8, "execution": 2},
	}

	insights := baseline.Analyze(snap)
	found := findInsight(insights, "Error dominance")
	require.NotNil(t, found)
	assert.Contains(t, found.Detail, "timeout")
	assert.Contains(t, found.Detail, "non-success outcomes")
}

func TestAnalyze_NoErrorDominance(t *testing.T) {
	t.Parallel()

	snap := baseline.MetricsSnapshot{
		TotalErrors:    10,
		ErrorBreakdown: map[string]int64{"timeout": 4, "execution": 3, "transient": 3},
	}

	insights := baseline.Analyze(snap)
	assert.Nil(t, findInsight(insights, "Error dominance"))
}

func TestAnalyze_ThroughputPlateau(t *testing.T) {
	t.Parallel()

	snap := baseline.MetricsSnapshot{
		ThroughputMPS:     5.0,
		ThroughputPeakMPS: 20.0,
	}

	insights := baseline.Analyze(snap)
	found := findInsight(insights, "Throughput plateau")
	require.NotNil(t, found)
	assert.Equal(t, "info", found.Severity)
}

func TestAnalyze_SlowPhase(t *testing.T) {
	t.Parallel()

	snap := baseline.MetricsSnapshot{
		E2E: baseline.LatencyProfile{P95: 0.2},
		PhaseLatency: map[string]baseline.LatencyProfile{
			"deploy": {P95: 0.1},
			"attack": {P95: 0.5},
		},
	}

	insights := baseline.Analyze(snap)
	found := findInsight(insights, "Slow phase")
	require.NotNil(t, found)
	assert.Contains(t, found.Detail, "attack")
}

func TestAnalyze_RESTHotspot(t *testing.T) {
	t.Parallel()

	snap := baseline.MetricsSnapshot{
		RESTLatency: map[string]baseline.LatencyProfile{
			"deploy":    {P95: 0.1},
			"attack":    {P95: 0.1},
			"reinforce": {P95: 0.5},
		},
	}

	insights := baseline.Analyze(snap)
	found := findInsight(insights, "REST action hotspot")
	require.NotNil(t, found)
	assert.Contains(t, found.Detail, "reinforce")
}

func TestAnalyze_PhaseFlowImbalance(t *testing.T) {
	t.Parallel()

	snap := baseline.MetricsSnapshot{
		PhaseEntries: map[string]int64{
			"deploy": 10,
			"attack": 50,
		},
	}

	insights := baseline.Analyze(snap)
	found := findInsight(insights, "Phase flow imbalance")
	require.NotNil(t, found)
	assert.Equal(t, "info", found.Severity)
}

func TestAnalyze_NoPhaseFlowImbalance(t *testing.T) {
	t.Parallel()

	snap := baseline.MetricsSnapshot{
		PhaseEntries: map[string]int64{
			"deploy": 40,
			"attack": 50,
		},
	}

	insights := baseline.Analyze(snap)
	assert.Nil(t, findInsight(insights, "Phase flow imbalance"))
}

func TestPrintInsights_Empty(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	baseline.PrintInsights(&buf, nil)
	assert.Contains(t, buf.String(), "all clear")
}

func TestPrintInsights_NonEmpty(t *testing.T) {
	t.Parallel()

	insights := []baseline.Insight{
		{Category: "health", Severity: "critical", Title: "Fatal games", Detail: "2 fatal"},
	}

	var buf bytes.Buffer
	baseline.PrintInsights(&buf, insights)
	assert.Contains(t, buf.String(), "1 detected")
	assert.Contains(t, buf.String(), "Fatal games")
}

func TestAnalyze_ErrorDominance_WithContention(t *testing.T) {
	t.Parallel()

	// Contention categories are included in the denominator.
	snap := baseline.MetricsSnapshot{
		TotalErrors:    2,
		ErrorBreakdown: map[string]int64{"conflict": 80, "execution": 2},
	}

	insights := baseline.Analyze(snap)
	found := findInsight(insights, "Error dominance")
	require.NotNil(t, found)
	assert.Contains(t, found.Detail, "conflict")
}

func TestAnalyze_HighContentionRate_Warning(t *testing.T) {
	t.Parallel()

	// 15% contention rate (above 10% warning threshold).
	snap := baseline.MetricsSnapshot{
		TotalMoves:     850,
		ErrorBreakdown: map[string]int64{"conflict": 100, "stale_state": 50},
	}

	insights := baseline.Analyze(snap)
	found := findInsight(insights, "High contention rate")
	require.NotNil(t, found)
	assert.Equal(t, "warning", found.Severity)
	assert.Contains(t, found.Detail, "conflicts=100")
	assert.Contains(t, found.Detail, "stale_state=50")
}

func TestAnalyze_HighContentionRate_Critical(t *testing.T) {
	t.Parallel()

	// 30% contention rate (above 25% critical threshold).
	snap := baseline.MetricsSnapshot{
		TotalMoves:     700,
		ErrorBreakdown: map[string]int64{"conflict": 250, "stale_state": 50},
	}

	insights := baseline.Analyze(snap)
	found := findInsight(insights, "High contention rate")
	require.NotNil(t, found)
	assert.Equal(t, "critical", found.Severity)
}

func TestAnalyze_LowContentionRate_NoInsight(t *testing.T) {
	t.Parallel()

	// 5% contention rate (below 10% warning threshold).
	snap := baseline.MetricsSnapshot{
		TotalMoves:     950,
		ErrorBreakdown: map[string]int64{"conflict": 40, "stale_state": 10},
	}

	insights := baseline.Analyze(snap)
	assert.Nil(t, findInsight(insights, "High contention rate"))
}

func findInsight(insights []baseline.Insight, title string) *baseline.Insight {
	for i := range insights {
		if insights[i].Title == title {
			return &insights[i]
		}
	}

	return nil
}
