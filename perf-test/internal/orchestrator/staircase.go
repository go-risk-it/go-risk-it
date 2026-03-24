package orchestrator

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/annotations"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/baseline"
)

// StaircaseConfig defines the staircase run parameters.
type StaircaseConfig struct {
	Steps             []int         // target concurrent games per step
	HoldDuration      time.Duration // how long to hold each step
	NumPlayers        int
	GameTimeout       time.Duration
	StopOnBreach      bool          // stop when SLOs fail
	StaggerDelay      time.Duration // between initial game launches within a step
	CooldownSec       int           // seconds between steps (default 5)
	WarmUpCompletions int           // games to complete before recording histograms (0 = disabled)
	WarmUpDurationSec int           // seconds to wait before recording histograms (0 = disabled)
	// SLOs is deprecated in RunStaircase (use SLOStopCondition instead) but
	// still consumed by adaptive.go until Task 9 refactors it.
	SLOs baseline.SLOSet
}

// cooldown returns the effective cooldown duration.
func (c StaircaseConfig) cooldown() time.Duration {
	if c.CooldownSec > 0 {
		return time.Duration(c.CooldownSec) * time.Second
	}

	return DefaultCooldownSec * time.Second
}

// RunStaircase executes the staircase: for each step, the executor runs games
// at the target concurrency. The stopper decides whether to stop after each step.
// Returns one StepOutput per executed step (including any breach step).
func RunStaircase(
	ctx context.Context,
	steps []int,
	cooldownSec int,
	executor StepExecutor,
	stopper StopCondition,
	annotator *annotations.Annotator,
) []StepOutput {
	var outputs []StepOutput

	indexOffset := 0

	cd := time.Duration(cooldownSec) * time.Second
	if cooldownSec <= 0 {
		cd = DefaultCooldownSec * time.Second
	}

	for i, targetGames := range steps {
		select {
		case <-ctx.Done():
			log.Printf("[staircase] cancelled at step %d/%d", i+1, len(steps))

			return outputs
		default:
		}

		log.Printf(
			"[staircase] step %d/%d: %d concurrent games",
			i+1, len(steps), targetGames,
		)

		output, err := executor.Execute(ctx, targetGames, indexOffset)
		if err != nil {
			log.Printf("[staircase] step %d/%d: execute: %v", i+1, len(steps), err)

			break
		}

		outputs = append(outputs, *output)
		indexOffset += targetGames * IndexOffsetMultiplier

		if stopper.ShouldStop(output) {
			log.Printf("[staircase] SLO breach at %d games, stopping", targetGames)
			annotator.Annotate(
				fmt.Sprintf("staircase: SLO breach at %d games", targetGames),
				"perf-test",
				"alert",
			)

			break
		}

		// Cooldown between steps.
		if i < len(steps)-1 {
			log.Printf("[staircase] cooldown %v before next step", cd)

			select {
			case <-ctx.Done():
				return outputs
			case <-time.After(cd):
			}
		}
	}

	return outputs
}
