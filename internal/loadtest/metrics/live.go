package metrics

import (
	"context"

	"github.com/go-risk-it/go-risk-it/internal/loadtest/health"
)

// LiveMetrics is a thin facade over cross-step OTel state.
// It owns the gamesActive UpDownCounter and delegates health classification
// recording to the underlying OTelExporter.
//
// LiveMetrics is nil-safe: a nil receiver is a no-op for all methods,
// enabling tests without OTel infrastructure.
type LiveMetrics struct {
	gamesActive healthCounter
	exporter    *OTelExporter
}

// NewLiveMetrics creates a LiveMetrics that wraps the given OTelExporter.
// The gamesActive counter is extracted from the exporter so LiveMetrics
// owns the increment/decrement lifecycle while the exporter retains
// provider shutdown responsibility.
func NewLiveMetrics(exporter *OTelExporter) *LiveMetrics {
	return &LiveMetrics{
		gamesActive: &otelUpDownCounter{counter: exporter.gamesActive},
		exporter:    exporter,
	}
}

// RecordGameStarted increments the active games gauge.
func (lm *LiveMetrics) RecordGameStarted() {
	if lm == nil {
		return
	}

	lm.gamesActive.Add(1)
}

// RecordGameStopped decrements the active games gauge.
// Call on game completion, timeout, or fatal error.
func (lm *LiveMetrics) RecordGameStopped() {
	if lm == nil {
		return
	}

	lm.gamesActive.Add(-1)
}

// RecordGameCancelled increments the cancelled games counter.
func (lm *LiveMetrics) RecordGameCancelled() {
	if lm == nil {
		return
	}

	lm.exporter.RecordGameCancelled()
}

// RecordHealthDistribution delegates health classification reporting to the exporter.
func (lm *LiveMetrics) RecordHealthDistribution(dist health.Distribution) {
	if lm == nil {
		return
	}

	lm.exporter.RecordHealthDistribution(dist)
}

// ResetHealthCounters zeroes health counters between staircase steps.
func (lm *LiveMetrics) ResetHealthCounters() {
	if lm == nil {
		return
	}

	lm.exporter.ResetHealthCounters()
}

// Shutdown flushes and stops both the meter and tracer providers.
func (lm *LiveMetrics) Shutdown(ctx context.Context) error {
	if lm == nil {
		return nil
	}

	return lm.exporter.Shutdown(ctx)
}
