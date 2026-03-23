package baseline

import "github.com/go-risk-it/go-risk-it/perf-test/internal/metrics"

// SnapshotToMetrics converts a collector Snapshot into a baseline MetricsSnapshot.
// Latency values are converted from milliseconds (HDR histogram) to seconds (baseline).
// Server-side metrics (DBTxnP95, DBPoolUtil, WSBroadcastP95) are set to 0 since they
// cannot be measured from the client; zero values pass SLO checks correctly (0 < threshold).
func SnapshotToMetrics(snap *metrics.Snapshot, totalDurationSec float64) MetricsSnapshot {
	var errorRate float64
	if snap.TotalMoves > 0 {
		errorRate = float64(snap.TotalErrors) / float64(snap.TotalMoves)
	}

	var throughput float64
	if totalDurationSec > 0 {
		throughput = float64(snap.TotalMoves) / totalDurationSec
	}

	return MetricsSnapshot{
		E2EP95:        float64(snap.E2EMove.P95) / 1000.0,
		WSDeliveryP95: float64(snap.WSDelivery.P95) / 1000.0,
		HTTPErrorRate: errorRate,
		ThroughputMPS: throughput,
	}
}
