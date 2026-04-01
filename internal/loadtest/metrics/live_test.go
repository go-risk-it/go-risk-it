package metrics //nolint:testpackage // whitebox tests access unexported helpers

import (
	"context"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/loadtest/health"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestLiveMetrics() (*LiveMetrics, *testUpDownCounter) {
	gamesActive := &testUpDownCounter{}
	exporter := &OTelExporter{}
	exporter.initTestCounters()

	lm := &LiveMetrics{
		gamesActive: gamesActive,
		exporter:    exporter,
	}

	return lm, gamesActive
}

func TestLiveMetrics_RecordGameStarted(t *testing.T) {
	t.Parallel()
	lm, counter := newTestLiveMetrics()

	lm.RecordGameStarted()
	assert.Equal(t, int64(1), counter.value)

	lm.RecordGameStarted()
	assert.Equal(t, int64(2), counter.value)
}

func TestLiveMetrics_RecordGameStopped(t *testing.T) {
	t.Parallel()
	lm, counter := newTestLiveMetrics()

	// Start 3 games, stop 2.
	lm.RecordGameStarted()
	lm.RecordGameStarted()
	lm.RecordGameStarted()
	lm.RecordGameStopped()
	lm.RecordGameStopped()

	assert.Equal(t, int64(1), counter.value)
}

func TestLiveMetrics_RecordGameStopped_ToZero(t *testing.T) {
	t.Parallel()
	lm, counter := newTestLiveMetrics()

	lm.RecordGameStarted()
	lm.RecordGameStopped()

	assert.Equal(t, int64(0), counter.value)
}

func TestLiveMetrics_RecordHealthDistribution(t *testing.T) {
	t.Parallel()
	lm, _ := newTestLiveMetrics()

	dist := health.Distribution{Healthy: 5, Slow: 2, Stalled: 1, Zombie: 0, Total: 8}
	lm.RecordHealthDistribution(dist)

	// Verify delegation: the exporter's health counters should reflect the distribution.
	assertHealthCounters(t, lm.exporter, 5, 2, 1, 0)
	assertPrevHealth(t, lm.exporter, dist)
}

func TestLiveMetrics_ResetHealthCounters(t *testing.T) {
	t.Parallel()
	lm, _ := newTestLiveMetrics()

	dist := health.Distribution{Healthy: 5, Slow: 3, Stalled: 2, Zombie: 1, Total: 11}
	lm.RecordHealthDistribution(dist)
	lm.ResetHealthCounters()

	assertHealthCounters(t, lm.exporter, 0, 0, 0, 0)
	assertPrevHealth(t, lm.exporter, health.Distribution{})
}

func TestLiveMetrics_Shutdown(t *testing.T) {
	t.Parallel()
	// Shutdown on a nil LiveMetrics should not panic.
	var lm *LiveMetrics
	require.NoError(t, lm.Shutdown(context.Background()))
}

func TestLiveMetrics_NilSafe(t *testing.T) {
	t.Parallel()
	var lm *LiveMetrics

	// None of these should panic.
	assert.NotPanics(t, func() { lm.RecordGameStarted() })
	assert.NotPanics(t, func() { lm.RecordGameStopped() })
	assert.NotPanics(t, func() {
		lm.RecordHealthDistribution(health.Distribution{Healthy: 1})
	})
	assert.NotPanics(t, func() { lm.ResetHealthCounters() })
	assert.NotPanics(t, func() {
		_ = lm.Shutdown(context.Background())
	})
}
