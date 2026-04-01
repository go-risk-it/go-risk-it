package metrics

// ConfigureWarmUp enables warm-up gating on the accumulator.
// When configured, histogram recording is blocked until MarkWarmUpDone is called.
// Counters continue recording during warm-up.
// Safe for concurrent use.
func (a *StepAccumulator) ConfigureWarmUp() {
	a.warmUpConfigured.Store(true)
}

// MarkWarmUpDone opens the warm-up gate, allowing histograms to record.
// Safe to call multiple times (idempotent).
func (a *StepAccumulator) MarkWarmUpDone() {
	a.warmUpDone.Store(true)
}

// isWarmUpDone returns true if warm-up is not configured or has been marked done.
func (a *StepAccumulator) isWarmUpDone() bool {
	if !a.warmUpConfigured.Load() {
		return true
	}

	return a.warmUpDone.Load()
}
