package baseline

import (
	"fmt"
	"io"
	"sort"
	"text/tabwriter"
)

// Insight detection thresholds.
const (
	// ThresholdTimeoutRate is the minimum timeout rate to flag as a warning.
	ThresholdTimeoutRate = 0.10
	// ThresholdTailLatencyRatio is the p99/p95 ratio above which tail blow-up is flagged.
	ThresholdTailLatencyRatio = 3.0
	// ThresholdRetryRate is the retry/move ratio above which a warning is emitted.
	ThresholdRetryRate = 0.05
	// ThresholdConflictRate is the conflict/move ratio above which a warning is emitted.
	ThresholdConflictRate = 0.03
	// ThresholdContentionCritical is the contention rate for a critical alert.
	ThresholdContentionCritical = 0.25
	// ThresholdContentionWarning is the contention rate for a warning.
	ThresholdContentionWarning = 0.10
	// ThresholdErrorDominance is the ratio above which a single error type dominates.
	ThresholdErrorDominance = 0.70
	// ThresholdThroughputPlateau is the peak/avg ratio for throughput plateau detection.
	ThresholdThroughputPlateau = 3.0
	// ThresholdSlowPhase is the multiplier of overall E2E p95 for flagging slow phases.
	ThresholdSlowPhase = 2.0
	// ThresholdRESTHotspot is the multiplier of median action p95 for flagging hotspots.
	ThresholdRESTHotspot = 3.0
	// ThresholdFlowImbalance is the deploy/attack entry ratio below which imbalance is flagged.
	ThresholdFlowImbalance = 0.5
)

// Insight describes a detected pattern from a load test run.
type Insight struct {
	Category string `json:"category"` // "bottleneck", "anomaly", "health"
	Severity string `json:"severity"` // "critical", "warning", "info"
	Title    string `json:"title"`
	Detail   string `json:"detail"`
}

// Analyze examines a MetricsSnapshot and returns detected patterns.
func Analyze(snap MetricsSnapshot) []Insight {
	insights := make([]Insight, 0, 8)

	insights = append(insights, detectFatalGames(snap)...)
	insights = append(insights, detectGameTimeouts(snap)...)
	insights = append(insights, detectTailLatencyBlowup(snap)...)
	insights = append(insights, detectHighRetryRate(snap)...)
	insights = append(insights, detectConflictStorm(snap)...)
	insights = append(insights, detectErrorDominance(snap)...)
	insights = append(insights, detectHighContentionRate(snap)...)
	insights = append(insights, detectThroughputPlateau(snap)...)
	insights = append(insights, detectSlowPhase(snap)...)
	insights = append(insights, detectRESTHotspot(snap)...)
	insights = append(insights, detectPhaseFlowImbalance(snap)...)

	return insights
}

// PrintInsights formats insights as a table.
func PrintInsights(w io.Writer, insights []Insight) {
	if len(insights) == 0 {
		fmt.Fprintln(w, "\nInsights: all clear — no anomalies detected")

		return
	}

	fmt.Fprintf(w, "\nInsights (%d detected):\n", len(insights))

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "SEVERITY\tCATEGORY\tFINDING")
	fmt.Fprintln(tw, "--------\t--------\t-------")

	for _, ins := range insights {
		fmt.Fprintf(tw, "%s\t%s\t%s: %s\n", ins.Severity, ins.Category, ins.Title, ins.Detail)
	}

	tw.Flush()
}

func detectFatalGames(snap MetricsSnapshot) []Insight {
	if snap.GamesFatal <= 0 {
		return nil
	}

	return []Insight{{
		Category: "health",
		Severity: "critical",
		Title:    "Fatal games",
		Detail:   fmt.Sprintf("%d game(s) had unrecoverable errors", snap.GamesFatal),
	}}
}

