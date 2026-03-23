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
		E2EMove:     metrics.HistogramSnapshot{P95: 350}, // 350ms
		WSDelivery:  metrics.HistogramSnapshot{P95: 120}, // 120ms
		TotalMoves:  1000,
		TotalErrors: 5,
	}

	m := baseline.SnapshotToMetrics(snap, 60.0)

	assert.InDelta(t, 0.350, m.E2EP95, 0.0001, "E2E p95 should convert ms to s")
	assert.InDelta(t, 0.120, m.WSDeliveryP95, 0.0001, "WS delivery p95 should convert ms to s")
	assert.InDelta(t, 0.005, m.HTTPErrorRate, 0.0001, "error rate = 5/1000")
	assert.InDelta(t, 16.6667, m.ThroughputMPS, 0.001, "throughput = 1000/60")
	assert.Equal(t, 0.0, m.DBTxnP95, "server-side metric should be zero")
	assert.Equal(t, 0.0, m.DBPoolUtil, "server-side metric should be zero")
	assert.Equal(t, 0.0, m.WSBroadcastP95, "server-side metric should be zero")
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
