package orchestrator

import (
	"context"
	"fmt"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/annotations"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/dbstats"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/health"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/metrics"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/resources"
	"go.opentelemetry.io/otel/attribute"
)

const defaultHealthPollInterval = 5 * time.Second

// StepExecutorConfig holds parameters for each step execution.
type StepExecutorConfig struct {
	NumPlayers         int
	GameTimeout        time.Duration
	StaggerDelay       time.Duration
	HoldDuration       time.Duration
	HealthPollInterval time.Duration // 0 = defaultHealthPollInterval (5s)
}

// StepExecutorDeps holds injected dependencies for testability.
type StepExecutorDeps struct {
	RunnerFactory    func(collector *metrics.StepAccumulator, observer GameObserver) RunFunc
	NewCollector     func(maxDuration time.Duration) *metrics.StepAccumulator
	CollectResources func() resources.ServerResources
	Annotator        *annotations.Annotator
	LiveMetrics      *metrics.LiveMetrics
	DBStats          *dbstats.Collector
}

// DefaultStepExecutor implements StepExecutor by creating a fresh pool per step,
// holding for the configured duration, then snapshotting metrics.
type DefaultStepExecutor struct {
	cfg        StepExecutorConfig
	deps       StepExecutorDeps
	totalSteps int
	stepCount  int
}

// NewStepExecutor creates a DefaultStepExecutor.
func NewStepExecutor(
	cfg StepExecutorConfig,
	deps StepExecutorDeps,
	totalSteps int,
) *DefaultStepExecutor {
	return &DefaultStepExecutor{
		cfg:        cfg,
		deps:       deps,
		totalSteps: totalSteps,
	}
}

// Execute runs a single staircase step at the given concurrency level.
func (e *DefaultStepExecutor) Execute(
	ctx context.Context,
	targetGames, indexOffset int,
) (*StepOutput, error) { //nolint:unparam // interface conformance / future use
	e.stepCount++
	stepIndex := e.stepCount

	// Reset health counters so each step starts from zero.
	e.deps.LiveMetrics.ResetHealthCounters()

	// Fresh collector per step for clean per-step percentiles.
	collector := e.deps.NewCollector(e.cfg.HoldDuration)
	collector.ConfigureWarmUp()

	tracker := health.NewTracker(health.DefaultThresholds())
	observer := health.NewTrackerObserver(tracker)
	runFunc := e.deps.RunnerFactory(collector, observer)

	output := e.runStep(ctx, targetGames, indexOffset, stepIndex, runFunc, collector, tracker)

	return &output, nil
}

// runStep executes a single step: pool up, hold, snapshot, drain.
//
//nolint:cyclop,funlen // sequential step execution
func (e *DefaultStepExecutor) runStep(
	ctx context.Context,
	targetGames int,
	indexOffset int,
	stepIndex int,
	runFunc RunFunc,
	collector *metrics.StepAccumulator,
	tracker *health.Tracker,
) StepOutput {
	stepCtx, stepCancel := context.WithCancel(ctx)
	defer stepCancel()

	pool := NewPool(
		PoolConfig{
			TargetGames:  targetGames,
			NumPlayers:   e.cfg.NumPlayers,
			StaggerDelay: e.cfg.StaggerDelay,
			IndexOffset:  indexOffset,
		},
		runFunc,
	)

	go pool.Run(stepCtx)

	// Wait for pool to reach target concurrency.
	select {
	case <-pool.Ready():
		collector.MarkWarmUpDone()

		observe.Info(ctx, "pool ready",
			attribute.Int("step", stepIndex),
			attribute.Int("total_steps", e.totalSteps),
			attribute.Int("games_active", targetGames),
		)
	case <-ctx.Done():
		stepCancel()
		pool.WaitDrain()

		return StepOutput{TargetGames: targetGames}
	}

	// Reset DB stats counters after pool reaches steady state.
	if e.deps.DBStats != nil {
		if err := e.deps.DBStats.Reset(ctx); err != nil {
			observe.Error(ctx, err, "db stats reset failed",
				attribute.Int("step", stepIndex),
				attribute.Int("total_steps", e.totalSteps),
			)
		}
	}

	e.deps.Annotator.Annotate(
		fmt.Sprintf("step %d/%d — %d games", stepIndex, e.totalSteps, targetGames),
		"perf-test",
		"step",
	)

	// Hold for the configured duration, periodically reporting health to OTel.
	holdStart := time.Now()

	pollInterval := e.cfg.HealthPollInterval
	if pollInterval <= 0 {
		pollInterval = defaultHealthPollInterval
	}

	holdTimer := time.NewTimer(e.cfg.HoldDuration)
	defer holdTimer.Stop()

	healthTicker := time.NewTicker(pollInterval)
	defer healthTicker.Stop()

holdLoop:
	for {
		select {
		case <-holdTimer.C:
			break holdLoop
		case <-healthTicker.C:
			e.deps.LiveMetrics.RecordHealthDistribution(tracker.Snapshot())
		case <-ctx.Done():
			break holdLoop
		}
	}

	// Emit final health distribution after the hold completes.
	e.deps.LiveMetrics.RecordHealthDistribution(tracker.Snapshot())

	holdDuration := time.Since(holdStart)

	// Collect server resources before draining.
	serverResources := e.deps.CollectResources()

	// Snapshot metrics while games are still running.
	snap := collector.Snapshot()

	// Snapshot health distribution before draining.
	healthDist := tracker.Snapshot()

	// Snapshot DB stats before draining.
	var stepDBStats *dbstats.StepDBStats

	if e.deps.DBStats != nil {
		stats, err := e.deps.DBStats.Snapshot(ctx, 10)
		if err != nil {
			observe.Error(ctx, err, "db stats snapshot failed",
				attribute.Int("step", stepIndex),
				attribute.Int("total_steps", e.totalSteps),
			)
		} else {
			stepDBStats = &stats
		}
	}

	// Drain the pool.
	stepCancel()
	pool.WaitDrain()

	return StepOutput{
		TargetGames:        targetGames,
		Snapshot:           snap,
		Duration:           holdDuration,
		ServerResources:    serverResources,
		HealthDistribution: &healthDist,
		DBStats:            stepDBStats,
	}
}
