package orchestrator_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/annotations"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/baseline"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/metrics"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/orchestrator"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/resources"
	"github.com/stretchr/testify/assert"
)

func makeAdaptiveDeps(
	t *testing.T,
	failAbove int,
) orchestrator.StaircaseDeps {
	t.Helper()

	return orchestrator.StaircaseDeps{
		RunnerFactory: func(c *metrics.Collector, _ orchestrator.GameObserver) orchestrator.RunFunc {
			return func(ctx context.Context, idx, players int) orchestrator.GameResult {
				// Record a move to produce metrics — large latency to trigger
				// SLO breach when failAbove > 0.
				c.RecordMove()
				c.RecordTimedMove()

				if failAbove > 0 {
					// Record high latency to make SLOs fail at the configured threshold.
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
		NewCollector: metrics.NewCollector,
		CollectResources: func() resources.ServerResources {
			return resources.ServerResources{}
		},
		Annotator: annotations.NewAnnotator(""),
	}
}

func TestRunAdaptive_ProbesUpward(t *testing.T) {
	t.Parallel()

	cfg := orchestrator.AdaptiveConfig{
		InitialCeiling:   5,
		AdditiveIncrease: 5,
		HoldDuration:     50 * time.Millisecond,
		NumPlayers:       2,
		GameTimeout:      time.Second,
		StaggerDelay:     5 * time.Millisecond,
		SLOs:             baseline.DefaultSLOs(),
		MaxSteps:         4,
		MaxGames:         500,
		CooldownSec:      1,
	}

	// No SLO failures — should probe up.
	deps := makeAdaptiveDeps(t, 0)
	result := orchestrator.RunAdaptive(context.Background(), cfg, deps)

	// Should have probed: 5, 10, 15, 20 (4 steps = MaxSteps).
	assert.Len(t, result.Steps, 4)
	assert.True(t, result.Converged)
	assert.Equal(t, 20, result.Ceiling)
}

func TestRunAdaptive_BisectsOnBreach(t *testing.T) {
	t.Parallel()

	// Fail above any threshold — first step will breach.
	cfg := orchestrator.AdaptiveConfig{
		InitialCeiling:   10,
		AdditiveIncrease: 5,
		HoldDuration:     50 * time.Millisecond,
		NumPlayers:       2,
		GameTimeout:      time.Second,
		StaggerDelay:     5 * time.Millisecond,
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

	deps := makeAdaptiveDeps(t, 1)
	result := orchestrator.RunAdaptive(context.Background(), cfg, deps)

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
		HoldDuration:     50 * time.Millisecond,
		NumPlayers:       2,
		GameTimeout:      time.Second,
		StaggerDelay:     5 * time.Millisecond,
		SLOs:             baseline.DefaultSLOs(),
		MaxSteps:         2,
		MaxGames:         500,
		CooldownSec:      1,
	}

	deps := makeAdaptiveDeps(t, 0)
	result := orchestrator.RunAdaptive(context.Background(), cfg, deps)

	// Should start at 10, probe to 15.
	assert.Len(t, result.Steps, 2)
	assert.Equal(t, 10, result.Steps[0].TargetGames)
}

func TestRunAdaptive_Convergence(t *testing.T) {
	t.Parallel()

	cfg := orchestrator.AdaptiveConfig{
		InitialCeiling:   10,
		AdditiveIncrease: 5,
		HoldDuration:     50 * time.Millisecond,
		NumPlayers:       2,
		GameTimeout:      time.Second,
		StaggerDelay:     5 * time.Millisecond,
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

	deps := makeAdaptiveDeps(t, 1)
	result := orchestrator.RunAdaptive(context.Background(), cfg, deps)

	// Should converge.
	assert.True(t, result.Converged)
}

func TestRunAdaptive_MaxSteps(t *testing.T) {
	t.Parallel()

	cfg := orchestrator.AdaptiveConfig{
		InitialCeiling:   5,
		AdditiveIncrease: 5,
		HoldDuration:     50 * time.Millisecond,
		NumPlayers:       2,
		GameTimeout:      time.Second,
		StaggerDelay:     5 * time.Millisecond,
		SLOs:             baseline.DefaultSLOs(),
		MaxSteps:         3,
		MaxGames:         500,
		CooldownSec:      1,
	}

	deps := makeAdaptiveDeps(t, 0)
	result := orchestrator.RunAdaptive(context.Background(), cfg, deps)

	// Should never exceed MaxSteps.
	assert.LessOrEqual(t, len(result.Steps), 3)
}
