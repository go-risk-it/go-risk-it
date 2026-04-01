package metrics //nolint:testpackage // whitebox tests access unexported helpers

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func makeTestSnapshot() *Snapshot {
	return &Snapshot{
		RESTLatency: map[string]HistogramSnapshot{
			"deploy": {P50: 10, P95: 20, P99: 30, Max: 40, Count: 50},
		},
		WSDelivery:     HistogramSnapshot{P50: 5, P95: 10, P99: 15, Max: 20, Count: 50},
		E2EMove:        HistogramSnapshot{P50: 15, P95: 30, P99: 45, Max: 60, Count: 50},
		TotalMoves:     100,
		TotalErrors:    2,
		GamesCompleted: 1,
		GamesTimedOut:  0,
		TotalRetries:   3,
		TotalConflicts: 1,
		PhaseLatency: map[string]HistogramSnapshot{
			"attack": {P50: 15, P95: 35, P99: 52, Max: 89, Count: 40},
			"deploy": {P50: 10, P95: 22, P99: 35, Max: 56, Count: 30},
		},
		PhaseEntries: map[string]int64{
			"cards": 0, "deploy": 10, "attack": 10, "conquer": 5, "reinforce": 10,
		},
		PhaseMoves: map[string]int64{
			"cards": 0, "deploy": 30, "attack": 40, "conquer": 5, "reinforce": 25,
		},
		HTTPStatusCounts: map[int]int64{200: 10, 204: 80, 409: 1, 500: 1},
		ErrorBreakdown: map[string]int64{
			"strategy":  1,
			"execution": 1,
			"transient": 0,
			"timeout":   0,
		},
		ThroughputBuckets: []ThroughputBucket{
			{OffsetSec: 0, Moves: 12},
			{OffsetSec: 5, Moves: 45},
			{OffsetSec: 10, Moves: 43},
		},
	}
}

func TestPrintReport_ContainsNewSections(t *testing.T) {
	t.Parallel()
	snap := makeTestSnapshot()

	var buf bytes.Buffer
	PrintReport(&buf, snap, 30*time.Second, 0, nil)
	out := buf.String()

	sections := []string{
		"Phase Latency (ms):",
		"Phase Flow:",
		"avg-moves/entry",
		"[execution=1 strategy=1]",
		"HTTP Status:",
		"204=80",
		"409=1",
		"Throughput:",
		"/s avg",
		"/s peak",
	}

	for _, s := range sections {
		if !strings.Contains(out, s) {
			t.Errorf("report missing section %q", s)
		}
	}
}

func TestPrintReport_PhaseFlowSkipsZeroPhases(t *testing.T) {
	t.Parallel()
	snap := makeTestSnapshot()

	var buf bytes.Buffer
	PrintReport(&buf, snap, 30*time.Second, 0, nil)
	out := buf.String()

	// "cards" has 0 entries and 0 moves — should not appear in Phase Flow table.
	lines := strings.Split(out, "\n")
	inPhaseFlow := false

	for _, line := range lines {
		if strings.Contains(line, "Phase Flow:") {
			inPhaseFlow = true

			continue
		}

		if inPhaseFlow && strings.TrimSpace(line) == "" {
			break
		}

		if inPhaseFlow && strings.Contains(line, "cards") {
			t.Error("Phase Flow should not include zero-count 'cards' phase")
		}
	}
}

func TestPrintJSON_ContainsNewFields(t *testing.T) {
	t.Parallel()
	snap := makeTestSnapshot()

	var buf bytes.Buffer
	if err := PrintJSON(&buf, snap, 30*time.Second, 0, nil); err != nil {
		t.Fatalf("PrintJSON: %v", err)
	}

	var report JSONReport
	if err := json.Unmarshal(buf.Bytes(), &report); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// Phase latency.
	if len(report.PhaseLatency) != 2 {
		t.Errorf("expected 2 phase latency entries, got %d", len(report.PhaseLatency))
	}

	if attack, ok := report.PhaseLatency["attack"]; !ok {
		t.Error("missing attack phase latency")
	} else if attack.Count != 40 {
		t.Errorf("expected 40 attack count, got %d", attack.Count)
	}

	// Phase flow.
	if len(report.PhaseFlow) != 5 {
		t.Errorf("expected 5 phase flow entries, got %d", len(report.PhaseFlow))
	}

	if af, ok := report.PhaseFlow["attack"]; !ok {
		t.Error("missing attack phase flow")
	} else {
		if af.Entries != 10 {
			t.Errorf("expected 10 attack entries, got %d", af.Entries)
		}

		if af.Moves != 40 {
			t.Errorf("expected 40 attack moves, got %d", af.Moves)
		}

		if af.AvgMovesPerEntry != 4.0 {
			t.Errorf("expected 4.0 avg moves/entry, got %.1f", af.AvgMovesPerEntry)
		}
	}

	// HTTP status counts.
	if len(report.HTTPStatusCounts) != 4 {
		t.Errorf("expected 4 HTTP status entries, got %d", len(report.HTTPStatusCounts))
	}

	if report.HTTPStatusCounts["204"] != 80 {
		t.Errorf("expected 80 for HTTP 204, got %d", report.HTTPStatusCounts["204"])
	}

	// Throughput series.
	if len(report.ThroughputSeries) != 3 {
		t.Errorf("expected 3 throughput buckets, got %d", len(report.ThroughputSeries))
	}

	// Error breakdown.
	if report.Errors.Breakdown["strategy"] != 1 {
		t.Errorf("expected 1 strategy error, got %d", report.Errors.Breakdown["strategy"])
	}
}

func TestPrintReport_EmptySnapshot(t *testing.T) {
	t.Parallel()
	snap := &Snapshot{
		PhaseEntries:   map[string]int64{},
		PhaseMoves:     map[string]int64{},
		ErrorBreakdown: map[string]int64{},
	}

	var buf bytes.Buffer

	// Should not panic with empty data.
	PrintReport(&buf, snap, 0, 0, nil)

	out := buf.String()
	if !strings.Contains(out, "Performance Test Report") {
		t.Error("report header missing")
	}
}
