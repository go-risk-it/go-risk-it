package orchestrator_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/loadtest/annotations"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/baseline"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/metrics"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/orchestrator"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/resources"
	"github.com/stretchr/testify/assert"
)

// makeTestStepExecutor builds a DefaultStepExecutor suitable for staircase tests.
func makeTestStepExecutor(
	t *testing.T,
	totalSteps int,
	collectorCalls *atomic.Int64,
) *orchestrator.DefaultStepExecutor {
	t.Helper()

	cfg := orchestrator.StepExecutorConfig{
		NumPlayers:   2,
		GameTimeout:  time.Second,
		StaggerDelay: 5 * time.Millisecond,
		HoldDuration: 100 * time.Millisecond,
	}

	deps := orchestrator.StepExecutorDeps{
		RunnerFactory: func(
			c *metrics.StepAccumulator,
			_ orchestrator.GameObserver,
		) orchestrator.RunFunc {
			return func(ctx context.Context, idx, players int) orchestrator.GameResult {
				select {
				case <-ctx.Done():
				case <-time.After(10 * time.Millisecond):
				}

				return orchestrator.GameResult{GameIndex: idx}
			}
		},
		NewCollector: func(d time.Duration) *metrics.StepAccumulator {
			if collectorCalls != nil {
				collectorCalls.Add(1)
			}

			return metrics.NewStepAccumulator(d)
		},
		CollectResources: func() resources.ServerResources {
			return resources.ServerResources{}
		},
		Annotator: annotations.NewAnnotator(""),
	}

	return orchestrator.NewStepExecutor(cfg, deps, totalSteps)
}

func TestRunStaircase_AllStepsPassing(t *testing.T) {
	t.Parallel()

	var collectorCalls atomic.Int64

	steps := []int{2, 4}
	executor := makeTestStepExecutor(t, len(steps), &collectorCalls)
	stopper := &orchestrator.NeverStop{}
	annotator := annotations.NewAnnotator("")

	outputs := orchestrator.RunStaircase(
		context.Background(), steps, 1, executor, stopper, annotator,
	)

	assert.Len(t, outputs, 2)
	assert.Equal(t, int64(2), collectorCalls.Load())
}

