package orchestrator_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/loadtest/annotations"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/baseline"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/metrics"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/orchestrator"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/resources"
	"github.com/stretchr/testify/assert"
)

func makeAdaptiveExecutor(
	t *testing.T,
	failAbove int,
	maxSteps int,
) *orchestrator.DefaultStepExecutor {
	t.Helper()

	cfg := orchestrator.StepExecutorConfig{
		NumPlayers:   2,
		GameTimeout:  time.Second,
		StaggerDelay: 5 * time.Millisecond,
		HoldDuration: 50 * time.Millisecond,
	}

	deps := orchestrator.StepExecutorDeps{
		RunnerFactory: func(c *metrics.StepAccumulator, _ orchestrator.GameObserver) orchestrator.RunFunc {
			return func(ctx context.Context, idx, players int) orchestrator.GameResult {
				// Record a move to produce metrics — large latency to trigger
				// SLO breach when failAbove > 0.
				c.RecordMove()
				c.RecordTimedMove()

				if failAbove > 0 {
					c.RecordE2E(500 * time.Millisecond)
				} else {
					c.RecordE2E(5 * time.Millisecond)
				}

				select {
				case <-ctx.Done():
				case <-time.After(10 * time.Millisecond):
				}

				return orchestrator.GameResult{GameIndex: idx}
			}
		},
		NewCollector: metrics.NewStepAccumulator,
		CollectResources: func() resources.ServerResources {
			return resources.ServerResources{}
		},
		Annotator: annotations.NewAnnotator(""),
	}

	return orchestrator.NewStepExecutor(cfg, deps, maxSteps)
}

func TestRunAdaptive_ProbesUpward(t *testing.T) {
	t.Parallel()

	cfg := orchestrator.AdaptiveConfig{
		InitialCeiling:   5,
		AdditiveIncrease: 5,
		SLOs:             baseline.DefaultSLOs(),
		MaxSteps:         4,
		MaxGames:         500,
		CooldownSec:      1,
	}

	executor := makeAdaptiveExecutor(t, 0, cfg.MaxSteps)
	result := orchestrator.RunAdaptive(context.Background(), cfg, executor)

	// Should have probed: 5, 10, 15, 20 (4 steps = MaxSteps).
	assert.Len(t, result.Steps, 4)
	assert.True(t, result.Converged)
	assert.Equal(t, 20, result.Ceiling)
}

func TestRunAdaptive_BisectsOnBreach(t *testing.T) {
	t.Parallel()

	cfg := orchestrator.AdaptiveConfig{
		InitialCeiling:   10,
		AdditiveIncrease: 5,
		SLOs: baseline.SLOSet{
			UserExperience: []baseline.SLO{
				{
					Name:      "impossible",
					Metric:    "e2e_p95_s",
					Threshold: 0.001,
					Unit:      "s",
				},
			},
		},
		MaxSteps:    10,
		MaxGames:    500,
		CooldownSec: 1,
	}

	executor := makeAdaptiveExecutor(t, 1, cfg.MaxSteps)
	result := orchestrator.RunAdaptive(context.Background(), cfg, executor)

	// First step at 10 breaches, lastGood=0.
	// Binary search: 0 to 10, AI=5 → converges quickly with ceiling=0.
	assert.Equal(t, 0, result.Ceiling)
	assert.True(t, result.Converged)
}

func TestRunAdaptive_ColdStart(t *testing.T) {
	t.Parallel()

	cfg := orchestrator.AdaptiveConfig{
		InitialCeiling:   0, // should start at 10
		AdditiveIncrease: 5,
		SLOs:             baseline.DefaultSLOs(),
		MaxSteps:         2,
		MaxGames:         500,
		CooldownSec:      1,
	}

	executor := makeAdaptiveExecutor(t, 0, cfg.MaxSteps)
	result := orchestrator.RunAdaptive(context.Background(), cfg, executor)

	// Should start at 10, probe to 15.
	assert.Len(t, result.Steps, 2)
	assert.Equal(t, 10, result.Steps[0].TargetGames)
}

func TestRunAdaptive_Convergence(t *testing.T) {
	t.Parallel()

	cfg := orchestrator.AdaptiveConfig{
		InitialCeiling:   10,
		AdditiveIncrease: 5,
		SLOs: baseline.SLOSet{
			UserExperience: []baseline.SLO{
				{
					Name:      "impossible",
					Metric:    "e2e_p95_s",
					Threshold: 0.001,
					Unit:      "s",
				},
			},
		},
		MaxSteps:    20,
		MaxGames:    500,
		CooldownSec: 1,
	}

	executor := makeAdaptiveExecutor(t, 1, cfg.MaxSteps)
	result := orchestrator.RunAdaptive(context.Background(), cfg, executor)

	// Should converge.
	assert.True(t, result.Converged)
}

func TestRunAdaptive_MaxSteps(t *testing.T) {
	t.Parallel()

	cfg := orchestrator.AdaptiveConfig{
		InitialCeiling:   5,
		AdditiveIncrease: 5,
		SLOs:             baseline.DefaultSLOs(),
		MaxSteps:         3,
		MaxGames:         500,
		CooldownSec:      1,
	}

	executor := makeAdaptiveExecutor(t, 0, cfg.MaxSteps)
	result := orchestrator.RunAdaptive(context.Background(), cfg, executor)

	// Should never exceed MaxSteps.
	assert.LessOrEqual(t, len(result.Steps), 3)
}
