package orchestrator

import (
	"context"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/loadtest/baseline"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/dbstats"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/health"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/metrics"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/resources"
)

// StepOutput holds raw data from a single staircase step.
type StepOutput struct {
	TargetGames        int
	Snapshot           *metrics.Snapshot
	Duration           time.Duration
	ServerResources    resources.ServerResources
	HealthDistribution *health.Distribution
	DBStats            *dbstats.StepDBStats
}

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
