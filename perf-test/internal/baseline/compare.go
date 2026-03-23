package baseline

import (
	"fmt"
	"io"
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
		{"E2E p95", "s", func(s MetricsSnapshot) float64 { return s.E2EP95 }, false},
		{
			"WS Delivery p95",
			"s",
			func(s MetricsSnapshot) float64 { return s.WSDeliveryP95 },
			false,
		},
		{"DB Txn p95", "s", func(s MetricsSnapshot) float64 { return s.DBTxnP95 }, false},
		{"DB Pool Util", "ratio", func(s MetricsSnapshot) float64 { return s.DBPoolUtil }, false},
		{
			"WS Broadcast p95",
			"s",
			func(s MetricsSnapshot) float64 { return s.WSBroadcastP95 },
			false,
		},
		{
			"HTTP Error Rate",
			"ratio",
			func(s MetricsSnapshot) float64 { return s.HTTPErrorRate },
			false,
		},
		{
			"Throughput",
			"moves/s",
			func(s MetricsSnapshot) float64 { return s.ThroughputMPS },
			true,
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

	tabWriter := tabwriter.NewWriter(writer, 0, 0, 2, ' ', 0)

	fmt.Fprintln(tabWriter, "METRIC\tBEFORE\tAFTER\tCHANGE\tSTATUS")
	fmt.Fprintln(tabWriter, "------\t------\t-----\t------\t------")

	for _, delta := range Compare(before.Metrics, after.Metrics) {
		status := "="
		if delta.ChangePercent > 1 {
			if delta.Improved {
				status = "BETTER"
			} else {
				status = "WORSE"
			}
		} else if delta.ChangePercent < -1 {
			if delta.Improved {
				status = "BETTER"
			} else {
				status = "WORSE"
			}
		}

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

// printEnvironmentWarnings prints a warning block if key environment fields differ.
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
			fmt.Sprintf("%d", before.NumCPU),
			fmt.Sprintf("%d", after.NumCPU),
		})
	}

	if before.GOMAXPROCS != after.GOMAXPROCS && before.GOMAXPROCS != 0 {
		diffs = append(diffs, envDiff{
			"GOMAXPROCS",
			fmt.Sprintf("%d", before.GOMAXPROCS),
			fmt.Sprintf("%d", after.GOMAXPROCS),
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
