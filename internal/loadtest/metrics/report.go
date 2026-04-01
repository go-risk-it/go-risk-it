package metrics

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"time"
)

// GameResult holds per-game stats for reporting.
type GameResult struct {
	GameIndex  int
	Duration   time.Duration
	Moves      int
	Errors     int
	Winner     string
	TimedOut   bool
	FatalError error
}

// PrintReport writes a human-readable performance report to w.
//
//nolint:gocognit // sequential report formatting
func PrintReport( //nolint:cyclop,funlen // sequential report formatting
	w io.Writer,
	snap *Snapshot,
	totalDuration time.Duration,
	fatalErrors int,
	results []GameResult,
) {
	fmt.Fprintln(w)
	fmt.Fprintln(w, "=== Performance Test Report ===")

	// Games summary.
	fmt.Fprintf(w, "Games: %d completed, %d timed out, %d fatal\n",
		snap.GamesCompleted, snap.GamesTimedOut, fatalErrors)

	// Moves summary.
	totalSec := totalDuration.Seconds()
	movesPerSec := float64(0)

	if totalSec > 0 {
		movesPerSec = float64(snap.TotalMoves) / totalSec
	}

	fmt.Fprintf(w, "Moves: %d total (%.1f/s avg)\n", snap.TotalMoves, movesPerSec)

	// REST latency table.
	if len(snap.RESTLatency) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "REST Latency (ms):")
		fmt.Fprintf(
			w,
			"  %-12s %6s %6s %6s %6s %8s\n",
			"Operation",
			"p50",
			"p95",
			"p99",
			"max",
			"count",
		)

		names := sortedKeys(snap.RESTLatency)
		for _, name := range names {
			h := snap.RESTLatency[name]
			fmt.Fprintf(w, "  %-12s %6d %6d %6d %6d %8d\n",
				name, h.P50, h.P95, h.P99, h.Max, h.Count)
		}
	}

	// Phase latency table.
	if len(snap.PhaseLatency) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Phase Latency (ms):")
		fmt.Fprintf(w, "  %-12s %6s %6s %6s %6s %8s\n",
			"Phase", "p50", "p95", "p99", "max", "count")

		names := sortedKeys(snap.PhaseLatency)
		for _, name := range names {
			h := snap.PhaseLatency[name]
			fmt.Fprintf(w, "  %-12s %6d %6d %6d %6d %8d\n",
				name, h.P50, h.P95, h.P99, h.Max, h.Count)
		}
	}

	// Phase flow table.
	if hasNonZeroPhaseFlow(snap) {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Phase Flow:")
		fmt.Fprintf(w, "  %-12s %8s %7s %17s\n",
			"Phase", "entries", "moves", "avg-moves/entry")

		phases := sortedStringKeys(snap.PhaseEntries)
		for _, phase := range phases {
			entries := snap.PhaseEntries[phase]
			moves := snap.PhaseMoves[phase]

			if entries == 0 && moves == 0 {
				continue
			}

			avg := float64(0)
			if entries > 0 {
				avg = float64(moves) / float64(entries)
			}

			fmt.Fprintf(w, "  %-12s %8d %7d %17.1f\n",
				phase, entries, moves, avg)
		}
	}

	// WS delivery.
	if snap.WSDelivery.Count > 0 {
		fmt.Fprintln(w)
		fmt.Fprintf(w, "WS Delivery (ms):  p50=%d  p95=%d  p99=%d  max=%d\n",
			snap.WSDelivery.P50, snap.WSDelivery.P95, snap.WSDelivery.P99, snap.WSDelivery.Max)
	}

	// E2E move.
	if snap.E2EMove.Count > 0 {
		fmt.Fprintf(w, "E2E Move (ms):     p50=%d  p95=%d  p99=%d  max=%d\n",
			snap.E2EMove.P50, snap.E2EMove.P95, snap.E2EMove.P99, snap.E2EMove.Max)
	}

	// Errors with breakdown.
	errorRate := float64(0)
	if snap.TotalMoves > 0 {
		errorRate = float64(snap.TotalErrors) / float64(snap.TotalMoves) * 100
	}

	fmt.Fprintf(w, "\nErrors: %d (%.2f%%)", snap.TotalErrors, errorRate)

	if len(snap.ErrorBreakdown) > 0 {
		fmt.Fprint(w, "  [")

		first := true
		for _, name := range sortedStringKeys(snap.ErrorBreakdown) {
			if count := snap.ErrorBreakdown[name]; count > 0 {
				if !first {
					fmt.Fprint(w, " ")
				}

				fmt.Fprintf(w, "%s=%d", name, count)
				first = false
			}
		}

		fmt.Fprint(w, "]")
	}

	fmt.Fprintln(w)

	// Resilience.
	fmt.Fprintf(
		w,
		"Resilience: %d retries, %d conflicts (409), %d reconnects, %d reconnect failures\n",
		snap.TotalRetries,
		snap.TotalConflicts,
		snap.TotalReconnects,
		snap.TotalReconnectFailures,
	)

	// Chaos events (only if any occurred).
	if len(snap.ChaosEvents) > 0 {
		fmt.Fprint(w, "Chaos:")

		for _, name := range sortedStringKeys(snap.ChaosEvents) {
			fmt.Fprintf(w, " %d %s,", snap.ChaosEvents[name], name)
		}

		fmt.Fprintln(w)
	}

	// HTTP status distribution.
	if len(snap.HTTPStatusCounts) > 0 {
		fmt.Fprint(w, "HTTP Status:")
		codes := sortedIntKeys(snap.HTTPStatusCounts)

		for _, code := range codes {
			count := snap.HTTPStatusCounts[code]
			if count > 0 {
				fmt.Fprintf(w, " %d=%d", code, count)
			}
		}

		fmt.Fprintln(w)
	}

	// Throughput summary.
	if len(snap.ThroughputBuckets) > 0 {
		var totalMoves int64

		var peakMoves int64

		for _, b := range snap.ThroughputBuckets {
			totalMoves += b.Moves
			if b.Moves > peakMoves {
				peakMoves = b.Moves
			}
		}

		// Average = total moves / number of seconds covered by non-zero buckets.
		numBuckets := len(snap.ThroughputBuckets)
		avgPerSec := float64(totalMoves) / (float64(numBuckets) * 5.0)
		peakPerSec := float64(peakMoves) / 5.0
		fmt.Fprintf(w, "Throughput: %.1f/s avg, %.1f/s peak (5s buckets)\n", avgPerSec, peakPerSec)
	}

	fmt.Fprintf(w, "Duration: %s\n", totalDuration.Round(time.Millisecond))

	// Per-game results table (only when multiple games).
	if len(results) > 1 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Per-Game Results:")
		fmt.Fprintf(w, "  %5s %10s %7s %7s  %s\n", "Game", "Duration", "Moves", "Errors", "Status")

		for _, r := range results {
			status := gameStatus(r)
			fmt.Fprintf(w, "  %5d %10s %7d %7d  %s\n",
				r.GameIndex,
				r.Duration.Round(100*time.Millisecond),
				r.Moves,
				r.Errors,
				status,
			)
		}
	}
}

