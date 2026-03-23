package baseline_test

import (
	"bytes"
	"testing"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/baseline"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompare_ShowsDeltas(t *testing.T) {
	t.Parallel()

	before := baseline.MetricsSnapshot{
		E2E:           baseline.LatencyProfile{P50: 0.200, P95: 0.312, P99: 0.450},
		ThroughputMPS: 45.2,
	}
	after := baseline.MetricsSnapshot{
		E2E:           baseline.LatencyProfile{P50: 0.180, P95: 0.287, P99: 0.410},
		ThroughputMPS: 52.1,
	}

	deltas := baseline.Compare(before, after)
	require.NotEmpty(t, deltas)

	e2eDelta := findDelta(deltas, "E2E p95")
	require.NotNil(t, e2eDelta)
	assert.InDelta(t, -8.0, e2eDelta.ChangePercent, 1.0)
	assert.True(t, e2eDelta.Improved)

	throughputDelta := findDelta(deltas, "Throughput (avg)")
	require.NotNil(t, throughputDelta)
	assert.True(t, throughputDelta.Improved)
}

func TestCompare_NoChange(t *testing.T) {
	t.Parallel()

	snap := baseline.MetricsSnapshot{E2E: baseline.LatencyProfile{P95: 0.5}}
	deltas := baseline.Compare(snap, snap)

	for _, delta := range deltas {
		assert.InDelta(t, 0.0, delta.ChangePercent, 0.001)
	}
}

func TestCompare_IncludesAllMetrics(t *testing.T) {
	t.Parallel()

	snap := baseline.MetricsSnapshot{
		E2E:               baseline.LatencyProfile{P50: 0.1, P95: 0.2, P99: 0.3},
		WSDelivery:        baseline.LatencyProfile{P95: 0.05},
		HTTPErrorRate:     0.01,
		ThroughputMPS:     10.0,
		ThroughputPeakMPS: 20.0,
		TotalConflicts:    5,
		TotalRetries:      3,
	}

	deltas := baseline.Compare(snap, snap)

	names := make([]string, len(deltas))
	for i, d := range deltas {
		names[i] = d.Name
	}

	assert.Contains(t, names, "E2E p50")
	assert.Contains(t, names, "E2E p95")
	assert.Contains(t, names, "E2E p99")
	assert.Contains(t, names, "WS Delivery p95")
	assert.Contains(t, names, "HTTP Error Rate")
	assert.Contains(t, names, "Move Failure Rate")
	assert.Contains(t, names, "Throughput (avg)")
	assert.Contains(t, names, "Throughput (peak)")
	assert.Contains(t, names, "Conflicts")
	assert.Contains(t, names, "Retries")
}

func TestPrintComparison_WithPhases(t *testing.T) {
	t.Parallel()

	before := baseline.Baseline{
		CommitSHA: "abc1234",
		Metrics: baseline.MetricsSnapshot{
			E2E: baseline.LatencyProfile{P95: 0.3},
			PhaseLatency: map[string]baseline.LatencyProfile{
				"deploy": {P95: 0.2},
				"attack": {P95: 0.4},
			},
		},
	}
	after := baseline.Baseline{
		CommitSHA: "def5678",
		Metrics: baseline.MetricsSnapshot{
			E2E: baseline.LatencyProfile{P95: 0.25},
			PhaseLatency: map[string]baseline.LatencyProfile{
				"deploy": {P95: 0.15},
				"attack": {P95: 0.35},
			},
		},
	}

	var buf bytes.Buffer
	baseline.PrintComparison(&buf, before, after)

	output := buf.String()
	assert.Contains(t, output, "Phase latency comparison")
	assert.Contains(t, output, "deploy")
	assert.Contains(t, output, "attack")
}

func TestPrintComparison_NoPhaseTableWhenMissing(t *testing.T) {
	t.Parallel()

	before := baseline.Baseline{
		CommitSHA: "abc1234",
		Metrics:   baseline.MetricsSnapshot{E2E: baseline.LatencyProfile{P95: 0.3}},
	}
	after := baseline.Baseline{
		CommitSHA: "def5678",
		Metrics:   baseline.MetricsSnapshot{E2E: baseline.LatencyProfile{P95: 0.25}},
	}

	var buf bytes.Buffer
	baseline.PrintComparison(&buf, before, after)

	assert.NotContains(t, buf.String(), "Phase latency comparison")
}

func findDelta(deltas []baseline.Delta, name string) *baseline.Delta {
	for i := range deltas {
		if deltas[i].Name == name {
			return &deltas[i]
		}
	}

	return nil
}