func TestRunStaircase_StopOnBreach(t *testing.T) {
	t.Parallel()

	steps := []int{2, 4, 8}

	// Build executor that records high latency to trigger SLO breach.
	cfg := orchestrator.StepExecutorConfig{
		NumPlayers:   2,
		GameTimeout:  time.Second,
		StaggerDelay: 5 * time.Millisecond,
		HoldDuration: 100 * time.Millisecond,
	}

	deps := orchestrator.StepExecutorDeps{
		RunnerFactory: func(
			c *metrics.StepAccumulator,
			_ orchestrator.GameObserver,
		) orchestrator.RunFunc {
			return func(ctx context.Context, idx, players int) orchestrator.GameResult {
				c.RecordMove()
				c.RecordTimedMove()
				c.RecordE2E(100 * time.Millisecond)

				select {
				case <-ctx.Done():
				case <-time.After(10 * time.Millisecond):
				}

				return orchestrator.GameResult{GameIndex: idx}
			}
		},
		NewCollector:     metrics.NewStepAccumulator,
		CollectResources: func() resources.ServerResources { return resources.ServerResources{} },
		Annotator:        annotations.NewAnnotator(""),
	}

	executor := orchestrator.NewStepExecutor(cfg, deps, len(steps))

	// Impossible SLO threshold.
	stopper := &orchestrator.SLOStopCondition{
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
	annotator := annotations.NewAnnotator("")

	outputs := orchestrator.RunStaircase(
		context.Background(), steps, 1, executor, stopper, annotator,
	)

	// Should stop at step 1 (first breach), returning 1 output.
	assert.Len(t, outputs, 1)
	assert.Equal(t, 2, outputs[0].TargetGames)
}

func TestRunStaircase_ContinueOnBreach(t *testing.T) {
	t.Parallel()

	steps := []int{2, 4, 8}
	executor := makeTestStepExecutor(t, len(steps), nil)

	outputs := orchestrator.RunStaircase(
		context.Background(), steps, 1,
		executor, &orchestrator.NeverStop{}, annotations.NewAnnotator(""),
	)

	assert.Len(t, outputs, 3)
}

func TestRunStaircase_ContextCancellation(t *testing.T) {
	t.Parallel()

	// Use long hold so cancellation hits during a step.
	cfg := orchestrator.StepExecutorConfig{
		NumPlayers:   2,
		GameTimeout:  time.Second,
		StaggerDelay: 5 * time.Millisecond,
		HoldDuration: 5 * time.Second,
	}

	deps := orchestrator.StepExecutorDeps{
		RunnerFactory: func(
			c *metrics.StepAccumulator,
			_ orchestrator.GameObserver,
		) orchestrator.RunFunc {
			return func(ctx context.Context, idx, players int) orchestrator.GameResult {
				select {
				case <-ctx.Done():
				case <-time.After(10 * time.Millisecond):
				}

				return orchestrator.GameResult{GameIndex: idx}
			}
		},
		NewCollector:     metrics.NewStepAccumulator,
		CollectResources: func() resources.ServerResources { return resources.ServerResources{} },
		Annotator:        annotations.NewAnnotator(""),
	}

	steps := []int{2, 4, 8}
	executor := orchestrator.NewStepExecutor(cfg, deps, len(steps))

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	outputs := orchestrator.RunStaircase(
		ctx, steps, 1, executor, &orchestrator.NeverStop{}, annotations.NewAnnotator(""),
	)

	assert.LessOrEqual(t, len(outputs), 1)
}

func TestRunStaircase_FreshCollectorPerStep(t *testing.T) {
	t.Parallel()

	var collectorCalls atomic.Int64

	steps := []int{2, 4, 8}

	cfg := orchestrator.StepExecutorConfig{
		NumPlayers:   2,
		GameTimeout:  time.Second,
		StaggerDelay: 5 * time.Millisecond,
		HoldDuration: 50 * time.Millisecond,
	}

	deps := orchestrator.StepExecutorDeps{
		RunnerFactory: func(
			c *metrics.StepAccumulator,
			_ orchestrator.GameObserver,
		) orchestrator.RunFunc {
			return func(ctx context.Context, idx, players int) orchestrator.GameResult {
				select {
				case <-ctx.Done():
				case <-time.After(10 * time.Millisecond):
				}

				return orchestrator.GameResult{GameIndex: idx}
			}
		},
		NewCollector: func(d time.Duration) *metrics.StepAccumulator {
			collectorCalls.Add(1)

			return metrics.NewStepAccumulator(d)
		},
		CollectResources: func() resources.ServerResources { return resources.ServerResources{} },
		Annotator:        annotations.NewAnnotator(""),
	}

	executor := orchestrator.NewStepExecutor(cfg, deps, len(steps))

	outputs := orchestrator.RunStaircase(
		context.Background(), steps, 1,
		executor, &orchestrator.NeverStop{}, annotations.NewAnnotator(""),
	)

	assert.Len(t, outputs, 3)
	assert.Equal(t, int64(3), collectorCalls.Load())
}

func TestRunStaircase_ResourcesCollected(t *testing.T) {
	t.Parallel()

	cfg := orchestrator.StepExecutorConfig{
		NumPlayers:   2,
		GameTimeout:  time.Second,
		StaggerDelay: 5 * time.Millisecond,
		HoldDuration: 50 * time.Millisecond,
	}

	deps := orchestrator.StepExecutorDeps{
		RunnerFactory: func(
			c *metrics.StepAccumulator,
			_ orchestrator.GameObserver,
		) orchestrator.RunFunc {
			return func(ctx context.Context, idx, players int) orchestrator.GameResult {
				select {
				case <-ctx.Done():
				case <-time.After(10 * time.Millisecond):
				}

				return orchestrator.GameResult{GameIndex: idx}
			}
		},
		NewCollector: metrics.NewStepAccumulator,
		CollectResources: func() resources.ServerResources {
			return resources.ServerResources{
				RiskIt: resources.ContainerStats{CPUPercent: 42.0, MemoryMB: 256},
			}
		},
		Annotator: annotations.NewAnnotator(""),
	}

	steps := []int{2}
	executor := orchestrator.NewStepExecutor(cfg, deps, len(steps))

	outputs := orchestrator.RunStaircase(
		context.Background(), steps, 0,
		executor, &orchestrator.NeverStop{}, annotations.NewAnnotator(""),
	)

	assert.Len(t, outputs, 1)
	assert.InDelta(t, 42.0, outputs[0].ServerResources.RiskIt.CPUPercent, 0.01)
}
