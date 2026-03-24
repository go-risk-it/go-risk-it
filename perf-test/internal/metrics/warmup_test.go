package metrics

import (
	"testing"
	"time"
)

func TestCollector_WarmUp_HistogramsGated(t *testing.T) {
	c := NewCollector(1 * time.Minute)
	c.ConfigureWarmUp(WarmUpConfig{
		MinCompletions: 2,
		MinDuration:    0, // only completions trigger
	})

	// Record before warm-up completes — histogram should not count.
	c.RecordE2E(50 * time.Millisecond)
	c.RecordREST("deploy", 30*time.Millisecond)
	c.RecordWSDelivery(10 * time.Millisecond)
	c.RecordPhaseLatency("attack", 20*time.Millisecond)

	snap := c.Snapshot()
	if snap.E2EMove.Count != 0 {
		t.Errorf("E2E histogram should be gated during warm-up, got count=%d", snap.E2EMove.Count)
	}

	if snap.RESTLatency["deploy"].Count != 0 {
		t.Errorf(
			"REST histogram should be gated during warm-up, got count=%d",
			snap.RESTLatency["deploy"].Count,
		)
	}

	if snap.WSDelivery.Count != 0 {
		t.Errorf(
			"WS histogram should be gated during warm-up, got count=%d",
			snap.WSDelivery.Count,
		)
	}

	if phase, ok := snap.PhaseLatency["attack"]; ok && phase.Count != 0 {
		t.Errorf("Phase histogram should be gated during warm-up, got count=%d", phase.Count)
	}

	if snap.WarmUpComplete {
		t.Error("warm-up should not be complete yet")
	}

	// Complete 2 games — warm-up satisfied.
	c.RecordGameComplete(1*time.Second, 10)
	c.RecordGameComplete(1*time.Second, 10)

	// Record after warm-up completes — histogram should count.
	c.RecordE2E(50 * time.Millisecond)

	snap = c.Snapshot()
	if snap.E2EMove.Count != 1 {
		t.Errorf("E2E histogram should record after warm-up, got count=%d", snap.E2EMove.Count)
	}

	if !snap.WarmUpComplete {
		t.Error("warm-up should be complete after 2 game completions")
	}
}

func TestCollector_WarmUp_CountersAlwaysRecorded(t *testing.T) {
	c := NewCollector(1 * time.Minute)
	c.ConfigureWarmUp(WarmUpConfig{
		MinCompletions: 100, // high bar — never reached
		MinDuration:    1 * time.Hour,
	})

	// Counters should increment even during warm-up.
	c.RecordMove()
	c.RecordMove()
	c.RecordError()
	c.RecordPhaseEntry("attack")
	c.RecordPhaseMove("attack")
	c.RecordHTTPStatus(200)
	c.RecordErrorType(ErrorTypeTransient)
	c.RecordTimedMove()
	c.RecordRetry()
	c.RecordConflict()
	c.RecordReconnect()
	c.RecordReconnectFailure()
	c.RecordChaosEvent(ChaosEventDisconnect)

	snap := c.Snapshot()
	if snap.TotalMoves != 2 {
		t.Errorf("TotalMoves: expected 2, got %d", snap.TotalMoves)
	}

	if snap.TotalErrors != 1 {
		t.Errorf("TotalErrors: expected 1, got %d", snap.TotalErrors)
	}

	if snap.PhaseEntries["attack"] != 1 {
		t.Errorf("PhaseEntries[attack]: expected 1, got %d", snap.PhaseEntries["attack"])
	}

	if snap.PhaseMoves["attack"] != 1 {
		t.Errorf("PhaseMoves[attack]: expected 1, got %d", snap.PhaseMoves["attack"])
	}

	if snap.HTTPStatusCounts[200] != 1 {
		t.Errorf("HTTPStatusCounts[200]: expected 1, got %d", snap.HTTPStatusCounts[200])
	}

	if snap.ErrorBreakdown[string(ErrorTypeTransient)] != 1 {
		t.Errorf(
			"ErrorBreakdown[transient]: expected 1, got %d",
			snap.ErrorBreakdown[string(ErrorTypeTransient)],
		)
	}

	if snap.TotalRetries != 1 {
		t.Errorf("TotalRetries: expected 1, got %d", snap.TotalRetries)
	}

	if snap.TotalConflicts != 1 {
		t.Errorf("TotalConflicts: expected 1, got %d", snap.TotalConflicts)
	}

	if snap.TotalReconnects != 1 {
		t.Errorf("TotalReconnects: expected 1, got %d", snap.TotalReconnects)
	}

	if snap.TotalReconnectFailures != 1 {
		t.Errorf("TotalReconnectFailures: expected 1, got %d", snap.TotalReconnectFailures)
	}

	if snap.ChaosEvents[string(ChaosEventDisconnect)] != 1 {
		t.Errorf(
			"ChaosEvents[disconnect]: expected 1, got %d",
			snap.ChaosEvents[string(ChaosEventDisconnect)],
		)
	}
}

