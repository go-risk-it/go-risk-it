package orchestrator

import (
	"context"
	"log"
	"time"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/baseline"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/health"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/metrics"
)

// AdaptiveConfig defines the TCP congestion control staircase parameters.
type AdaptiveConfig struct {
	InitialCeiling    int           // from session's last ceiling; 0 starts at 10
	AdditiveIncrease  int           // games to add per successful step (default 5)
	HoldDuration      time.Duration // how long to hold each level
	NumPlayers        int
	GameTimeout       time.Duration
	StaggerDelay      time.Duration
	SLOs              baseline.SLOSet
	MaxSteps          int // upper bound on steps (default 20)
	MaxGames          int // hard ceiling (default 500)
	CooldownSec       int
	WarmUpCompletions int
	WarmUpDurationSec int
}

// AdaptiveResult holds the outcome of an adaptive staircase run.
type AdaptiveResult struct {
	Steps     []StepOutput
	Ceiling   int
	Converged bool
}

func (c AdaptiveConfig) cooldown() time.Duration {
	if c.CooldownSec > 0 {
		return time.Duration(c.CooldownSec) * time.Second
	}

	return DefaultCooldownSec * time.Second
}

// RunAdaptive executes a TCP congestion control staircase:
// 1. Probe up from InitialCeiling, adding AdditiveIncrease per passing step.
// 2. On breach, binary search between lastGood and lastBad.
// 3. Converge when |lastBad - lastGood| <= AdditiveIncrease.
func RunAdaptive(
	ctx context.Context,
	cfg AdaptiveConfig,
	deps StaircaseDeps,
) AdaptiveResult {
	if cfg.AdditiveIncrease <= 0 {
		cfg.AdditiveIncrease = 5
	}

	if cfg.MaxSteps <= 0 {
		cfg.MaxSteps = 20
	}

	if cfg.MaxGames <= 0 {
		cfg.MaxGames = 500
	}

	initial := cfg.InitialCeiling
	if initial <= 0 {
		initial = 10
	}

	var result AdaptiveResult
	indexOffset := 0
	stepNum := 0

	// Phase 1: Probe up.
	current := initial
	lastGood := 0

	for stepNum < cfg.MaxSteps && current <= cfg.MaxGames {
		select {
		case <-ctx.Done():
			return result
		default:
		}

		log.Printf("[adaptive] probe step %d: %d games", stepNum+1, current)

		output, passed := runAdaptiveStep(ctx, cfg, deps, current, indexOffset, stepNum)
		result.Steps = append(result.Steps, output)
		indexOffset += current * IndexOffsetMultiplier
		stepNum++

		if !passed {
			log.Printf("[adaptive] breach at %d games, switching to binary search", current)

			// Phase 2: Binary search.
			lastBad := current
			result = binarySearch(
				ctx, cfg, deps, lastGood, lastBad,
				indexOffset, stepNum, result,
			)

			return result
		}

		lastGood = current
		current += cfg.AdditiveIncrease

		if stepNum < cfg.MaxSteps {
			cooldown(ctx, cfg.cooldown())
		}
	}

	// Never breached — ceiling is at lastGood.
	result.Ceiling = lastGood
	result.Converged = true

	return result
}

func binarySearch(
	ctx context.Context,
	cfg AdaptiveConfig,
	deps StaircaseDeps,
	lastGood, lastBad int,
	indexOffset, stepNum int,
	result AdaptiveResult,
) AdaptiveResult {
	for stepNum < cfg.MaxSteps && (lastBad-lastGood) > cfg.AdditiveIncrease {
		select {
		case <-ctx.Done():
			result.Ceiling = lastGood

			return result
		default:
		}

		mid := (lastGood + lastBad) / 2
		log.Printf(
			"[adaptive] bisect step %d: %d games (good=%d, bad=%d)",
			stepNum+1,
			mid,
			lastGood,
			lastBad,
		)

		output, passed := runAdaptiveStep(ctx, cfg, deps, mid, indexOffset, stepNum)
		result.Steps = append(result.Steps, output)
		indexOffset += mid * IndexOffsetMultiplier
		stepNum++

		if passed {
			lastGood = mid
		} else {
			lastBad = mid
		}

		cooldown(ctx, cfg.cooldown())
	}

	result.Ceiling = lastGood
	result.Converged = (lastBad - lastGood) <= cfg.AdditiveIncrease

	return result
}

func runAdaptiveStep(
	ctx context.Context,
	cfg AdaptiveConfig,
	deps StaircaseDeps,
	targetGames int,
	indexOffset int,
	stepIndex int,
) (StepOutput, bool) {
	collector := deps.NewCollector(cfg.HoldDuration)

	if cfg.WarmUpCompletions > 0 || cfg.WarmUpDurationSec > 0 {
		collector.ConfigureWarmUp(metrics.WarmUpConfig{
			MinCompletions: int64(cfg.WarmUpCompletions),
			MinDuration:    time.Duration(cfg.WarmUpDurationSec) * time.Second,
		})
	}

	if deps.OTelExporter != nil {
		collector.SetOTelExporter(deps.OTelExporter)
	}

	tracker := health.NewTracker(health.DefaultThresholds())
	observer := health.NewTrackerObserver(tracker)
	runFunc := deps.RunnerFactory(collector, observer)

	staircaseCfg := StaircaseConfig{
		Steps:        []int{targetGames},
		HoldDuration: cfg.HoldDuration,
		NumPlayers:   cfg.NumPlayers,
		GameTimeout:  cfg.GameTimeout,
		StaggerDelay: cfg.StaggerDelay,
		SLOs:         cfg.SLOs,
	}

	output := runStep(
		ctx,
		staircaseCfg,
		deps,
		targetGames,
		indexOffset,
		stepIndex,
		runFunc,
		collector,
		tracker,
	)

	metricsSnap := baseline.SnapshotToMetrics(output.Snapshot, output.Duration.Seconds())
	evalResult := cfg.SLOs.Evaluate(metricsSnap)

	return output, evalResult.AllPassing()
}

func cooldown(ctx context.Context, d time.Duration) {
	select {
	case <-time.After(d):
	case <-ctx.Done():
	}
}
