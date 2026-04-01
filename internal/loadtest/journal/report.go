package journal

import (
	"fmt"
	"io"
	"text/tabwriter"
)

// PrintStaircaseReport writes a human-readable staircase result to w.
//
//nolint:cyclop // sequential orchestration logic
func PrintStaircaseReport( //nolint:funlen,gocognit // sequential report formatting
	w io.Writer,
	entry Entry,
) {
	fmt.Fprintln(w, "\n=== Staircase SLO Ceiling ===")

	if entry.SLOCeiling.Games == 0 {
		fmt.Fprintln(w, "No SLO ceiling found — first step already failed")
	} else {
		ceilingLine := fmt.Sprintf(
			"Ceiling: %d games | %.1f moves/s | %.1f%% completion",
			entry.SLOCeiling.Games,
			entry.SLOCeiling.ThroughputMPS,
			entry.SLOCeiling.CompletionRate*100,
		)
		if entry.SLOCeiling.EffectiveConcurrency > 0 {
			ceilingLine += fmt.Sprintf(
				" | effective: %d",
				entry.SLOCeiling.EffectiveConcurrency,
			)
		}

		fmt.Fprintln(w, ceilingLine)
	}

	fmt.Fprintln(w, "\nSteps:")

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "GAMES\tE2E p95\tWS p95\tMOVES/S\tCPU%\tMEM\tSLO")
	fmt.Fprintln(tw, "-----\t-------\t------\t-------\t----\t---\t---")

	for _, step := range entry.Steps {
		sloStatus := "PASS"
		if !step.SLOEval.AllPassing() {
			sloStatus = "FAIL"
			if len(step.SLOEval.Violations) > 0 {
				sloStatus = "FAIL <- " + step.SLOEval.Violations[0].SLO.Name
			}
		}

		mem := formatMemory(step.ServerResources.RiskIt.MemoryMB)

		fmt.Fprintf(
			tw,
			"%d\t%.3fs\t%.3fs\t%.1f\t%.0f%%\t%s\t%s\n",
			step.TargetGames,
			step.Metrics.E2E.P95,
			step.Metrics.WSDelivery.P95,
			step.Metrics.ThroughputMPS,
			step.ServerResources.RiskIt.CPUPercent,
			mem,
			sloStatus,
		)
	}

	tw.Flush()

	// Print health distribution if available.
	hasHealth := false

	for _, step := range entry.Steps {
		if step.HealthDistribution != nil {
			hasHealth = true

			break
		}
	}

	if hasHealth {
		fmt.Fprintln(w, "\nHealth Distribution:")

		hw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		fmt.Fprintln(hw, "GAMES\tHEALTHY\tSLOW\tSTALLED\tZOMBIE\tEFFECTIVE")
		fmt.Fprintln(hw, "-----\t-------\t----\t-------\t------\t---------")

		for _, step := range entry.Steps {
			if step.HealthDistribution == nil {
				continue
			}

			d := step.HealthDistribution
			fmt.Fprintf(
				hw,
				"%d\t%d\t%d\t%d\t%d\t%d\n",
				step.TargetGames,
				d.Healthy,
				d.Slow,
				d.Stalled,
				d.Zombie,
				d.EffectiveConcurrency(),
			)
		}

		hw.Flush()
	}

	if len(entry.BreakingPoints) > 0 {
		fmt.Fprintln(w, "\nBreaking points:")

		for _, bp := range entry.BreakingPoints {
			fmt.Fprintf(
				w,
				"  %s: breaks at %d games (last good: %.3fs -> %.3fs)\n",
				bp.SLOName,
				bp.BreaksAtGames,
				bp.LastGoodValue,
				bp.BreakValue,
			)
		}
	}

	// Print top DB queries if available.
	hasDBStats := false

	for _, step := range entry.Steps {
		if step.DBStats != nil && len(step.DBStats.TopQueries) > 0 {
			hasDBStats = true

			break
		}
	}

	if hasDBStats {
		fmt.Fprintln(w, "\nTop DB Queries (by total time):")

		for _, step := range entry.Steps {
			if step.DBStats == nil || len(step.DBStats.TopQueries) == 0 {
				continue
			}

			fmt.Fprintf(
				w,
				"\n  Step %d games (total: %.1fms, conns: %d):\n",
				step.TargetGames,
				step.DBStats.TotalQueryTimeMs,
				step.DBStats.ActiveConnections,
			)

			dw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
			fmt.Fprintln(dw, "    CALLS\tTOTAL ms\tMEAN ms\tMAX ms\tQUERY")

			for _, q := range step.DBStats.TopQueries {
				query := q.Query
				if len(query) > 60 {
					query = query[:57] + "..."
				}

				fmt.Fprintf(
					dw,
					"    %d\t%.1f\t%.2f\t%.1f\t%s\n",
					q.Calls,
					q.TotalTimeMs,
					q.MeanTimeMs,
					q.MaxTimeMs,
					query,
				)
			}

			dw.Flush()
		}
	}
}

// PrintCeilingComparison writes a delta report comparing two entries.
func PrintCeilingComparison(w io.Writer, before, after Entry) {
	delta := CompareCeilings(before.SLOCeiling, after.SLOCeiling)
	shift := DetectBottleneckShift(before, after)

	fmt.Fprintln(w, "\n=== Ceiling Comparison ===")
	fmt.Fprintf(
		w,
		"Games:      %d -> %d (%+d)\n",
		delta.GamesBefore,
		delta.GamesAfter,
		delta.GamesDelta,
	)

	if delta.GamesBefore > 0 {
		fmt.Fprintf(
			w,
			"Throughput: %+.1f%%\n",
			delta.ThroughputDeltaPct,
		)
		fmt.Fprintf(
			w,
			"Completion: %+.1f%%\n",
			delta.CompletionDeltaPct,
		)
	}

	if shift.Shifted {
		fmt.Fprintf(
			w,
			"\nBottleneck shift: %q -> %q\n",
			shift.Before,
			shift.After,
		)
	} else if shift.Before != "" {
		fmt.Fprintf(w, "\nBottleneck unchanged: %q\n", shift.Before)
	}
}

// formatMemory formats MB to a human-readable string.
func formatMemory(mb float64) string {
	if mb == 0 {
		return "-"
	}

	if mb >= 1024 {
		return fmt.Sprintf("%.1fGB", mb/1024)
	}

	return fmt.Sprintf("%.0fMB", mb)
}
