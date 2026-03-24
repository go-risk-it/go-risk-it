package orchestrator

import (
	"context"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/baseline"
)

// StepExecutor runs a single staircase step at the given concurrency level.
type StepExecutor interface {
	Execute(ctx context.Context, targetGames, indexOffset int) (*StepOutput, error)
}

// StopCondition determines whether a staircase should stop after a step.
type StopCondition interface {
	ShouldStop(output *StepOutput) bool
}

// NeverStop is a StopCondition that never stops.
type NeverStop struct{}

// ShouldStop always returns false.
func (NeverStop) ShouldStop(*StepOutput) bool { return false }

// SLOStopCondition stops when SLOs are breached.
type SLOStopCondition struct {
	SLOs baseline.SLOSet
}

// ShouldStop returns true if any SLO is breached in the step output.
func (s *SLOStopCondition) ShouldStop(output *StepOutput) bool {
	if output == nil || output.Snapshot == nil {
		return false
	}

	snap := baseline.SnapshotToMetrics(output.Snapshot, output.Duration.Seconds())

	return !s.SLOs.Evaluate(snap).AllPassing()
}