func TestCollector_WarmUp_DurationTrigger(t *testing.T) {
	c := NewCollector(1 * time.Minute)
	c.ConfigureWarmUp(WarmUpConfig{
		MinCompletions: 0, // no completion trigger
		MinDuration:    50 * time.Millisecond,
	})

	// Complete a game immediately — duration not met yet.
	c.RecordGameComplete(1*time.Second, 5)

	// Record before duration elapses — should be gated.
	c.RecordE2E(50 * time.Millisecond)

	snap := c.Snapshot()
	if snap.E2EMove.Count != 0 {
		t.Errorf("E2E should be gated before duration trigger, got count=%d", snap.E2EMove.Count)
	}

	// Wait for duration trigger.
	time.Sleep(60 * time.Millisecond)

	// Another game complete triggers the check.
	c.RecordGameComplete(1*time.Second, 5)

	// Now recording should work.
	c.RecordE2E(50 * time.Millisecond)

	snap = c.Snapshot()
	if snap.E2EMove.Count != 1 {
		t.Errorf("E2E should record after duration trigger, got count=%d", snap.E2EMove.Count)
	}
}

func TestCollector_WarmUp_BothRequired(t *testing.T) {
	c := NewCollector(1 * time.Minute)
	c.ConfigureWarmUp(WarmUpConfig{
		MinCompletions: 2,
		MinDuration:    50 * time.Millisecond,
	})

	// Satisfy completions but not duration.
	c.RecordGameComplete(1*time.Second, 5)
	c.RecordGameComplete(1*time.Second, 5)

	c.RecordE2E(50 * time.Millisecond)

	snap := c.Snapshot()
	if snap.E2EMove.Count != 0 {
		t.Errorf("E2E should be gated when only completions met, got count=%d", snap.E2EMove.Count)
	}

	// Wait for duration.
	time.Sleep(60 * time.Millisecond)

	// Need another game complete to trigger the check after duration is met.
	c.RecordGameComplete(1*time.Second, 5)

	c.RecordE2E(50 * time.Millisecond)

	snap = c.Snapshot()
	if snap.E2EMove.Count != 1 {
		t.Errorf("E2E should record when both triggers met, got count=%d", snap.E2EMove.Count)
	}
}

func TestCollector_NoWarmUp_BackwardCompat(t *testing.T) {
	c := NewCollector(1 * time.Minute)

	// No ConfigureWarmUp called — should behave exactly as before.
	c.RecordE2E(50 * time.Millisecond)
	c.RecordREST("deploy", 30*time.Millisecond)
	c.RecordWSDelivery(10 * time.Millisecond)
	c.RecordPhaseLatency("attack", 20*time.Millisecond)

	snap := c.Snapshot()
	if snap.E2EMove.Count != 1 {
		t.Errorf("E2E should record without warm-up config, got count=%d", snap.E2EMove.Count)
	}

	if snap.RESTLatency["deploy"].Count != 1 {
		t.Errorf(
			"REST should record without warm-up config, got count=%d",
			snap.RESTLatency["deploy"].Count,
		)
	}

	if snap.WSDelivery.Count != 1 {
		t.Errorf("WS should record without warm-up config, got count=%d", snap.WSDelivery.Count)
	}

	if snap.PhaseLatency["attack"].Count != 1 {
		t.Errorf(
			"Phase should record without warm-up config, got count=%d",
			snap.PhaseLatency["attack"].Count,
		)
	}

	if !snap.WarmUpComplete {
		t.Error("WarmUpComplete should be true when no warm-up configured")
	}
}
