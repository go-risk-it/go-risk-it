package metrics //nolint:testpackage // whitebox tests access unexported helpers

import (
	"testing"
	"time"
)

func TestNewStepAccumulator(t *testing.T) {
	t.Parallel()
	c := NewStepAccumulator(1 * time.Minute)

	snap := c.Snapshot()
	if snap.TotalMoves != 0 {
		t.Errorf("expected 0 total moves, got %d", snap.TotalMoves)
	}

	if len(snap.ThroughputBuckets) != 0 {
		t.Errorf("expected 0 throughput buckets, got %d", len(snap.ThroughputBuckets))
	}

	// Pre-initialized phase maps should exist with zero values.
	for _, phase := range knownPhases {
		if _, ok := snap.PhaseEntries[string(phase)]; !ok {
			t.Errorf("missing phase entry for %q", phase)
		}

		if _, ok := snap.PhaseMoves[string(phase)]; !ok {
			t.Errorf("missing phase move for %q", phase)
		}
	}

	// Pre-initialized error categories should exist.
	for _, cat := range []ErrorType{
		ErrorTypeStrategy, ErrorTypeExecution, ErrorTypeTransient, ErrorTypeTimeout,
		ErrorTypeStaleState, ErrorTypeConflict,
	} {
		if _, ok := snap.ErrorBreakdown[string(cat)]; !ok {
			t.Errorf("missing error category %q", cat)
		}
	}
}

func TestRecordPhaseLatency(t *testing.T) {
	t.Parallel()
	c := NewStepAccumulator(1 * time.Minute)

	c.RecordPhaseLatency("attack", 15*time.Millisecond)
	c.RecordPhaseLatency("attack", 25*time.Millisecond)
	c.RecordPhaseLatency("deploy", 10*time.Millisecond)

	snap := c.Snapshot()

	attack, ok := snap.PhaseLatency["attack"]
	if !ok {
		t.Fatal("missing phase latency for attack")
	}

	if attack.Count != 2 {
		t.Errorf("expected 2 attack samples, got %d", attack.Count)
	}

	deploy, ok := snap.PhaseLatency["deploy"]
	if !ok {
		t.Fatal("missing phase latency for deploy")
	}

	if deploy.Count != 1 {
		t.Errorf("expected 1 deploy sample, got %d", deploy.Count)
	}
}

func TestRecordPhaseEntryAndMove(t *testing.T) {
	t.Parallel()
	c := NewStepAccumulator(1 * time.Minute)

	c.RecordPhaseEntry("attack")
	c.RecordPhaseEntry("attack")
	c.RecordPhaseEntry("deploy")

	c.RecordPhaseMove("attack")
	c.RecordPhaseMove("attack")
	c.RecordPhaseMove("attack")
	c.RecordPhaseMove("deploy")

	snap := c.Snapshot()

	if snap.PhaseEntries["attack"] != 2 {
		t.Errorf("expected 2 attack entries, got %d", snap.PhaseEntries["attack"])
	}

	if snap.PhaseEntries["deploy"] != 1 {
		t.Errorf("expected 1 deploy entry, got %d", snap.PhaseEntries["deploy"])
	}

	if snap.PhaseMoves["attack"] != 3 {
		t.Errorf("expected 3 attack moves, got %d", snap.PhaseMoves["attack"])
	}

	if snap.PhaseMoves["deploy"] != 1 {
		t.Errorf("expected 1 deploy move, got %d", snap.PhaseMoves["deploy"])
	}
}

func TestRecordPhaseEntry_UnknownPhaseIgnored(t *testing.T) {
	t.Parallel()
	c := NewStepAccumulator(1 * time.Minute)

	// Should not panic or create an entry.
	c.RecordPhaseEntry("nonexistent")
	c.RecordPhaseMove("nonexistent")

	snap := c.Snapshot()
	if _, ok := snap.PhaseEntries["nonexistent"]; ok {
		t.Error("unexpected entry for unknown phase")
	}
}

