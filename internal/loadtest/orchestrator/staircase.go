package orchestrator

import (
	"context"
	"fmt"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/annotations"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/baseline"
	"go.opentelemetry.io/otel/attribute"
)

// StaircaseConfig defines the staircase run parameters.
// RunStaircase uses Steps and CooldownSec. The remaining fields are consumed
// by StepExecutorConfig and SLOStopCondition at construction time in the CLI.
type StaircaseConfig struct {
	Steps        []int         // target concurrent games per step
	HoldDuration time.Duration // how long to hold each step
	NumPlayers   int
	GameTimeout  time.Duration
	StopOnBreach bool          // stop when SLOs fail
	StaggerDelay time.Duration // between initial game launches within a step
	CooldownSec  int           // seconds between steps (default 5)
	SLOs         baseline.SLOSet
}

// RunStaircase executes the staircase: for each step, the executor runs games
// at the target concurrency. The stopper decides whether to stop after each step.
// Returns one StepOutput per executed step (including any breach step).
//
//nolint:funlen // sequential step orchestration
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
			observe.Info(ctx, "staircase cancelled",
				attribute.Int("step", i+1),
				attribute.Int("total_steps", len(steps)),
			)

			return outputs
		default:
		}

		observe.Info(ctx, "staircase step starting",
			attribute.Int("step", i+1),
			attribute.Int("total_steps", len(steps)),
			attribute.Int("targetGames", targetGames),
		)

		output, err := executor.Execute(ctx, targetGames, indexOffset)
		if err != nil {
			observe.Error(ctx, err, "staircase step execute failed",
				attribute.Int("step", i+1),
				attribute.Int("total_steps", len(steps)),
			)

			break
		}

		outputs = append(outputs, *output)
		indexOffset += targetGames * IndexOffsetMultiplier

		if stopper.ShouldStop(output) {
			observe.Warn(ctx, "staircase SLO breach, stopping",
				attribute.Int("targetGames", targetGames),
			)
			annotator.Annotate(
				fmt.Sprintf("staircase: SLO breach at %d games", targetGames),
				"perf-test",
				"alert",
			)

			break
		}

		// Cooldown between steps.
		if i < len(steps)-1 {
			observe.Info(ctx, "staircase cooldown", attribute.String("duration", cd.String()))

			select {
			case <-ctx.Done():
				return outputs
			case <-time.After(cd):
			}
		}
	}

	return outputs
}
