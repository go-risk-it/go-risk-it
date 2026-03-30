package metrics

// ConfigureWarmUp enables warm-up gating on the collector.
// When configured, histogram recording is blocked until MarkWarmUpDone is called.
// Counters and OTel instruments continue recording during warm-up.
// Must be called before any recording begins.
func (c *Collector) ConfigureWarmUp() {
	c.warmUpConfigured = true
}

// MarkWarmUpDone opens the warm-up gate, allowing histograms to record.
// Safe to call multiple times (idempotent).
func (c *Collector) MarkWarmUpDone() {
	c.warmUpDone.Store(true)
}

// isWarmUpDone returns true if warm-up is not configured or has been marked done.
func (c *Collector) isWarmUpDone() bool {
	if !c.warmUpConfigured {
		return true
	}

	return c.warmUpDone.Load()
}
