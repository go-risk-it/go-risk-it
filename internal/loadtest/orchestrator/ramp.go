package orchestrator

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/annotations"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/metrics"
	"go.opentelemetry.io/otel/attribute"
)

// RampConfig holds parameters for continuous ramp mode.
type RampConfig struct {
	GamesPerMinute int
	MaxGames       int
	ErrorThreshold float64
	GameTimeout    time.Duration
	NumPlayers     int
	// Multiplier per step (e.g., 1.5 = 50% increase). 0 means constant.
	Multiplier   float64
	StepInterval time.Duration // How often to apply the multiplier. 0 defaults to 1 minute.
}

// stepInterval returns the effective step interval, defaulting to 1 minute.
func (c RampConfig) stepInterval() time.Duration {
	if c.StepInterval > 0 {
		return c.StepInterval
	}

	return time.Minute
}

// RunContinuousRamp spawns games at a steady rate, stopping when MaxGames is reached
// or error rate exceeds ErrorThreshold.
//
//nolint:cyclop,funlen // ramp with rate escalation and error threshold
func RunContinuousRamp(
	ctx context.Context,
	cfg RampConfig,
	runGame RunFunc,
	collector *metrics.StepAccumulator,
	annotator *annotations.Annotator,
) []GameResult {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() { //nolint:gosec // G118: fire-and-forget signal handler
		<-sigCh
		observe.Info(context.Background(), "shutdown signal received, stopping ramp")
		cancel()
	}()

	var (
		mu      sync.Mutex
		results []GameResult
		wg      sync.WaitGroup
	)

	currentRate := float64(cfg.GamesPerMinute)
	interval := time.Minute / time.Duration(currentRate)
	launched := 0
	minuteStart := time.Now()
	minuteNumber := 1

	if cfg.Multiplier > 0 {
		observe.Info(ctx, "ramp starting",
			attribute.Int("games_per_min", cfg.GamesPerMinute),
			attribute.Float64("multiplier", cfg.Multiplier),
			attribute.String("step_interval", cfg.stepInterval().String()),
			attribute.Int("max_games", cfg.MaxGames),
			attribute.Float64("error_threshold_pct", cfg.ErrorThreshold*100),
		)
	} else {
		observe.Info(ctx, "ramp starting",
			attribute.Int("games_per_min", cfg.GamesPerMinute),
			attribute.Int("max_games", cfg.MaxGames),
			attribute.Float64("error_threshold_pct", cfg.ErrorThreshold*100),
		)
	}

	annotator.Annotate("ramp: started", "perf-test", "phase")

	// Progress reporter.
	progressDone := make(chan struct{})

	go func() { //nolint:gosec // G118: background progress reporter
		defer close(progressDone)

		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				snap := collector.Snapshot()
				total := snap.GamesCompleted + snap.GamesTimedOut +
					snap.GamesFatal + snap.GamesCancelled
				active := int64(launched) - total

				observe.Info(context.Background(), "ramp progress",
					attribute.Int("launched", launched),
					attribute.Int64("active", active),
					attribute.Int64("completed", snap.GamesCompleted),
					attribute.Int64("fatal", snap.GamesFatal),
					attribute.Int64("moves", snap.TotalMoves),
					attribute.Int64("errors", snap.TotalErrors),
				)
			case <-ctx.Done():
				return
			}
		}
	}()

	for launched < cfg.MaxGames {
		select {
		case <-ctx.Done():
			observe.Info(ctx, "ramp cancelled", attribute.Int("games_launched", launched))

			goto wait
		default:
		}

		// Check error threshold.
		snap := collector.Snapshot()
		totalFinished := snap.GamesCompleted + snap.GamesTimedOut +
			snap.GamesFatal + snap.GamesCancelled

		if totalFinished > 0 {
			errorRate := float64(snap.GamesFatal+snap.GamesTimedOut) / float64(totalFinished)
			if errorRate > cfg.ErrorThreshold {
				observe.Warn(ctx, "ramp error threshold exceeded",
					attribute.Float64("error_rate_pct", errorRate*100),
					attribute.Float64("threshold_pct", cfg.ErrorThreshold*100),
					attribute.Int("games_launched", launched),
				)

				annotator.Annotate(
					"ramp: error threshold exceeded",
					"perf-test",
					"alert",
				)

				break
			}
		}

		wg.Add(1)

		idx := launched

		go func() {
			defer wg.Done()

			result := runGame(ctx, idx, cfg.NumPlayers)

			mu.Lock()
			results = append(results, result)
			mu.Unlock()
		}()

		launched++

		// Check if we need to escalate rate.
		if cfg.Multiplier > 0 && time.Since(minuteStart) >= cfg.stepInterval() {
			minuteNumber++
			currentRate *= cfg.Multiplier
			interval = time.Minute / time.Duration(currentRate)

			observe.Info(ctx, "ramp escalating",
				attribute.Int("step", minuteNumber),
				attribute.Float64("games_per_min", currentRate),
			)

			annotator.Annotate(
				fmt.Sprintf("ramp: %.0f games/min", currentRate),
				"perf-test",
				"rate",
			)

			minuteStart = time.Now()
		}

		if launched < cfg.MaxGames {
			time.Sleep(interval)
		}
	}

wait:
	observe.Info(ctx, "ramp all games launched, waiting for completion",
		attribute.Int("games_launched", launched),
	)
	wg.Wait()
	cancel()
	<-progressDone

	annotator.Annotate("ramp: complete", "perf-test", "phase")

	return results
}
