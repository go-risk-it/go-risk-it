package orchestrator_test

import (
	"testing"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/loadtest/baseline"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/metrics"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/orchestrator"
	"github.com/stretchr/testify/assert"
)

func TestNeverStop(t *testing.T) {
	t.Parallel()

	ns := orchestrator.NeverStop{}
	assert.False(t, ns.ShouldStop(nil))
	assert.False(t, ns.ShouldStop(&orchestrator.StepOutput{}))
	assert.False(t, ns.ShouldStop(&orchestrator.StepOutput{
		TargetGames: 100,
		Snapshot:    metrics.NewStepAccumulator(time.Second).Snapshot(),
		Duration:    time.Second,
	}))
}

func TestSLOStopCondition_Passing(t *testing.T) {
	t.Parallel()

	// Default SLOs with empty snapshot — passes because all thresholds are loose.
	cond := &orchestrator.SLOStopCondition{SLOs: baseline.DefaultSLOs()}
	output := &orchestrator.StepOutput{
		TargetGames: 5,
		Snapshot:    metrics.NewStepAccumulator(time.Second).Snapshot(),
		Duration:    time.Second,
	}

	assert.False(t, cond.ShouldStop(output))
}

func TestSLOStopCondition_Breaching(t *testing.T) {
	t.Parallel()

	// Impossible SLO: p95 < 1 microsecond.
	cond := &orchestrator.SLOStopCondition{
		SLOs: baseline.SLOSet{
			UserExperience: []baseline.SLO{
				{
					Name:      "impossible",
					Metric:    "e2e_p95_s",
					Threshold: 0.000001,
					Unit:      "s",
				},
			},
		},
	}

	// Record latency that will breach.
	collector := metrics.NewStepAccumulator(time.Second)
	collector.RecordMove()
	collector.RecordTimedMove()
	collector.RecordE2E(100 * time.Millisecond)

	output := &orchestrator.StepOutput{
		TargetGames: 5,
		Snapshot:    collector.Snapshot(),
		Duration:    time.Second,
	}

	assert.True(t, cond.ShouldStop(output))
}

func TestSLOStopCondition_NilSnapshot(t *testing.T) {
	t.Parallel()

	cond := &orchestrator.SLOStopCondition{SLOs: baseline.DefaultSLOs()}

	// Nil output.
	assert.False(t, cond.ShouldStop(nil))

	// Output with nil snapshot.
	assert.False(t, cond.ShouldStop(&orchestrator.StepOutput{TargetGames: 5}))
}
