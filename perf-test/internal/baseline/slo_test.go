package baseline_test

import (
	"testing"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/baseline"
	"github.com/stretchr/testify/assert"
)

func TestSLOs_Evaluate_AllGreen(t *testing.T) {
	t.Parallel()

	result := baseline.DefaultSLOs().Evaluate(baseline.MetricsSnapshot{
		E2EP95:         0.3,
		WSDeliveryP95:  0.1,
		DBTxnP95:       0.03,
		DBPoolUtil:     0.5,
		WSBroadcastP95: 0.06,
		HTTPErrorRate:  0.005,
	})

	assert.True(t, result.AllPassing())
	assert.Empty(t, result.Violations)
}

func TestSLOs_Evaluate_WithViolations(t *testing.T) {
	t.Parallel()

	result := baseline.DefaultSLOs().Evaluate(baseline.MetricsSnapshot{
		E2EP95:         0.6,
		WSDeliveryP95:  0.1,
		DBTxnP95:       0.08,
		DBPoolUtil:     0.5,
		WSBroadcastP95: 0.06,
		HTTPErrorRate:  0.005,
	})

	assert.False(t, result.AllPassing())
	assert.Len(t, result.Violations, 2)

	names := make([]string, len(result.Violations))
	for i, violation := range result.Violations {
		names[i] = violation.SLO.Name
	}

	assert.Contains(t, names, "E2E move latency")
	assert.Contains(t, names, "DB transaction latency")
}
