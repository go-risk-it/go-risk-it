package health //nolint:testpackage // whitebox tests access unexported helpers

import (
	"sync"
	"testing"
	"time"
)

// fakeClock returns a controllable clock function. Calling advance() moves the
// clock forward by the given duration. The returned now() is safe for concurrent
// use within a single test but shares state — don't use across parallel tests.
func fakeClock(start time.Time) (func() time.Time, func(d time.Duration)) {
	current := start

	return func() time.Time {
			return current
		}, func(d time.Duration) {
			current = current.Add(d)
		}
}

func newTestTracker(thresholds Thresholds, now func() time.Time) *Tracker {
	t := NewTracker(thresholds)
	t.now = now

	return t
}

func TestTracker_HealthyGame(t *testing.T) {
	t.Parallel()
	now, _ := fakeClock(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	tracker := newTestTracker(DefaultThresholds(), now)
	tracker.RegisterGame(1)

	// Moves at regular intervals → classified Healthy.
	tracker.RecordMove(1, "deploy")
	tracker.RecordMove(1, "attack")

	dist := tracker.Snapshot()
	if dist.Healthy != 1 {
		t.Errorf("expected 1 healthy, got %d", dist.Healthy)
	}

	if dist.Total != 1 {
		t.Errorf("expected total 1, got %d", dist.Total)
	}
}

func TestTracker_SlowGame(t *testing.T) {
	t.Parallel()
	now, advance := fakeClock(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	tracker := newTestTracker(Thresholds{
		SlowMultiplier:    1.5,
		StalledMultiplier: 3.0,
		ZombieAge:         10 * time.Second, // disable zombie detection
	}, now)

	// Create a baseline: 20ms between second moves.
	// Each game: register (t=X), move deploy (t=X, interval ~0), advance 20ms, move attack (interval=20ms).
	// Mean interval = (0 + 20ms) / 2 = 10ms per game pair.
	for i := range 5 {
		tracker.RegisterGame(i)
		tracker.RecordMove(i, "deploy")
		advance(20 * time.Millisecond)
		tracker.RecordMove(i, "attack")
		tracker.CompleteGame(i)
	}

	// Mean interval = 10ms.
	// Slow threshold: 1.5 * 10ms = 15ms.
	// Stalled threshold: 3 * 10ms = 30ms.
	// Register a game and advance 20ms (> 15ms, < 30ms) → slow.
	tracker.RegisterGame(100)
	tracker.RecordMove(100, "deploy")
	advance(20 * time.Millisecond)

	dist := tracker.Snapshot()
	if dist.Slow != 1 {
		t.Errorf("expected 1 slow, got %d (healthy=%d, stalled=%d, zombie=%d)",
			dist.Slow, dist.Healthy, dist.Stalled, dist.Zombie)
	}
}

func TestTracker_StalledGame(t *testing.T) {
	t.Parallel()
	now, advance := fakeClock(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	tracker := newTestTracker(Thresholds{
		SlowMultiplier:    1.5,
		StalledMultiplier: 3.0,
		ZombieAge:         10 * time.Second, // disable zombie detection
	}, now)

	// Establish baseline: 20ms between moves → mean 10ms.
	for i := range 5 {
		tracker.RegisterGame(i)
		tracker.RecordMove(i, "deploy")
		advance(20 * time.Millisecond)
		tracker.RecordMove(i, "attack")
		tracker.CompleteGame(i)
	}

	// Mean interval 10ms. Stalled threshold: 3 * 10ms = 30ms.
	// Register a game and advance 40ms (> 30ms) → stalled.
	tracker.RegisterGame(100)
	tracker.RecordMove(100, "deploy")
	advance(40 * time.Millisecond)

	dist := tracker.Snapshot()
	if dist.Stalled != 1 {
		t.Errorf("expected 1 stalled, got %d (healthy=%d, slow=%d, zombie=%d)",
			dist.Stalled, dist.Healthy, dist.Slow, dist.Zombie)
	}
}

func TestTracker_ZombieGame(t *testing.T) {
	t.Parallel()
	now, advance := fakeClock(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	tracker := newTestTracker(Thresholds{
		SlowMultiplier:    1.5,
		StalledMultiplier: 3.0,
		ZombieAge:         50 * time.Millisecond, // explicit zombie threshold
	}, now)

	tracker.RegisterGame(1)
	tracker.RecordMove(1, "deploy")

	// Advance past zombie age.
	advance(60 * time.Millisecond)

	dist := tracker.Snapshot()
	if dist.Zombie != 1 {
		t.Errorf("expected 1 zombie, got %d (healthy=%d, slow=%d, stalled=%d)",
			dist.Zombie, dist.Healthy, dist.Slow, dist.Stalled)
	}
}

func TestTracker_DistributionCounts(t *testing.T) {
	t.Parallel()
	now, advance := fakeClock(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	tracker := newTestTracker(Thresholds{
		SlowMultiplier:    1.5,
		StalledMultiplier: 3.0,
		ZombieAge:         10 * time.Second, // disable zombie detection
	}, now)

	// Establish baseline: 20ms between moves → mean 10ms.
	for i := range 10 {
		tracker.RegisterGame(i)
		tracker.RecordMove(i, "deploy")
		advance(20 * time.Millisecond)
		tracker.RecordMove(i, "attack")
		tracker.CompleteGame(i)
	}

	// Register 3 games at known time offsets from their last move.
	// Stalled game: 40ms since last move (> 3 * 10ms).
	tracker.RegisterGame(103)
	tracker.RecordMove(103, "deploy")
	advance(40 * time.Millisecond)

	// Slow game: 20ms since last move (> 1.5 * 10ms, < 3 * 10ms).
	tracker.RegisterGame(102)
	tracker.RecordMove(102, "deploy")
	advance(20 * time.Millisecond)

	// Healthy game: just moved (0ms since last move).
	tracker.RegisterGame(101)
	tracker.RecordMove(101, "deploy")

	dist := tracker.Snapshot()
	if dist.Total != 3 {
		t.Errorf("expected total 3, got %d", dist.Total)
	}

	// Game 103 had its last move 60ms ago (40ms + 20ms from slow game's advance).
	// Game 102 had its last move 20ms ago.
	// Game 101 just moved (0ms).
	// With mean interval 10ms: stalled > 30ms, slow > 15ms.
	// So game 103 is stalled (60ms > 30ms), game 102 is slow (20ms > 15ms), game 101 is healthy.
	if dist.Stalled != 1 {
		t.Errorf("expected 1 stalled, got %d", dist.Stalled)
	}

	if dist.Slow != 1 {
		t.Errorf("expected 1 slow, got %d", dist.Slow)
	}

	if dist.Healthy != 1 {
		t.Errorf("expected 1 healthy, got %d", dist.Healthy)
	}
}

func TestTracker_EffectiveConcurrency(t *testing.T) {
	t.Parallel()
	dist := Distribution{
		Healthy: 5,
		Slow:    3,
		Stalled: 1,
		Zombie:  1,
		Total:   10,
	}

	if got := dist.EffectiveConcurrency(); got != 8 {
		t.Errorf("EffectiveConcurrency: expected 8 (5+3), got %d", got)
	}
}

func TestTracker_CompleteGameRemoves(t *testing.T) {
	t.Parallel()
	now, _ := fakeClock(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	tracker := newTestTracker(DefaultThresholds(), now)

	tracker.RegisterGame(1)
	tracker.RegisterGame(2)
	tracker.RecordMove(1, "deploy")
	tracker.RecordMove(2, "deploy")

	if dist := tracker.Snapshot(); dist.Total != 2 {
		t.Fatalf("expected 2 active, got %d", dist.Total)
	}

	tracker.CompleteGame(1)

	dist := tracker.Snapshot()
	if dist.Total != 1 {
		t.Errorf("expected 1 active after completion, got %d", dist.Total)
	}
}

func TestTracker_Concurrent(t *testing.T) {
	t.Parallel()
	// Concurrent test uses real clock — the deterministic clock isn't safe
	// for concurrent goroutines advancing time independently.
	tracker := NewTracker(DefaultThresholds())

	var wg sync.WaitGroup

	// 10 goroutines each registering, recording, and completing games.
	for g := range 10 {
		wg.Add(1)

		go func(base int) {
			defer wg.Done()

			for i := range 10 {
				idx := base*100 + i
				tracker.RegisterGame(idx)
				tracker.RecordMove(idx, "deploy")
				tracker.RecordPhaseChange(idx, "attack")
				tracker.RecordMove(idx, "attack")
				tracker.CompleteGame(idx)
			}
		}(g)
	}

	// Concurrent snapshots.
	for range 5 {
		wg.Go(func() {
			for range 20 {
				tracker.Snapshot()
			}
		})
	}

	wg.Wait()

	// After all games complete, snapshot should show 0 active.
	dist := tracker.Snapshot()
	if dist.Total != 0 {
		t.Errorf("expected 0 active games after all complete, got %d", dist.Total)
	}
}
