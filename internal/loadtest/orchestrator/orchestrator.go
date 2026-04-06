package orchestrator

import (
	"context"
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

// Config holds orchestration parameters for running concurrent games.
type Config struct {
	NumGames    int
	NumPlayers  int
	RampUp      time.Duration
	GameTimeout time.Duration
}

// Run executes NumGames concurrently with ramp-up, collecting metrics.
// It returns results for all games and handles graceful shutdown on SIGINT/SIGTERM.
//
//nolint:cyclop,funlen // batch orchestration with signal handling
func Run(
	ctx context.Context,
	cfg Config,
	runGame RunFunc,
	collector *metrics.StepAccumulator,
	annotator *annotations.Annotator,
) []GameResult {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Graceful shutdown on signal.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() { //nolint:gosec // G118: fire-and-forget signal handler
		<-sigCh
		observe.Info(
			context.Background(),
			"shutdown signal received, waiting for games to finish",
		)
		cancel()
	}()

	results := make([]GameResult, cfg.NumGames)

	var wg sync.WaitGroup

	// Calculate ramp interval.
	var rampInterval time.Duration
	if cfg.NumGames > 1 && cfg.RampUp > 0 {
		rampInterval = cfg.RampUp / time.Duration(cfg.NumGames-1)
	}

	// Launch games with ramp-up.
	annotator.Annotate("batch: started", "perf-test", "phase")

	for i := range cfg.NumGames {
		// Check if cancelled before launching.
		select {
		case <-ctx.Done():
			observe.Info(ctx, "cancelled before launching game",
				attribute.Int("game_index", i),
			)

			return results[:i]
		default:
		}

		if i > 0 && rampInterval > 0 {
			time.Sleep(rampInterval)
		}

		wg.Add(1)

		go func(idx int) {
			defer wg.Done()

			results[idx] = runGame(ctx, idx, cfg.NumPlayers)
		}(i)

		observe.Info(ctx, "launched game",
			attribute.Int("game", i+1),
			attribute.Int("total", cfg.NumGames),
		)
	}

	// Progress reporting in background.
	progressDone := make(chan struct{})

	go func() { //nolint:gosec // G118: intentional background goroutine
		defer close(progressDone)

		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				snap := collector.Snapshot()
				active := int64(
					cfg.NumGames,
				) - snap.GamesCompleted - snap.GamesTimedOut -
					snap.GamesFatal - snap.GamesCancelled
				observe.Info(context.Background(), "batch progress",
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

	wg.Wait()
	cancel() // Stop progress reporter.
	<-progressDone

	annotator.Annotate("batch: complete", "perf-test", "phase")

	return results
}
