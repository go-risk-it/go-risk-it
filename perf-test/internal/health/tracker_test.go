package health

import (
	"sync"
	"testing"
	"time"
)

func TestTracker_HealthyGame(t *testing.T) {
	tracker := NewTracker(DefaultThresholds())
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
	tracker := NewTracker(Thresholds{
		SlowMultiplier:    1.5,
		StalledMultiplier: 3.0,
		ZombieAge:         10 * time.Second, // disable zombie detection
	})

	// Create a baseline: ~20ms between moves (register→move is near-zero,
	// so only the second interval contributes meaningfully).
	for i := 0; i < 5; i++ {
		tracker.RegisterGame(i)
		tracker.RecordMove(i, "deploy")
		time.Sleep(20 * time.Millisecond)
		tracker.RecordMove(i, "attack")
		tracker.CompleteGame(i)
	}

	// Mean interval is ~10ms (avg of ~0ms first + ~20ms second).
	// Slow threshold: 1.5 * 10ms = 15ms.
	// Stalled threshold: 3 * 10ms = 30ms.
	// Register a game and let it sit for 20ms (> 15ms, < 30ms).
	tracker.RegisterGame(100)
	tracker.RecordMove(100, "deploy")
	time.Sleep(20 * time.Millisecond)

	dist := tracker.Snapshot()
	if dist.Slow != 1 {
		t.Errorf("expected 1 slow, got %d (healthy=%d, stalled=%d, zombie=%d)",
			dist.Slow, dist.Healthy, dist.Stalled, dist.Zombie)
	}
}

func TestTracker_StalledGame(t *testing.T) {
	tracker := NewTracker(Thresholds{
		SlowMultiplier:    1.5,
		StalledMultiplier: 3.0,
		ZombieAge:         10 * time.Second, // disable zombie detection
	})

	// Establish baseline: ~20ms between moves.
	for i := 0; i < 5; i++ {
		tracker.RegisterGame(i)
		tracker.RecordMove(i, "deploy")
		time.Sleep(20 * time.Millisecond)
		tracker.RecordMove(i, "attack")
		tracker.CompleteGame(i)
	}

	// Mean interval ~10ms. Stalled threshold: 3 * 10ms = 30ms.
	// Register a game and let it sit for 40ms (> 30ms).
	tracker.RegisterGame(100)
	tracker.RecordMove(100, "deploy")
	time.Sleep(40 * time.Millisecond)

	dist := tracker.Snapshot()
	if dist.Stalled != 1 {
		t.Errorf("expected 1 stalled, got %d (healthy=%d, slow=%d, zombie=%d)",
			dist.Stalled, dist.Healthy, dist.Slow, dist.Zombie)
	}
}

func TestTracker_ZombieGame(t *testing.T) {
	tracker := NewTracker(Thresholds{
		SlowMultiplier:    1.5,
		StalledMultiplier: 3.0,
		ZombieAge:         50 * time.Millisecond, // explicit zombie threshold
	})

	tracker.RegisterGame(1)
	tracker.RecordMove(1, "deploy")

	// Wait past zombie age.
	time.Sleep(60 * time.Millisecond)

	dist := tracker.Snapshot()
	if dist.Zombie != 1 {
		t.Errorf("expected 1 zombie, got %d (healthy=%d, slow=%d, stalled=%d)",
			dist.Zombie, dist.Healthy, dist.Slow, dist.Stalled)
	}
}

func TestTracker_DistributionCounts(t *testing.T) {
	tracker := NewTracker(Thresholds{
		SlowMultiplier:    1.5,
		StalledMultiplier: 3.0,
		ZombieAge:         10 * time.Second, // disable zombie detection
	})

	// Establish baseline: ~20ms between moves → mean ~10ms.
	for i := 0; i < 10; i++ {
		tracker.RegisterGame(i)
		tracker.RecordMove(i, "deploy")
		time.Sleep(20 * time.Millisecond)
		tracker.RecordMove(i, "attack")
		tracker.CompleteGame(i)
	}

	// Register 3 games that will be in different states.
	// Stalled game first (needs longest sleep).
	tracker.RegisterGame(103)
	tracker.RecordMove(103, "deploy")
	time.Sleep(40 * time.Millisecond) // > 3 * 10ms = stalled

	// Slow game.
	tracker.RegisterGame(102)
	tracker.RecordMove(102, "deploy")
	time.Sleep(20 * time.Millisecond) // > 1.5 * 10ms = slow

	// Healthy game: just moved.
	tracker.RegisterGame(101)
	tracker.RecordMove(101, "deploy")

	dist := tracker.Snapshot()
	if dist.Total != 3 {
		t.Errorf("expected total 3, got %d", dist.Total)
	}

	if dist.Stalled < 1 {
		t.Errorf("expected at least 1 stalled, got %d", dist.Stalled)
	}
}

func TestTracker_EffectiveConcurrency(t *testing.T) {
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
	tracker := NewTracker(DefaultThresholds())

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
	tracker := NewTracker(DefaultThresholds())

	var wg sync.WaitGroup

	// 10 goroutines each registering, recording, and completing games.
	for g := 0; g < 10; g++ {
		wg.Add(1)

		go func(base int) {
			defer wg.Done()

			for i := 0; i < 10; i++ {
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
	for s := 0; s < 5; s++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for i := 0; i < 20; i++ {
				tracker.Snapshot()
			}
		}()
	}

	wg.Wait()

	// After all games complete, snapshot should show 0 active.
	dist := tracker.Snapshot()
	if dist.Total != 0 {
		t.Errorf("expected 0 active games after all complete, got %d", dist.Total)
	}
}
