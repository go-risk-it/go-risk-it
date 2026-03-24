package chaos_test

import (
	"testing"
	"time"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/chaos"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/gamestate"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/metrics"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/player"
)

// fakeStrategy is a trivial strategy for testing the chaos wrapper.
type fakeStrategy struct{}

func (f *fakeStrategy) Name() string { return "fake" }

func (f *fakeStrategy) DecideMove(
	_ gamestate.ViewSnapshot,
	_ string,
) (*player.Action, error) {
	return &player.Action{Type: player.ActionAdvance}, nil
}

func TestStrategy_Name(t *testing.T) {
	t.Parallel()

	collector := metrics.NewCollector(1 * time.Minute)
	s := chaos.WrapStrategy(&fakeStrategy{}, chaos.Config{}, collector)

	if got := s.Name(); got != "chaos(fake)" {
		t.Fatalf("Name() = %q, want %q", got, "chaos(fake)")
	}
}

func TestStrategy_NoInjection(t *testing.T) {
	t.Parallel()

	collector := metrics.NewCollector(1 * time.Minute)
	cfg := chaos.Config{
		ErrorMoveRate: 0,
		SlowMoveRate:  0,
	}
	s := chaos.WrapStrategy(&fakeStrategy{}, cfg, collector)

	snap := gamestate.ViewSnapshot{}

	action, err := s.DecideMove(snap, "user1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if action.Type != player.ActionAdvance {
		t.Fatalf("expected ActionAdvance, got %d", action.Type)
	}
}

func TestStrategy_AlwaysError(t *testing.T) {
	t.Parallel()

	collector := metrics.NewCollector(1 * time.Minute)
	cfg := chaos.Config{
		ErrorMoveRate: 1.0, // always inject error
	}
	s := chaos.WrapStrategy(&fakeStrategy{}, cfg, collector)

	_, err := s.DecideMove(gamestate.ViewSnapshot{}, "user1")
	if err == nil {
		t.Fatal("expected error from chaos strategy")
	}

	// Verify chaos event was recorded.
	snap := collector.Snapshot()
	if snap.ChaosEvents[string(metrics.ChaosEventErrorMove)] != 1 {
		t.Fatalf(
			"expected 1 error_move chaos event, got %d",
			snap.ChaosEvents[string(metrics.ChaosEventErrorMove)],
		)
	}
}

func TestStrategy_AlwaysSlow(t *testing.T) {
	t.Parallel()

	collector := metrics.NewCollector(1 * time.Minute)
	cfg := chaos.Config{
		SlowMoveRate:  1.0, // always slow
		SlowMoveDelay: 10 * time.Millisecond,
	}
	s := chaos.WrapStrategy(&fakeStrategy{}, cfg, collector)

	start := time.Now()

	action, err := s.DecideMove(gamestate.ViewSnapshot{}, "user1")

	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if action.Type != player.ActionAdvance {
		t.Fatalf("expected ActionAdvance, got %d", action.Type)
	}

	if elapsed < 10*time.Millisecond {
		t.Fatalf("expected delay of at least 10ms, got %v", elapsed)
	}

	snap := collector.Snapshot()
	if snap.ChaosEvents[string(metrics.ChaosEventSlowMove)] != 1 {
		t.Fatalf(
			"expected 1 slow_move chaos event, got %d",
			snap.ChaosEvents[string(metrics.ChaosEventSlowMove)],
		)
	}
}
