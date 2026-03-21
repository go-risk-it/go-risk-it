package metrics

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"time"
)

// PrintReport writes a human-readable performance report to w.
func PrintReport(w io.Writer, snap *Snapshot, totalDuration time.Duration, fatalErrors int) {
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

	// Errors.
	errorRate := float64(0)
	if snap.TotalMoves > 0 {
		errorRate = float64(snap.TotalErrors) / float64(snap.TotalMoves) * 100
	}

	fmt.Fprintf(w, "\nErrors: %d (%.2f%%)\n", snap.TotalErrors, errorRate)
	fmt.Fprintf(w, "Duration: %s\n", totalDuration.Round(time.Millisecond))
}

// JSONReport is the machine-readable report structure.
type JSONReport struct {
	Games struct {
		Completed int64 `json:"completed"`
		TimedOut  int64 `json:"timed_out"`
		Fatal     int   `json:"fatal"`
	} `json:"games"`
	Moves struct {
		Total  int64   `json:"total"`
		PerSec float64 `json:"per_sec"`
	} `json:"moves"`
	RESTLatency map[string]JSONHistogram `json:"rest_latency"`
	WSDelivery  JSONHistogram            `json:"ws_delivery"`
	E2EMove     JSONHistogram            `json:"e2e_move"`
	Errors      struct {
		Total int64   `json:"total"`
		Rate  float64 `json:"rate_pct"`
	} `json:"errors"`
	DurationMs int64 `json:"duration_ms"`
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
func PrintJSON(w io.Writer, snap *Snapshot, totalDuration time.Duration, fatalErrors int) error {
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
		report.RESTLatency[name] = JSONHistogram{
			P50: h.P50, P95: h.P95, P99: h.P99, Max: h.Max, Count: h.Count,
		}
	}

	report.WSDelivery = toJSONHist(snap.WSDelivery)
	report.E2EMove = toJSONHist(snap.E2EMove)
	report.Errors.Total = snap.TotalErrors

	if snap.TotalMoves > 0 {
		report.Errors.Rate = float64(snap.TotalErrors) / float64(snap.TotalMoves) * 100
	}

	report.DurationMs = totalDuration.Milliseconds()

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")

	return enc.Encode(report)
}

func toJSONHist(h HistogramSnapshot) JSONHistogram {
	return JSONHistogram{P50: h.P50, P95: h.P95, P99: h.P99, Max: h.Max, Count: h.Count}
}

func sortedKeys(m map[string]HistogramSnapshot) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}
