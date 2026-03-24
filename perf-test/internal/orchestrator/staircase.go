package orchestrator

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/annotations"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/baseline"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/dbstats"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/health"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/metrics"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/resources"
)

// StaircaseConfig defines the staircase run parameters.
type StaircaseConfig struct {
	Steps             []int         // target concurrent games per step
	HoldDuration      time.Duration // how long to hold each step
	NumPlayers        int
	GameTimeout       time.Duration
	StopOnBreach      bool          // stop when SLOs fail
	StaggerDelay      time.Duration // between initial game launches within a step
	SLOs              baseline.SLOSet
	CooldownSec       int // seconds between steps (default 5)
	WarmUpCompletions int // games to complete before recording histograms (0 = disabled)
	WarmUpDurationSec int // seconds to wait before recording histograms (0 = disabled)
}

// cooldown returns the effective cooldown duration.
func (c StaircaseConfig) cooldown() time.Duration {
	if c.CooldownSec > 0 {
		return time.Duration(c.CooldownSec) * time.Second
	}

	return 5 * time.Second
}

// StepOutput holds raw data from a single staircase step.
type StepOutput struct {
	TargetGames        int
	Snapshot           *metrics.Snapshot
	Duration           time.Duration
	ServerResources    resources.ServerResources
	HealthDistribution *health.Distribution
	DBStats            *dbstats.StepDBStats
}

// StaircaseDeps holds injected dependencies for testability.
type StaircaseDeps struct {
	RunnerFactory    func(collector *metrics.Collector, observer GameObserver) RunFunc
	NewCollector     func(maxDuration time.Duration) *metrics.Collector
	CollectResources func() resources.ServerResources
	Annotator        *annotations.Annotator
	OTelExporter     *metrics.OTelExporter // shared across all steps, may be nil
	DBStats          *dbstats.Collector    // optional — nil disables DB stats collection
}

// RunStaircase executes the staircase: for each step, creates a pool at target
// concurrency, holds for HoldDuration, snapshots metrics, evaluates SLOs.
// Stops on breach if configured. Returns one StepOutput per executed step
// (including the breach step).
func RunStaircase(
	ctx context.Context,
	cfg StaircaseConfig,
	deps StaircaseDeps,
) []StepOutput {
	var outputs []StepOutput

	indexOffset := 0

	for i, targetGames := range cfg.Steps {
		select {
		case <-ctx.Done():
			log.Printf("[staircase] cancelled at step %d/%d", i+1, len(cfg.Steps))

			return outputs
		default:
		}

		log.Printf(
			"[staircase] step %d/%d: %d concurrent games, hold %v",
			i+1,
			len(cfg.Steps),
			targetGames,
			cfg.HoldDuration,
		)

		// Fresh collector per step for clean per-step percentiles.
		collector := deps.NewCollector(cfg.HoldDuration)

		// Configure warm-up filtering if specified.
		if cfg.WarmUpCompletions > 0 || cfg.WarmUpDurationSec > 0 {
			collector.ConfigureWarmUp(metrics.WarmUpConfig{
				MinCompletions: int64(cfg.WarmUpCompletions),
				MinDuration:    time.Duration(cfg.WarmUpDurationSec) * time.Second,
			})
		}

		// Share the single OTelExporter across all steps.
		if deps.OTelExporter != nil {
			collector.SetOTelExporter(deps.OTelExporter)
		}

		// Create health tracker for this step.
		tracker := health.NewTracker(health.DefaultThresholds())
		observer := health.NewTrackerObserver(tracker)

		runFunc := deps.RunnerFactory(collector, observer)

		output := runStep(ctx, cfg, deps, targetGames, indexOffset, i, runFunc, collector, tracker)
		outputs = append(outputs, output)

		indexOffset += targetGames * 10 // generous offset for replacement games

		// Check SLOs if stop-on-breach is enabled.
		if cfg.StopOnBreach {
			metricsSnap := baseline.SnapshotToMetrics(
				output.Snapshot,
				output.Duration.Seconds(),
			)
			evalResult := cfg.SLOs.Evaluate(metricsSnap)

			if !evalResult.AllPassing() {
				log.Printf(
					"[staircase] SLO breach at %d games, stopping",
					targetGames,
				)
				deps.Annotator.Annotate(
					fmt.Sprintf("staircase: SLO breach at %d games", targetGames),
					"perf-test",
					"alert",
				)

				break
			}
		}

		// Cooldown between steps.
		if i < len(cfg.Steps)-1 {
			cooldown := cfg.cooldown()

			log.Printf("[staircase] cooldown %v before next step", cooldown)

			select {
			case <-ctx.Done():
				return outputs
			case <-time.After(cooldown):
			}
		}
	}

	return outputs
}

// runStep executes a single staircase step: pool up, hold, snapshot, drain.
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

	// Reset DB stats counters after pool reaches steady state (best-effort).
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

	// Snapshot DB stats before draining (best-effort).
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