func TestRecordHTTPStatus(t *testing.T) {
	t.Parallel()
	c := NewStepAccumulator(1 * time.Minute)

	c.RecordHTTPStatus(200)
	c.RecordHTTPStatus(204)
	c.RecordHTTPStatus(204)
	c.RecordHTTPStatus(409)
	c.RecordHTTPStatus(500)

	snap := c.Snapshot()

	want := map[int]int64{200: 1, 204: 2, 409: 1, 500: 1}
	for code, expected := range want {
		got := snap.HTTPStatusCounts[code]
		if got != expected {
			t.Errorf("HTTP %d: expected %d, got %d", code, expected, got)
		}
	}
}

func TestRecordErrorType(t *testing.T) {
	t.Parallel()
	c := NewStepAccumulator(1 * time.Minute)

	c.RecordErrorType(ErrorTypeStrategy)
	c.RecordErrorType(ErrorTypeStrategy)
	c.RecordErrorType(ErrorTypeExecution)
	c.RecordErrorType(ErrorTypeTransient)

	snap := c.Snapshot()

	if snap.ErrorBreakdown[string(ErrorTypeStrategy)] != 2 {
		t.Errorf(
			"expected 2 strategy errors, got %d",
			snap.ErrorBreakdown[string(ErrorTypeStrategy)],
		)
	}

	if snap.ErrorBreakdown[string(ErrorTypeExecution)] != 1 {
		t.Errorf(
			"expected 1 execution error, got %d",
			snap.ErrorBreakdown[string(ErrorTypeExecution)],
		)
	}

	if snap.ErrorBreakdown[string(ErrorTypeTransient)] != 1 {
		t.Errorf(
			"expected 1 transient error, got %d",
			snap.ErrorBreakdown[string(ErrorTypeTransient)],
		)
	}

	if snap.ErrorBreakdown[string(ErrorTypeTimeout)] != 0 {
		t.Errorf(
			"expected 0 timeout errors, got %d",
			snap.ErrorBreakdown[string(ErrorTypeTimeout)],
		)
	}
}

func TestRecordErrorType_UnknownIgnored(t *testing.T) {
	t.Parallel()
	c := NewStepAccumulator(1 * time.Minute)

	// Should not panic or create an entry.
	c.RecordErrorType(ErrorType("unknown_category"))

	snap := c.Snapshot()
	if _, ok := snap.ErrorBreakdown["unknown_category"]; ok {
		t.Error("unexpected entry for unknown error type")
	}
}

func TestRecordTimedMove_CorrectBucket(t *testing.T) {
	t.Parallel()
	c := NewStepAccumulator(30 * time.Second)

	// Bucket 0 = 0-5s from start. Recording immediately should land in bucket 0.
	c.RecordTimedMove()
	c.RecordTimedMove()

	snap := c.Snapshot()

	if len(snap.ThroughputBuckets) != 1 {
		t.Fatalf("expected 1 non-zero bucket, got %d", len(snap.ThroughputBuckets))
	}

	if snap.ThroughputBuckets[0].OffsetSec != 0 {
		t.Errorf("expected bucket offset 0, got %.1f", snap.ThroughputBuckets[0].OffsetSec)
	}

	if snap.ThroughputBuckets[0].Moves != 2 {
		t.Errorf("expected 2 moves in bucket, got %d", snap.ThroughputBuckets[0].Moves)
	}
}

func TestRecordTimedMove_OutOfRangeIgnored(t *testing.T) {
	t.Parallel()
	// Only 1 bucket (5s window).
	c := NewStepAccumulator(1 * time.Millisecond)

	// Manually push startTime back so elapsed > maxDuration.
	c.startTime = time.Now().Add(-10 * time.Second)
	c.RecordTimedMove()

	snap := c.Snapshot()
	if len(snap.ThroughputBuckets) != 0 {
		t.Errorf("expected 0 buckets for out-of-range move, got %d", len(snap.ThroughputBuckets))
	}
}

func TestSnapshot_ZeroBucketsOmitted(t *testing.T) {
	t.Parallel()
	c := NewStepAccumulator(1 * time.Minute)

	// No moves recorded — all buckets should be zero and omitted.
	snap := c.Snapshot()
	if len(snap.ThroughputBuckets) != 0 {
		t.Errorf("expected 0 throughput buckets, got %d", len(snap.ThroughputBuckets))
	}
}