// JSONReport is the machine-readable report structure.
type JSONReport struct {
	Games struct {
		Completed int64 `json:"completed"`
		TimedOut  int64 `json:"timedOut"`
		Fatal     int   `json:"fatal"`
	} `json:"games"`
	Moves struct {
		Total  int64   `json:"total"`
		PerSec float64 `json:"perSec"`
	} `json:"moves"`
	RESTLatency      map[string]JSONHistogram `json:"restLatency"`
	WSDelivery       JSONHistogram            `json:"wsDelivery"`
	E2EMove          JSONHistogram            `json:"e2eMove"`
	PhaseLatency     map[string]JSONHistogram `json:"phaseLatency"`
	PhaseFlow        map[string]JSONPhaseFlow `json:"phaseFlow"`
	HTTPStatusCounts map[string]int64         `json:"httpStatusCounts"`
	ThroughputSeries []JSONThroughputBucket   `json:"throughputSeries"`
	Errors           struct {
		Total     int64            `json:"total"`
		Rate      float64          `json:"ratePct"`
		Breakdown map[string]int64 `json:"breakdown"`
	} `json:"errors"`
	Resilience struct {
		Retries           int64 `json:"retries"`
		Conflicts         int64 `json:"conflicts"`
		Reconnects        int64 `json:"reconnects"`
		ReconnectFailures int64 `json:"reconnectFailures"`
	} `json:"resilience"`
	DurationMs int64            `json:"durationMs"`
	PerGame    []JSONGameResult `json:"perGame,omitempty"`

	// Chaos events (only present when chaos was active).
	ChaosEvents map[string]int64 `json:"chaosEvents,omitempty"`
}

// JSONPhaseFlow is the JSON representation of phase transition stats.
type JSONPhaseFlow struct {
	Entries          int64   `json:"entries"`
	Moves            int64   `json:"moves"`
	AvgMovesPerEntry float64 `json:"avgMovesPerEntry"`
}

// JSONThroughputBucket is the JSON representation of a throughput time bucket.
type JSONThroughputBucket struct {
	OffsetSec float64 `json:"offsetSec"`
	Moves     int64   `json:"moves"`
}

// JSONGameResult is the JSON representation of a single game's result.
type JSONGameResult struct {
	GameIndex  int    `json:"gameIndex"`
	DurationMs int64  `json:"durationMs"`
	Moves      int    `json:"moves"`
	Errors     int    `json:"errors"`
	Status     string `json:"status"`
	Winner     string `json:"winner,omitempty"`
}

