package baseline_test

import (
	"testing"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/baseline"
	"github.com/stretchr/testify/assert"
)

func TestSLOs_Evaluate_AllGreen(t *testing.T) {
	t.Parallel()

	result := baseline.DefaultSLOs().Evaluate(baseline.MetricsSnapshot{
		E2E:           baseline.LatencyProfile{P95: 0.3, P99: 0.6},
		WSDelivery:    baseline.LatencyProfile{P95: 0.1, P99: 0.3},
		HTTPErrorRate: 0.005,
	})

	assert.True(t, result.AllPassing())
	assert.Empty(t, result.Violations)
}

func TestSLOs_Evaluate_WithViolations(t *testing.T) {
	t.Parallel()

	result := baseline.DefaultSLOs().Evaluate(baseline.MetricsSnapshot{
		E2E:           baseline.LatencyProfile{P95: 0.6, P99: 1.2},
		WSDelivery:    baseline.LatencyProfile{P95: 0.1, P99: 0.3},
		HTTPErrorRate: 0.005,
	})

	assert.False(t, result.AllPassing())
	assert.Len(t, result.Violations, 2)

	names := make([]string, len(result.Violations))
	for i, violation := range result.Violations {
		names[i] = violation.SLO.Name
	}

	assert.Contains(t, names, "E2E move latency p95")
	assert.Contains(t, names, "E2E move latency p99")
}

func TestSLOs_Evaluate_P99Violations(t *testing.T) {
	t.Parallel()

	result := baseline.DefaultSLOs().Evaluate(baseline.MetricsSnapshot{
		E2E:           baseline.LatencyProfile{P95: 0.4, P99: 0.8},
		WSDelivery:    baseline.LatencyProfile{P95: 0.15, P99: 0.6},
		HTTPErrorRate: 0.005,
	})

	assert.False(t, result.AllPassing())

	names := make([]string, len(result.Violations))
	for i, violation := range result.Violations {
		names[i] = violation.SLO.Name
	}

	assert.Contains(t, names, "WS delivery latency p99")
}

func TestSLOs_BoundaryHealth(t *testing.T) {
	t.Parallel()

	slos := baseline.DefaultSLOs()
	assert.Len(t, slos.BoundaryHealth, 1)
	assert.Equal(t, "move_failure_rate", slos.BoundaryHealth[0].Metric)
}