func detectGameTimeouts(snap MetricsSnapshot) []Insight {
	total := snap.GamesCompleted + snap.GamesTimedOut + snap.GamesFatal
	if total == 0 {
		return nil
	}

	rate := float64(snap.GamesTimedOut) / float64(total)
	if rate <= ThresholdTimeoutRate {
		return nil
	}

	return []Insight{{
		Category: "health",
		Severity: "warning",
		Title:    "High timeout rate",
		Detail: fmt.Sprintf(
			"%d/%d games timed out (%.0f%%)",
			snap.GamesTimedOut, total, rate*100,
		),
	}}
}

func detectTailLatencyBlowup(snap MetricsSnapshot) []Insight {
	insights := make([]Insight, 0, 8)

	if snap.E2E.P95 > 0 {
		ratio := snap.E2E.P99 / snap.E2E.P95
		if ratio > ThresholdTailLatencyRatio {
			insights = append(insights, Insight{
				Category: "anomaly",
				Severity: "warning",
				Title:    "E2E tail latency blow-up",
				Detail: fmt.Sprintf(
					"p99/p95 ratio = %.1fx (p95=%.3fs, p99=%.3fs)",
					ratio, snap.E2E.P95, snap.E2E.P99,
				),
			})
		}
	}

	if snap.WSDelivery.P95 > 0 {
		ratio := snap.WSDelivery.P99 / snap.WSDelivery.P95
		if ratio > ThresholdTailLatencyRatio {
			insights = append(insights, Insight{
				Category: "anomaly",
				Severity: "warning",
				Title:    "WS delivery tail latency blow-up",
				Detail: fmt.Sprintf(
					"p99/p95 ratio = %.1fx (p95=%.3fs, p99=%.3fs)",
					ratio, snap.WSDelivery.P95, snap.WSDelivery.P99,
				),
			})
		}
	}

	return insights
}

func detectHighRetryRate(snap MetricsSnapshot) []Insight {
	if snap.TotalMoves == 0 {
		return nil
	}

	rate := float64(snap.TotalRetries) / float64(snap.TotalMoves)
	if rate <= ThresholdRetryRate {
		return nil
	}

	return []Insight{{
		Category: "anomaly",
		Severity: "warning",
		Title:    "High retry rate",
		Detail: fmt.Sprintf(
			"%d retries / %d moves (%.1f%%)",
			snap.TotalRetries, snap.TotalMoves, rate*100,
		),
	}}
}

func detectConflictStorm(snap MetricsSnapshot) []Insight {
	if snap.TotalMoves == 0 {
		return nil
	}

	rate := float64(snap.TotalConflicts) / float64(snap.TotalMoves)
	if rate <= ThresholdConflictRate {
		return nil
	}

	return []Insight{{
		Category: "bottleneck",
		Severity: "warning",
		Title:    "Conflict storm",
		Detail: fmt.Sprintf(
			"%d conflicts / %d moves (%.1f%%)",
			snap.TotalConflicts, snap.TotalMoves, rate*100,
		),
	}}
}

func detectHighContentionRate(snap MetricsSnapshot) []Insight {
	conflicts := snap.ErrorBreakdown["conflict"]
	staleState := snap.ErrorBreakdown["stale_state"]
	contentionTotal := conflicts + staleState

	if contentionTotal == 0 {
		return nil
	}

	// Total attempts = successful moves + all non-success outcomes.
	var totalNonSuccess int64
	for _, count := range snap.ErrorBreakdown {
		totalNonSuccess += count
	}

	totalAttempts := snap.TotalMoves + totalNonSuccess
	if totalAttempts == 0 {
		return nil
	}

	rate := float64(contentionTotal) / float64(totalAttempts)

	// Check thresholds from highest to lowest severity.
	type level struct {
		threshold float64
		severity  string
	}

	for _, l := range []level{
		{ThresholdContentionCritical, "critical"},
		{ThresholdContentionWarning, "warning"},
	} {
		if rate > l.threshold {
			return []Insight{{
				Category: "bottleneck",
				Severity: l.severity,
				Title:    "High contention rate",
				Detail: fmt.Sprintf(
					//nolint:lll // long string literal
					"%.0f%% of move attempts are contention (conflicts=%d, stale_state=%d, total_attempts=%d)",
					rate*100,
					conflicts,
					staleState,
					totalAttempts,
				),
			}}
		}
	}

	return nil
}

