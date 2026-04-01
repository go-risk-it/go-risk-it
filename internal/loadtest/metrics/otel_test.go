package metrics

import (
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/loadtest/health"
)

func TestRecordHealthDistribution_FromZero(t *testing.T) {
	// When prevHealth is zero-valued, all counts are new deltas.
	o := &OTelExporter{}
	o.initTestCounters()

	dist := health.Distribution{Healthy: 5, Slow: 2, Stalled: 1, Zombie: 0, Total: 8}
	o.RecordHealthDistribution(dist)

	assertHealthCounters(t, o, 5, 2, 1, 0)
	assertPrevHealth(t, o, dist)
}

func TestRecordHealthDistribution_Deltas(t *testing.T) {
	tests := []struct {
		name        string
		prev        health.Distribution
		next        health.Distribution
		wantHealthy int64
		wantSlow    int64
		wantStalled int64
		wantZombie  int64
	}{
		{
			name:        "all increase",
			prev:        health.Distribution{Healthy: 5, Slow: 2, Stalled: 1, Zombie: 0},
			next:        health.Distribution{Healthy: 8, Slow: 3, Stalled: 2, Zombie: 1},
			wantHealthy: 8, wantSlow: 3, wantStalled: 2, wantZombie: 1,
		},
		{
			name:        "some decrease (games moved from healthy to stalled)",
			prev:        health.Distribution{Healthy: 10, Slow: 0, Stalled: 0, Zombie: 0},
			next:        health.Distribution{Healthy: 7, Slow: 1, Stalled: 2, Zombie: 0},
			wantHealthy: 7, wantSlow: 1, wantStalled: 2, wantZombie: 0,
		},
		{
			name:        "all decrease to zero",
			prev:        health.Distribution{Healthy: 5, Slow: 3, Stalled: 2, Zombie: 1},
			next:        health.Distribution{Healthy: 0, Slow: 0, Stalled: 0, Zombie: 0},
			wantHealthy: 0, wantSlow: 0, wantStalled: 0, wantZombie: 0,
		},
		{
			name:        "no change",
			prev:        health.Distribution{Healthy: 5, Slow: 2, Stalled: 1, Zombie: 1},
			next:        health.Distribution{Healthy: 5, Slow: 2, Stalled: 1, Zombie: 1},
			wantHealthy: 5, wantSlow: 2, wantStalled: 1, wantZombie: 1,
		},
		{
			name:        "large swing",
			prev:        health.Distribution{Healthy: 100, Slow: 0, Stalled: 0, Zombie: 0},
			next:        health.Distribution{Healthy: 0, Slow: 0, Stalled: 0, Zombie: 100},
			wantHealthy: 0, wantSlow: 0, wantStalled: 0, wantZombie: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &OTelExporter{}
			o.initTestCounters()

			// Record the "prev" state first.
			o.RecordHealthDistribution(tt.prev)
			// Record the "next" state — deltas should apply.
			o.RecordHealthDistribution(tt.next)

			assertHealthCounters(t, o, tt.wantHealthy, tt.wantSlow, tt.wantStalled, tt.wantZombie)
			assertPrevHealth(t, o, tt.next)
		})
	}
}

func TestRecordHealthDistribution_MultipleSequential(t *testing.T) {
	// Simulate a realistic sequence of distributions during a staircase hold.
	o := &OTelExporter{}
	o.initTestCounters()

	steps := []health.Distribution{
		{Healthy: 10, Slow: 0, Stalled: 0, Zombie: 0}, // pool just started
		{Healthy: 8, Slow: 2, Stalled: 0, Zombie: 0},  // some games slowing
		{Healthy: 6, Slow: 2, Stalled: 2, Zombie: 0},  // degradation
		{Healthy: 6, Slow: 2, Stalled: 1, Zombie: 1},  // a stalled became zombie
		{Healthy: 9, Slow: 1, Stalled: 0, Zombie: 0},  // recovery
	}

	for _, dist := range steps {
		o.RecordHealthDistribution(dist)
	}

	// Final counters should reflect the last distribution.
	last := steps[len(steps)-1]
	assertHealthCounters(
		t,
		o,
		int64(last.Healthy),
		int64(last.Slow),
		int64(last.Stalled),
		int64(last.Zombie),
	)
}

func TestResetHealthCounters(t *testing.T) {
	o := &OTelExporter{}
	o.initTestCounters()

	// Build up some state.
	o.RecordHealthDistribution(health.Distribution{Healthy: 5, Slow: 3, Stalled: 2, Zombie: 1})

	assertHealthCounters(t, o, 5, 3, 2, 1)

	// Reset should zero everything.
	o.ResetHealthCounters()

	assertHealthCounters(t, o, 0, 0, 0, 0)
	assertPrevHealth(t, o, health.Distribution{})
}

func TestResetHealthCounters_FromZero(t *testing.T) {
	// Resetting when already at zero should be a no-op.
	o := &OTelExporter{}
	o.initTestCounters()

	o.ResetHealthCounters()

	assertHealthCounters(t, o, 0, 0, 0, 0)
}

func TestResetHealthCounters_ThenRecord(t *testing.T) {
	o := &OTelExporter{}
	o.initTestCounters()

	// Build state, reset, then record new state.
	o.RecordHealthDistribution(health.Distribution{Healthy: 10, Slow: 0, Stalled: 0, Zombie: 0})
	o.ResetHealthCounters()
	o.RecordHealthDistribution(health.Distribution{Healthy: 3, Slow: 1, Stalled: 0, Zombie: 0})

	// Counters should reflect only the post-reset distribution.
	assertHealthCounters(t, o, 3, 1, 0, 0)
}

// --- test helpers ---

// testUpDownCounter is a fake UpDownCounter that tracks net additions for assertion.
type testUpDownCounter struct {
	value int64
}

func (c *testUpDownCounter) Add(delta int64) {
	c.value += delta
}

// initTestCounters wires up test fakes so we can assert counter values
// without a real OTel meter provider.
func (o *OTelExporter) initTestCounters() {
	o.healthHealthy = &testUpDownCounter{}
	o.healthSlow = &testUpDownCounter{}
	o.healthStalled = &testUpDownCounter{}
	o.healthZombie = &testUpDownCounter{}
}

func getTestCounter(counter healthCounter) *testUpDownCounter {
	c, ok := counter.(*testUpDownCounter)
	if !ok {
		return nil
	}

	return c
}

func assertHealthCounters(t *testing.T, o *OTelExporter, wantH, wantS, wantSt, wantZ int64) {
	t.Helper()

	if got := getTestCounter(o.healthHealthy).value; got != wantH {
		t.Errorf("healthHealthy: want %d, got %d", wantH, got)
	}

	if got := getTestCounter(o.healthSlow).value; got != wantS {
		t.Errorf("healthSlow: want %d, got %d", wantS, got)
	}

	if got := getTestCounter(o.healthStalled).value; got != wantSt {
		t.Errorf("healthStalled: want %d, got %d", wantSt, got)
	}

	if got := getTestCounter(o.healthZombie).value; got != wantZ {
		t.Errorf("healthZombie: want %d, got %d", wantZ, got)
	}
}

func assertPrevHealth(t *testing.T, o *OTelExporter, want health.Distribution) {
	t.Helper()

	if o.prevHealth != want {
		t.Errorf("prevHealth: want %+v, got %+v", want, o.prevHealth)
	}
}
