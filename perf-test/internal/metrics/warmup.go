package metrics

import "time"

// WarmUpConfig configures warm-up filtering for histogram recording.
// Both triggers must be satisfied before histograms start recording:
// - MinCompletions games must complete (pool has churned its first generation)
// - MinDuration must elapse since warm-up was configured
type WarmUpConfig struct {
	MinCompletions int64
	MinDuration    time.Duration
}

// ConfigureWarmUp sets the warm-up filtering configuration.
// Must be called before any recording begins.
// When configured, histogram recording is gated until both triggers are met.
// Counters and OTel instruments continue recording during warm-up.
func (c *Collector) ConfigureWarmUp(cfg WarmUpConfig) {
	c.warmUpConfig = &cfg
	c.warmUpStart = time.Now()
}

// isWarmUpDone returns true if warm-up filtering is not configured or has completed.
func (c *Collector) isWarmUpDone() bool {
	if c.warmUpConfig == nil {
		return true
	}

	return c.warmUpDone.Load()
}

// checkWarmUpCompletion checks if both warm-up triggers have been met.
// Called after each game completion.
func (c *Collector) checkWarmUpCompletion() {
	if c.warmUpConfig == nil || c.warmUpDone.Load() {
		return
	}

	completions := c.warmUpCompletions.Load()
	elapsed := time.Since(c.warmUpStart)

	if completions >= c.warmUpConfig.MinCompletions && elapsed >= c.warmUpConfig.MinDuration {
		c.warmUpDone.Store(true)
	}
}