func TestClampMs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		d    time.Duration
		want int64
	}{
		{"sub-millisecond clamps to 1", 500 * time.Microsecond, 1},
		{"zero clamps to 1", 0, 1},
		{"normal value passes through", 50 * time.Millisecond, 50},
		{"exceeds max clamps to max", 60 * time.Second, maxLatencyMs},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := clampMs(tt.d)
			if got != tt.want {
				t.Errorf("clampMs(%v) = %d, want %d", tt.d, got, tt.want)
			}
		})
	}
}

func TestBucketSizing(t *testing.T) {
	t.Parallel()
	// 30s max duration with 5s buckets = 7 buckets (30/5 + 1).
	c := NewStepAccumulator(30 * time.Second)

	if len(c.moveBuckets) != 7 {
		t.Errorf("expected 7 buckets for 30s, got %d", len(c.moveBuckets))
	}

	// 10m max duration with 5s buckets = 121 buckets (600/5 + 1).
	c2 := NewStepAccumulator(10 * time.Minute)

	if len(c2.moveBuckets) != 121 {
		t.Errorf("expected 121 buckets for 10m, got %d", len(c2.moveBuckets))
	}
}

func TestRecordErrorType_StaleState(t *testing.T) {
	t.Parallel()
	c := NewStepAccumulator(1 * time.Minute)

	c.RecordErrorType(ErrorTypeStaleState)
	c.RecordErrorType(ErrorTypeStaleState)
	c.RecordErrorType(ErrorTypeStaleState)

	snap := c.Snapshot()

	if snap.ErrorBreakdown[string(ErrorTypeStaleState)] != 3 {
		t.Errorf(
			"expected 3 stale_state errors, got %d",
			snap.ErrorBreakdown[string(ErrorTypeStaleState)],
		)
	}
}

func TestRecordConflict_AppearsInBreakdown(t *testing.T) {
	t.Parallel()
	c := NewStepAccumulator(1 * time.Minute)

	c.RecordConflict()
	c.RecordConflict()

	snap := c.Snapshot()

	// Should appear in both TotalConflicts and ErrorBreakdown.
	if snap.TotalConflicts != 2 {
		t.Errorf("expected 2 total conflicts, got %d", snap.TotalConflicts)
	}

	if snap.ErrorBreakdown[string(ErrorTypeConflict)] != 2 {
		t.Errorf(
			"expected 2 conflict errors in breakdown, got %d",
			snap.ErrorBreakdown[string(ErrorTypeConflict)],
		)
	}
}

func TestRecordGameComplete_CounterOnly(t *testing.T) {
	t.Parallel()
	c := NewStepAccumulator(1 * time.Minute)

	c.RecordGameComplete(5*time.Second, 42)
	c.RecordGameComplete(3*time.Second, 30)

	snap := c.Snapshot()
	if snap.GamesCompleted != 2 {
		t.Errorf("expected 2 games completed, got %d", snap.GamesCompleted)
	}
}

func TestRecordGameTimedOut_CounterOnly(t *testing.T) {
	t.Parallel()
	c := NewStepAccumulator(1 * time.Minute)

	c.RecordGameTimedOut(10*time.Minute, 100)

	snap := c.Snapshot()
	if snap.GamesTimedOut != 1 {
		t.Errorf("expected 1 game timed out, got %d", snap.GamesTimedOut)
	}
}

func TestRecordGameFatal_CounterOnly(t *testing.T) {
	t.Parallel()
	c := NewStepAccumulator(1 * time.Minute)

	c.RecordGameFatal()
	c.RecordGameFatal()
	c.RecordGameFatal()

	snap := c.Snapshot()
	if snap.GamesFatal != 3 {
		t.Errorf("expected 3 fatal games, got %d", snap.GamesFatal)
	}
}

func TestRecordGameStarted_NoOp(t *testing.T) {
	t.Parallel()
	c := NewStepAccumulator(1 * time.Minute)

	// RecordGameStarted is now a no-op (OTel coupling removed).
	// Calling it should not panic or affect any snapshot counters.
	c.RecordGameStarted()
	c.RecordGameStarted()

	snap := c.Snapshot()
	// No "games started" counter exists in the snapshot — verify nothing else changed.
	if snap.GamesCompleted != 0 {
		t.Errorf("expected 0 games completed, got %d", snap.GamesCompleted)
	}
}
