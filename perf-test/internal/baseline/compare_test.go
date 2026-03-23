package baseline_test

import (
	"testing"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/baseline"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompare_ShowsDeltas(t *testing.T) {
	t.Parallel()

	before := baseline.MetricsSnapshot{
		E2EP95:        0.312,
		DBTxnP95:      0.042,
		ThroughputMPS: 45.2,
	}
	after := baseline.MetricsSnapshot{
		E2EP95:        0.287,
		DBTxnP95:      0.038,
		ThroughputMPS: 52.1,
	}

	deltas := baseline.Compare(before, after)
	require.NotEmpty(t, deltas)

	e2eDelta := findDelta(deltas, "E2E p95")
	require.NotNil(t, e2eDelta)
	assert.InDelta(t, -8.0, e2eDelta.ChangePercent, 1.0)
	assert.True(t, e2eDelta.Improved)

	throughputDelta := findDelta(deltas, "Throughput")
	require.NotNil(t, throughputDelta)
	assert.True(t, throughputDelta.Improved)
}

func TestCompare_NoChange(t *testing.T) {
	t.Parallel()

	snap := baseline.MetricsSnapshot{E2EP95: 0.5}
	deltas := baseline.Compare(snap, snap)

	for _, delta := range deltas {
		assert.InDelta(t, 0.0, delta.ChangePercent, 0.001)
	}
}

func findDelta(deltas []baseline.Delta, name string) *baseline.Delta {
	for i := range deltas {
		if deltas[i].Name == name {
			return &deltas[i]
		}
	}

	return nil
}