func detectErrorDominance(snap MetricsSnapshot) []Insight {
	if len(snap.ErrorBreakdown) == 0 {
		return nil
	}

	var total int64
	for _, count := range snap.ErrorBreakdown {
		total += count
	}

	if total == 0 {
		return nil
	}

	for cat, count := range snap.ErrorBreakdown {
		ratio := float64(count) / float64(total)
		if ratio > ThresholdErrorDominance {
			return []Insight{{
				Category: "anomaly",
				Severity: "warning",
				Title:    "Error dominance",
				Detail: fmt.Sprintf(
					"%q accounts for %.0f%% of non-success outcomes (%d/%d)",
					cat, ratio*100, count, total,
				),
			}}
		}
	}

	return nil
}

func detectThroughputPlateau(snap MetricsSnapshot) []Insight {
	if snap.ThroughputMPS == 0 {
		return nil
	}

	ratio := snap.ThroughputPeakMPS / snap.ThroughputMPS
	if ratio <= ThresholdThroughputPlateau {
		return nil
	}

	return []Insight{{
		Category: "bottleneck",
		Severity: "info",
		Title:    "Throughput plateau",
		Detail: fmt.Sprintf(
			"peak/avg ratio = %.1fx (avg=%.1f, peak=%.1f moves/s)",
			ratio, snap.ThroughputMPS, snap.ThroughputPeakMPS,
		),
	}}
}

func detectSlowPhase(snap MetricsSnapshot) []Insight {
	if len(snap.PhaseLatency) == 0 || snap.E2E.P95 == 0 {
		return nil
	}

	threshold := snap.E2E.P95 * ThresholdSlowPhase
	insights := make([]Insight, 0, 8)

	for phase, profile := range snap.PhaseLatency {
		if profile.P95 > threshold {
			insights = append(insights, Insight{
				Category: "bottleneck",
				Severity: "warning",
				Title:    "Slow phase",
				Detail: fmt.Sprintf(
					"%q p95=%.3fs is %.1fx the overall E2E p95 (%.3fs)",
					phase, profile.P95, profile.P95/snap.E2E.P95, snap.E2E.P95,
				),
			})
		}
	}

	return insights
}

func detectRESTHotspot(snap MetricsSnapshot) []Insight {
	if len(snap.RESTLatency) < 2 {
		return nil
	}

	// Compute median p95.
	p95s := make([]float64, 0, len(snap.RESTLatency))
	for _, profile := range snap.RESTLatency {
		p95s = append(p95s, profile.P95)
	}

	sort.Float64s(p95s)

	median := p95s[len(p95s)/2]
	if median == 0 {
		return nil
	}

	threshold := median * ThresholdRESTHotspot
	insights := make([]Insight, 0, 8)

	for action, profile := range snap.RESTLatency {
		if profile.P95 > threshold {
			insights = append(insights, Insight{
				Category: "bottleneck",
				Severity: "warning",
				Title:    "REST action hotspot",
				Detail: fmt.Sprintf(
					"%q p95=%.3fs is %.1fx the median action p95 (%.3fs)",
					action, profile.P95, profile.P95/median, median,
				),
			})
		}
	}

	return insights
}

func detectPhaseFlowImbalance(snap MetricsSnapshot) []Insight {
	if len(snap.PhaseEntries) == 0 {
		return nil
	}

	deployEntries := snap.PhaseEntries["deploy"]
	attackEntries := snap.PhaseEntries["attack"]

	if attackEntries == 0 || deployEntries == 0 {
		return nil
	}

	// If deploy entries are less than half of attack entries, games are getting stuck.
	ratio := float64(deployEntries) / float64(attackEntries)
	if ratio >= ThresholdFlowImbalance {
		return nil
	}

	return []Insight{{
		Category: "anomaly",
		Severity: "info",
		Title:    "Phase flow imbalance",
		Detail: fmt.Sprintf(
			"deploy entries (%d) are %.0f%% of attack entries (%d) — games may be getting stuck",
			deployEntries, ratio*100, attackEntries,
		),
	}}
}
