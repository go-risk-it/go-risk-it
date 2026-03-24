package orchestrator

import (
	"context"
	"log"
	"time"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/baseline"
)

// AdaptiveConfig defines the TCP congestion control staircase parameters.
type AdaptiveConfig struct {
	InitialCeiling   int // from session's last ceiling; 0 starts at 10
	AdditiveIncrease int // games to add per successful step (default 5)
	SLOs             baseline.SLOSet
	MaxSteps         int // upper bound on steps (default 20)
	MaxGames         int // hard ceiling (default 500)
	CooldownSec      int

	// Fields below are consumed by StepExecutorConfig at construction time.
	// RunAdaptive does not use them directly.
	HoldDuration      time.Duration
	NumPlayers        int
	GameTimeout       time.Duration
	StaggerDelay      time.Duration
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
	executor StepExecutor,
) AdaptiveResult {
	if cfg.AdditiveIncrease <= 0 {
		cfg.AdditiveIncrease = DefaultAdaptiveIncrease
	}

	if cfg.MaxSteps <= 0 {
		cfg.MaxSteps = DefaultAdaptiveMaxSteps
	}

	if cfg.MaxGames <= 0 {
		cfg.MaxGames = DefaultAdaptiveMaxGames
	}

	initial := cfg.InitialCeiling
	if initial <= 0 {
		initial = DefaultAdaptiveInitialProbe
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

		output, passed := executeAndEvaluate(ctx, executor, cfg.SLOs, current, indexOffset)
		result.Steps = append(result.Steps, output)
		indexOffset += current * IndexOffsetMultiplier
		stepNum++

		if !passed {
			log.Printf("[adaptive] breach at %d games, switching to binary search", current)

			// Phase 2: Binary search.
			lastBad := current
			result = binarySearch(
				ctx, cfg, executor, lastGood, lastBad,
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
	executor StepExecutor,
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

		output, passed := executeAndEvaluate(ctx, executor, cfg.SLOs, mid, indexOffset)
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

// executeAndEvaluate runs a step and evaluates SLOs, returning the output and
// whether all SLOs passed.
func executeAndEvaluate(
	ctx context.Context,
	executor StepExecutor,
	slos baseline.SLOSet,
	targetGames, indexOffset int,
) (StepOutput, bool) {
	output, err := executor.Execute(ctx, targetGames, indexOffset)
	if err != nil {
		log.Printf("[adaptive] execute: %v", err)

		return StepOutput{TargetGames: targetGames}, false
	}

	if output.Snapshot == nil {
		return *output, true
	}

	metricsSnap := baseline.SnapshotToMetrics(output.Snapshot, output.Duration.Seconds())
	evalResult := slos.Evaluate(metricsSnap)

	return *output, evalResult.AllPassing()
}

func cooldown(ctx context.Context, d time.Duration) {
	select {
	case <-time.After(d):
	case <-ctx.Done():
	}
}