// JSONHistogram is the JSON representation of a histogram snapshot.
type JSONHistogram struct {
	P50   int64 `json:"p50"`
	P95   int64 `json:"p95"`
	P99   int64 `json:"p99"`
	Max   int64 `json:"max"`
	Count int64 `json:"count"`
}

// PrintJSON writes a machine-readable JSON report to w.
//
//nolint:cyclop,funlen // sequential JSON report building
func PrintJSON(
	w io.Writer,
	snap *Snapshot,
	totalDuration time.Duration,
	fatalErrors int,
	results []GameResult,
) error {
	report := JSONReport{}
	report.Games.Completed = snap.GamesCompleted
	report.Games.TimedOut = snap.GamesTimedOut
	report.Games.Fatal = fatalErrors
	report.Moves.Total = snap.TotalMoves

	totalSec := totalDuration.Seconds()
	if totalSec > 0 {
		report.Moves.PerSec = float64(snap.TotalMoves) / totalSec
	}

	report.RESTLatency = make(map[string]JSONHistogram, len(snap.RESTLatency))
	for name, h := range snap.RESTLatency {
		report.RESTLatency[name] = JSONHistogram(h)
	}

	report.WSDelivery = toJSONHist(snap.WSDelivery)
	report.E2EMove = toJSONHist(snap.E2EMove)
	report.Errors.Total = snap.TotalErrors

	if snap.TotalMoves > 0 {
		report.Errors.Rate = float64(snap.TotalErrors) / float64(snap.TotalMoves) * 100
	}

	report.Errors.Breakdown = snap.ErrorBreakdown

	// Phase latency.
	report.PhaseLatency = make(map[string]JSONHistogram, len(snap.PhaseLatency))
	for name, h := range snap.PhaseLatency {
		report.PhaseLatency[name] = JSONHistogram(h)
	}

	// Phase flow.
	report.PhaseFlow = make(map[string]JSONPhaseFlow, len(snap.PhaseEntries))
	for phase, entries := range snap.PhaseEntries {
		moves := snap.PhaseMoves[phase]
		avg := float64(0)

		if entries > 0 {
			avg = float64(moves) / float64(entries)
		}

		report.PhaseFlow[phase] = JSONPhaseFlow{
			Entries:          entries,
			Moves:            moves,
			AvgMovesPerEntry: avg,
		}
	}

	// HTTP status counts (use string keys for JSON).
	report.HTTPStatusCounts = make(map[string]int64, len(snap.HTTPStatusCounts))
	for code, count := range snap.HTTPStatusCounts {
		report.HTTPStatusCounts[strconv.Itoa(code)] = count
	}

	// Throughput series.
	report.ThroughputSeries = make([]JSONThroughputBucket, len(snap.ThroughputBuckets))
	for i, b := range snap.ThroughputBuckets {
		report.ThroughputSeries[i] = JSONThroughputBucket(b)
	}

	report.DurationMs = totalDuration.Milliseconds()

	report.Resilience.Retries = snap.TotalRetries
	report.Resilience.Conflicts = snap.TotalConflicts
	report.Resilience.Reconnects = snap.TotalReconnects
	report.Resilience.ReconnectFailures = snap.TotalReconnectFailures

	if len(snap.ChaosEvents) > 0 {
		report.ChaosEvents = snap.ChaosEvents
	}

	if len(results) > 0 {
		report.PerGame = make([]JSONGameResult, len(results))
		for i, r := range results {
			report.PerGame[i] = JSONGameResult{
				GameIndex:  r.GameIndex,
				DurationMs: r.Duration.Milliseconds(),
				Moves:      r.Moves,
				Errors:     r.Errors,
				Status:     gameStatus(r),
				Winner:     r.Winner,
			}
		}
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")

	if err := enc.Encode(report); err != nil {
		return fmt.Errorf("encode json report: %w", err)
	}

	return nil
}

func toJSONHist(h HistogramSnapshot) JSONHistogram {
	return JSONHistogram(h)
}

func sortedKeys(m map[string]HistogramSnapshot) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}

func sortedStringKeys(m map[string]int64) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}

func sortedIntKeys(m map[int]int64) []int {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	sort.Ints(keys)

	return keys
}

func hasNonZeroPhaseFlow(snap *Snapshot) bool {
	for _, v := range snap.PhaseEntries {
		if v > 0 {
			return true
		}
	}

	for _, v := range snap.PhaseMoves {
		if v > 0 {
			return true
		}
	}

	return false
}

func gameStatus(r GameResult) string {
	switch {
	case r.FatalError != nil:
		return "fatal: " + r.FatalError.Error()
	case r.TimedOut:
		return "timed out"
	case r.Winner != "":
		winner := r.Winner
		if len(winner) > 8 {
			winner = winner[:8]
		}

		return "completed (winner: " + winner + ")"
	default:
		return "completed"
	}
}
