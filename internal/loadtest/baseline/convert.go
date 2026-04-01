package baseline

import "github.com/go-risk-it/go-risk-it/internal/loadtest/metrics"

// histToProfile converts a collector HistogramSnapshot (milliseconds) to a LatencyProfile (seconds).
func histToProfile(h metrics.HistogramSnapshot) LatencyProfile {
	return LatencyProfile{
		P50: float64(h.P50) / 1000.0,
		P95: float64(h.P95) / 1000.0,
		P99: float64(h.P99) / 1000.0,
		Max: float64(h.Max) / 1000.0,
	}
}

// SnapshotToMetrics converts a collector Snapshot into a baseline MetricsSnapshot.
// Latency values are converted from milliseconds (HDR histogram) to seconds (baseline).
func SnapshotToMetrics(snap *metrics.Snapshot, totalDurationSec float64) MetricsSnapshot {
	var errorRate float64
	if snap.TotalMoves > 0 {
		errorRate = float64(snap.TotalErrors) / float64(snap.TotalMoves)
	}

	var throughput float64
	if totalDurationSec > 0 {
		throughput = float64(snap.TotalMoves) / totalDurationSec
	}

	// Find peak throughput from time-series buckets.
	var peakMPS float64
	for _, bucket := range snap.ThroughputBuckets {
		bucketMPS := float64(bucket.Moves) / 5.0 // 5-second buckets
		if bucketMPS > peakMPS {
			peakMPS = bucketMPS
		}
	}

	// Convert phase latency histograms.
	var phaseLatency map[string]LatencyProfile
	if len(snap.PhaseLatency) > 0 {
		phaseLatency = make(map[string]LatencyProfile, len(snap.PhaseLatency))
		for phase, hist := range snap.PhaseLatency {
			phaseLatency[phase] = histToProfile(hist)
		}
	}

	// Convert REST per-action latency histograms.
	var restLatency map[string]LatencyProfile
	if len(snap.RESTLatency) > 0 {
		restLatency = make(map[string]LatencyProfile, len(snap.RESTLatency))
		for action, hist := range snap.RESTLatency {
			restLatency[action] = histToProfile(hist)
		}
	}

	// Copy error breakdown, omitting zero entries.
	errorBreakdown := copyNonZeroMap(snap.ErrorBreakdown)

	// Copy phase flow, omitting zero entries.
	phaseEntries := copyNonZeroMap(snap.PhaseEntries)
	phaseMoves := copyNonZeroMap(snap.PhaseMoves)

	// Compute move failure rate from ErrorBreakdown.
	// TotalMoves counts successful moves only, so total attempts = TotalMoves + sum(ErrorBreakdown).
	var moveFailureRate float64
	if len(snap.ErrorBreakdown) > 0 {
		var totalNonSuccess int64
		for _, count := range snap.ErrorBreakdown {
			totalNonSuccess += count
		}

		if totalAttempts := snap.TotalMoves + totalNonSuccess; totalAttempts > 0 {
			moveFailureRate = float64(totalNonSuccess) / float64(totalAttempts)
		}
	}

	return MetricsSnapshot{
		E2E:                    histToProfile(snap.E2EMove),
		WSDelivery:             histToProfile(snap.WSDelivery),
		ThroughputMPS:          throughput,
		ThroughputPeakMPS:      peakMPS,
		HTTPErrorRate:          errorRate,
		MoveFailureRate:        moveFailureRate,
		TotalMoves:             snap.TotalMoves,
		TotalErrors:            snap.TotalErrors,
		GamesCompleted:         snap.GamesCompleted,
		GamesTimedOut:          snap.GamesTimedOut,
		GamesFatal:             snap.GamesFatal,
		TotalRetries:           snap.TotalRetries,
		TotalConflicts:         snap.TotalConflicts,
		TotalReconnects:        snap.TotalReconnects,
		TotalReconnectFailures: snap.TotalReconnectFailures,
		PhaseLatency:           phaseLatency,
		RESTLatency:            restLatency,
		ErrorBreakdown:         errorBreakdown,
		PhaseEntries:           phaseEntries,
		PhaseMoves:             phaseMoves,
		DurationSec:            totalDurationSec,
	}
}

func copyNonZeroMap(src map[string]int64) map[string]int64 {
	if len(src) == 0 {
		return nil
	}

	dst := make(map[string]int64, len(src))
	for k, v := range src {
		if v > 0 {
			dst[k] = v
		}
	}

	if len(dst) == 0 {
		return nil
	}

	return dst
}
