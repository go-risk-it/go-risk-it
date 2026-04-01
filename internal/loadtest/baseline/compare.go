package baseline

import (
	"fmt"
	"io"
	"sort"
	"strconv"
	"text/tabwriter"
)

// Delta describes the change in a single metric between two snapshots.
type Delta struct {
	Name          string
	Before        float64
	After         float64
	ChangePercent float64
	Improved      bool
	Unit          string
}

// metricDef describes a metric for comparison purposes.
type metricDef struct {
	name     string
	unit     string
	extract  func(MetricsSnapshot) float64
	lowerBad bool // true if decrease is bad (e.g., throughput)
}

func comparisonMetrics() []metricDef {
	return []metricDef{
		{"E2E p50", "s", func(s MetricsSnapshot) float64 { return s.E2E.P50 }, false},
		{"E2E p95", "s", func(s MetricsSnapshot) float64 { return s.E2E.P95 }, false},
		{"E2E p99", "s", func(s MetricsSnapshot) float64 { return s.E2E.P99 }, false},
		{
			"WS Delivery p95",
			"s",
			func(s MetricsSnapshot) float64 { return s.WSDelivery.P95 },
			false,
		},
		{
			"HTTP Error Rate",
			"ratio",
			func(s MetricsSnapshot) float64 { return s.HTTPErrorRate },
			false,
		},
		{
			"Move Failure Rate",
			"ratio",
			func(s MetricsSnapshot) float64 { return s.MoveFailureRate },
			false,
		},
		{
			"Throughput (avg)",
			"moves/s",
			func(s MetricsSnapshot) float64 { return s.ThroughputMPS },
			true,
		},
		{
			"Throughput (peak)",
			"moves/s",
			func(s MetricsSnapshot) float64 { return s.ThroughputPeakMPS },
			true,
		},
		{
			"Conflicts",
			"count",
			func(s MetricsSnapshot) float64 { return float64(s.TotalConflicts) },
			false,
		},
		{
			"Retries",
			"count",
			func(s MetricsSnapshot) float64 { return float64(s.TotalRetries) },
			false,
		},
	}
}

// Compare computes deltas between two MetricsSnapshots.
func Compare(before, after MetricsSnapshot) []Delta {
	metrics := comparisonMetrics()
	deltas := make([]Delta, 0, len(metrics))

	for _, metricDefinition := range metrics {
		beforeVal := metricDefinition.extract(before)
		afterVal := metricDefinition.extract(after)

		var changePct float64
		if beforeVal != 0 {
			changePct = ((afterVal - beforeVal) / beforeVal) * 100
		}

		improved := changePct < 0
		if metricDefinition.lowerBad {
			improved = changePct > 0
		}

		deltas = append(deltas, Delta{
			Name:          metricDefinition.name,
			Before:        beforeVal,
			After:         afterVal,
			ChangePercent: changePct,
			Improved:      improved,
			Unit:          metricDefinition.unit,
		})
	}

	return deltas
}

// PrintComparison formats a comparison table between two baselines.
func PrintComparison(writer io.Writer, before, after Baseline) {
	fmt.Fprintf(writer, "Baseline comparison: %s → %s\n\n", before.CommitSHA, after.CommitSHA)

	printEnvironmentWarnings(writer, before.Environment, after.Environment)

	printMainComparison(writer, before.Metrics, after.Metrics)
	printPhaseComparison(writer, before.Metrics, after.Metrics)
}

func printMainComparison(writer io.Writer, before, after MetricsSnapshot) {
	tabWriter := tabwriter.NewWriter(writer, 0, 0, 2, ' ', 0)

	fmt.Fprintln(tabWriter, "METRIC\tBEFORE\tAFTER\tCHANGE\tSTATUS")
	fmt.Fprintln(tabWriter, "------\t------\t-----\t------\t------")

	for _, delta := range Compare(before, after) {
		status := deltaStatus(delta)

		fmt.Fprintf(
			tabWriter,
			"%s\t%.4f %s\t%.4f %s\t%+.1f%%\t%s\n",
			delta.Name,
			delta.Before, delta.Unit,
			delta.After, delta.Unit,
			delta.ChangePercent,
			status,
		)
	}

	tabWriter.Flush()
}

//nolint:cyclop // sequential report formatting
func printPhaseComparison(writer io.Writer, before, after MetricsSnapshot) {
	if len(before.PhaseLatency) == 0 || len(after.PhaseLatency) == 0 {
		return
	}

	// Collect all phase names from both snapshots.
	phases := make(map[string]bool)
	for p := range before.PhaseLatency {
		phases[p] = true
	}

	for p := range after.PhaseLatency {
		phases[p] = true
	}

	sorted := make([]string, 0, len(phases))
	for p := range phases {
		sorted = append(sorted, p)
	}

	sort.Strings(sorted)

	fmt.Fprintf(writer, "\nPhase latency comparison (p95):\n")

	tabWriter := tabwriter.NewWriter(writer, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tabWriter, "PHASE\tBEFORE\tAFTER\tCHANGE\tSTATUS")
	fmt.Fprintln(tabWriter, "-----\t------\t-----\t------\t------")

	for _, phase := range sorted {
		beforeP95 := before.PhaseLatency[phase].P95
		afterP95 := after.PhaseLatency[phase].P95

		var changePct float64
		if beforeP95 != 0 {
			changePct = ((afterP95 - beforeP95) / beforeP95) * 100
		}

		improved := changePct < 0
		status := "="

		if changePct > 1 || changePct < -1 {
			if improved {
				status = "BETTER"
			} else {
				status = "WORSE"
			}
		}

		fmt.Fprintf(
			tabWriter,
			"%s\t%.4f s\t%.4f s\t%+.1f%%\t%s\n",
			phase,
			beforeP95,
			afterP95,
			changePct,
			status,
		)
	}

	tabWriter.Flush()
}

func deltaStatus(delta Delta) string {
	if delta.ChangePercent > 1 || delta.ChangePercent < -1 {
		if delta.Improved {
			return "BETTER"
		}

		return "WORSE"
	}

	return "="
}

// printEnvironmentWarnings prints a warning block if key environment fields differ.
//
//nolint:cyclop // sequential environment diff checks
func printEnvironmentWarnings(writer io.Writer, before, after Environment) {
	type envDiff struct {
		field     string
		beforeVal string
		afterVal  string
	}

	var diffs []envDiff

	if before.GOOS != after.GOOS && before.GOOS != "" {
		diffs = append(diffs, envDiff{"GOOS", before.GOOS, after.GOOS})
	}

	if before.GOARCH != after.GOARCH && before.GOARCH != "" {
		diffs = append(diffs, envDiff{"GOARCH", before.GOARCH, after.GOARCH})
	}

	if before.NumCPU != after.NumCPU && before.NumCPU != 0 {
		diffs = append(diffs, envDiff{
			"NumCPU",
			strconv.Itoa(before.NumCPU),
			strconv.Itoa(after.NumCPU),
		})
	}

	if before.GOMAXPROCS != after.GOMAXPROCS && before.GOMAXPROCS != 0 {
		diffs = append(diffs, envDiff{
			"GOMAXPROCS",
			strconv.Itoa(before.GOMAXPROCS),
			strconv.Itoa(after.GOMAXPROCS),
		})
	}

	if len(diffs) == 0 {
		return
	}

	fmt.Fprintln(writer, "WARNING: Environment differences detected:")

	for _, d := range diffs {
		fmt.Fprintf(writer, "  %s: %s → %s\n", d.field, d.beforeVal, d.afterVal)
	}

	fmt.Fprintln(writer)
}
