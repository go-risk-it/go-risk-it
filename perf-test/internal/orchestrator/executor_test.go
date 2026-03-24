package orchestrator_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/annotations"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/metrics"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/orchestrator"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/resources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeTestExecutor(t *testing.T, totalSteps int) *orchestrator.DefaultStepExecutor {
	t.Helper()

	cfg := orchestrator.StepExecutorConfig{
		NumPlayers:   2,
		GameTimeout:  time.Second,
		StaggerDelay: 5 * time.Millisecond,
		HoldDuration: 100 * time.Millisecond,
	}

	deps := orchestrator.StepExecutorDeps{
		RunnerFactory: func(
			c *metrics.Collector,
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
		NewCollector: metrics.NewCollector,
		CollectResources: func() resources.ServerResources {
			return resources.ServerResources{}
		},
		Annotator: annotations.NewAnnotator(""),
	}

	return orchestrator.NewStepExecutor(cfg, deps, totalSteps)
}

func TestDefaultStepExecutor_Basic(t *testing.T) {
	t.Parallel()

	executor := makeTestExecutor(t, 3)

	output, err := executor.Execute(context.Background(), 2, 0)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, 2, output.TargetGames)
	assert.NotNil(t, output.Snapshot)
	assert.Greater(t, output.Duration.Milliseconds(), int64(0))
}

func TestDefaultStepExecutor_Resources(t *testing.T) {
	t.Parallel()

	cfg := orchestrator.StepExecutorConfig{
		NumPlayers:   2,
		GameTimeout:  time.Second,
		StaggerDelay: 5 * time.Millisecond,
		HoldDuration: 50 * time.Millisecond,
	}

	deps := orchestrator.StepExecutorDeps{
		RunnerFactory: func(
			c *metrics.Collector,
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
		NewCollector: metrics.NewCollector,
		CollectResources: func() resources.ServerResources {
			return resources.ServerResources{
				RiskIt: resources.ContainerStats{CPUPercent: 42.0, MemoryMB: 256},
			}
		},
		Annotator: annotations.NewAnnotator(""),
	}

	executor := orchestrator.NewStepExecutor(cfg, deps, 1)

	output, err := executor.Execute(context.Background(), 2, 0)
	require.NoError(t, err)
	assert.InDelta(t, 42.0, output.ServerResources.RiskIt.CPUPercent, 0.01)
}

func TestDefaultStepExecutor_Health(t *testing.T) {
	t.Parallel()

	executor := makeTestExecutor(t, 1)

	output, err := executor.Execute(context.Background(), 2, 0)
	require.NoError(t, err)
	assert.NotNil(t, output.HealthDistribution)
}

func TestDefaultStepExecutor_ContextCancel(t *testing.T) {
	t.Parallel()

	executor := makeTestExecutor(t, 1)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	output, err := executor.Execute(ctx, 4, 0)
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, 4, output.TargetGames)
}

func TestDefaultStepExecutor_MultipleSteps(t *testing.T) {
	t.Parallel()

	executor := makeTestExecutor(t, 3)

	// Execute multiple steps sequentially.
	for i := range 3 {
		output, err := executor.Execute(context.Background(), (i+1)*2, i*20)
		require.NoError(t, err)
		assert.Equal(t, (i+1)*2, output.TargetGames)
	}
}
