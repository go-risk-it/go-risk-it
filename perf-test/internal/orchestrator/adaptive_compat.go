package orchestrator

// Temporary compatibility layer — StaircaseDeps and runStep are used by
// adaptive.go until it is refactored to use StepExecutor (Task 9).

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/annotations"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/dbstats"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/health"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/metrics"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/resources"
)

// StaircaseDeps holds injected dependencies for testability.
// Deprecated: use StepExecutorDeps instead; kept for adaptive.go compatibility.
type StaircaseDeps struct {
	RunnerFactory    func(collector *metrics.Collector, observer GameObserver) RunFunc
	NewCollector     func(maxDuration time.Duration) *metrics.Collector
	CollectResources func() resources.ServerResources
	Annotator        *annotations.Annotator
	OTelExporter     *metrics.OTelExporter
	DBStats          *dbstats.Collector
}

// runStep executes a single step: pool up, hold, snapshot, drain.
// Deprecated: use DefaultStepExecutor.Execute instead; kept for adaptive.go compatibility.
func runStep(
	ctx context.Context,
	cfg StaircaseConfig,
	deps StaircaseDeps,
	targetGames int,
	indexOffset int,
	stepIndex int,
	runFunc RunFunc,
	collector *metrics.Collector,
	tracker *health.Tracker,
) StepOutput {
	stepCtx, stepCancel := context.WithCancel(ctx)
	defer stepCancel()

	pool := NewPool(
		PoolConfig{
			TargetGames:  targetGames,
			NumPlayers:   cfg.NumPlayers,
			StaggerDelay: cfg.StaggerDelay,
			IndexOffset:  indexOffset,
		},
		runFunc,
	)

	go pool.Run(stepCtx)

	// Wait for pool to reach target concurrency.
	select {
	case <-pool.Ready():
		log.Printf(
			"[staircase] step %d: pool ready (%d games active)",
			stepIndex+1,
			targetGames,
		)
	case <-ctx.Done():
		stepCancel()
		pool.WaitDrain()

		return StepOutput{TargetGames: targetGames}
	}

	// Reset DB stats counters after pool reaches steady state.
	if deps.DBStats != nil {
		if err := deps.DBStats.Reset(ctx); err != nil {
			log.Printf("[staircase] step %d: db stats reset: %v", stepIndex+1, err)
		}
	}

	deps.Annotator.Annotate(
		fmt.Sprintf("staircase: step %d/%d — %d games", stepIndex+1, len(cfg.Steps), targetGames),
		"perf-test",
		"step",
	)

	// Hold for the configured duration.
	holdStart := time.Now()

	select {
	case <-time.After(cfg.HoldDuration):
	case <-ctx.Done():
	}

	holdDuration := time.Since(holdStart)

	// Collect server resources before draining.
	serverResources := deps.CollectResources()

	// Snapshot metrics while games are still running.
	snap := collector.Snapshot()

	// Snapshot health distribution before draining.
	healthDist := tracker.Snapshot()

	// Snapshot DB stats before draining.
	var stepDBStats *dbstats.StepDBStats

	if deps.DBStats != nil {
		stats, err := deps.DBStats.Snapshot(ctx, 10)
		if err != nil {
			log.Printf("[staircase] step %d: db stats snapshot: %v", stepIndex+1, err)
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
