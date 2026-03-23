package orchestrator_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/annotations"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/baseline"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/metrics"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/orchestrator"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/resources"
	"github.com/stretchr/testify/assert"
)

func makeFakeDeps(
	t *testing.T,
	collectorCalls *atomic.Int64,
) orchestrator.StaircaseDeps {
	t.Helper()

	return orchestrator.StaircaseDeps{
		RunnerFactory: func(c *metrics.Collector) orchestrator.RunFunc {
			return func(
				ctx context.Context,
				idx, players int,
			) orchestrator.GameResult {
				select {
				case <-ctx.Done():
				case <-time.After(10 * time.Millisecond):
				}

				return orchestrator.GameResult{GameIndex: idx}
			}
		},
		NewCollector: func(d time.Duration) *metrics.Collector {
			if collectorCalls != nil {
				collectorCalls.Add(1)
			}

			return metrics.NewCollector(d)
		},
		CollectResources: func() resources.ServerResources {
			return resources.ServerResources{}
		},
		Annotator: annotations.NewAnnotator(""), // no-op
	}
}

func TestRunStaircase_AllStepsPassing(t *testing.T) {
	t.Parallel()

	var collectorCalls atomic.Int64

	cfg := orchestrator.StaircaseConfig{
		Steps:        []int{2, 4},
		HoldDuration: 100 * time.Millisecond,
		NumPlayers:   2,
		GameTimeout:  time.Second,
		StopOnBreach: true,
		StaggerDelay: 5 * time.Millisecond,
		SLOs:         baseline.DefaultSLOs(),
		CooldownSec:  1,
	}

	deps := makeFakeDeps(t, &collectorCalls)
	outputs := orchestrator.RunStaircase(context.Background(), cfg, deps)
	assert.Len(t, outputs, 2)
	assert.Equal(t, int64(2), collectorCalls.Load())
}

func TestRunStaircase_StopOnBreach(t *testing.T) {
	t.Parallel()

	stepCount := 0

	cfg := orchestrator.StaircaseConfig{
		Steps:        []int{2, 4, 8},
		HoldDuration: 100 * time.Millisecond,
		NumPlayers:   2,
		GameTimeout:  time.Second,
		StopOnBreach: true,
		StaggerDelay: 5 * time.Millisecond,
		SLOs: baseline.SLOSet{
			UserExperience: []baseline.SLO{
				// Impossible threshold: E2E p95 < 0.000001s — will always fail.
				{
					Name:      "impossible",
					Metric:    "e2e_p95_s",
					Threshold: 0.000001,
					Unit:      "s",
				},
			},
		},
		CooldownSec: 1,
	}

	deps := orchestrator.StaircaseDeps{
		RunnerFactory: func(c *metrics.Collector) orchestrator.RunFunc {
			return func(
				ctx context.Context,
				idx, players int,
			) orchestrator.GameResult {
				// Record a move to produce metrics.
				c.RecordMove()
				c.RecordTimedMove()
				c.RecordE2E(100 * time.Millisecond) // p95 will be 100ms > 0.000001s threshold

				select {
				case <-ctx.Done():
				case <-time.After(10 * time.Millisecond):
				}

				return orchestrator.GameResult{GameIndex: idx}
			}
		},
		NewCollector: func(d time.Duration) *metrics.Collector {
			stepCount++

			return metrics.NewCollector(d)
		},
		CollectResources: func() resources.ServerResources {
			return resources.ServerResources{}
		},
		Annotator: annotations.NewAnnotator(""),
	}

	outputs := orchestrator.RunStaircase(context.Background(), cfg, deps)
	// Should stop at step 1 (first breach), returning 1 output.
	assert.Equal(t, 1, len(outputs))
	assert.Equal(t, 2, outputs[0].TargetGames)
}

func TestRunStaircase_ContinueOnBreach(t *testing.T) {
	t.Parallel()

	cfg := orchestrator.StaircaseConfig{
		Steps:        []int{2, 4, 8},
		HoldDuration: 50 * time.Millisecond,
		NumPlayers:   2,
		GameTimeout:  time.Second,
		StopOnBreach: false,
		StaggerDelay: 5 * time.Millisecond,
		SLOs:         baseline.DefaultSLOs(),
		CooldownSec:  1,
	}

	deps := makeFakeDeps(t, nil)
	outputs := orchestrator.RunStaircase(context.Background(), cfg, deps)
	assert.Len(t, outputs, 3) // all 3 run regardless of SLOs
}

func TestRunStaircase_ContextCancellation(t *testing.T) {
	t.Parallel()

	cfg := orchestrator.StaircaseConfig{
		Steps:        []int{2, 4, 8},
		HoldDuration: 5 * time.Second, // long hold, will be cancelled
		NumPlayers:   2,
		GameTimeout:  time.Second,
		StopOnBreach: false,
		StaggerDelay: 5 * time.Millisecond,
		SLOs:         baseline.DefaultSLOs(),
		CooldownSec:  1,
	}

	deps := makeFakeDeps(t, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	outputs := orchestrator.RunStaircase(ctx, cfg, deps)
	// Should get at most 1 step (cancelled during hold).
	assert.LessOrEqual(t, len(outputs), 1)
}

func TestRunStaircase_FreshCollectorPerStep(t *testing.T) {
	t.Parallel()

	var collectorCalls atomic.Int64

	cfg := orchestrator.StaircaseConfig{
		Steps:        []int{2, 4, 8},
		HoldDuration: 50 * time.Millisecond,
		NumPlayers:   2,
		GameTimeout:  time.Second,
		StopOnBreach: false,
		StaggerDelay: 5 * time.Millisecond,
		SLOs:         baseline.DefaultSLOs(),
		CooldownSec:  1,
	}

	deps := makeFakeDeps(t, &collectorCalls)
	outputs := orchestrator.RunStaircase(context.Background(), cfg, deps)
	assert.Len(t, outputs, 3)
	assert.Equal(t, int64(3), collectorCalls.Load())
}

func TestRunStaircase_ResourcesCollected(t *testing.T) {
	t.Parallel()

	cfg := orchestrator.StaircaseConfig{
		Steps:        []int{2},
		HoldDuration: 50 * time.Millisecond,
		NumPlayers:   2,
		GameTimeout:  time.Second,
		StopOnBreach: false,
		StaggerDelay: 5 * time.Millisecond,
		SLOs:         baseline.DefaultSLOs(),
	}

	deps := makeFakeDeps(t, nil)
	deps.CollectResources = func() resources.ServerResources {
		return resources.ServerResources{
			RiskIt: resources.ContainerStats{CPUPercent: 42.0, MemoryMB: 256},
		}
	}

	outputs := orchestrator.RunStaircase(context.Background(), cfg, deps)
	assert.Len(t, outputs, 1)
	assert.InDelta(t, 42.0, outputs[0].ServerResources.RiskIt.CPUPercent, 0.01)
}
