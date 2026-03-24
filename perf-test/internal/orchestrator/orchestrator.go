package orchestrator

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/annotations"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/metrics"
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
func Run(
	cfg Config,
	runGame RunFunc,
	collector *metrics.Collector,
	annotator *annotations.Annotator,
) []GameResult {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Graceful shutdown on signal.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Println("shutdown signal received, waiting for running games to finish...")
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

	for i := 0; i < cfg.NumGames; i++ {
		// Check if cancelled before launching.
		select {
		case <-ctx.Done():
			log.Printf("cancelled before launching game %d", i)

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

		log.Printf("launched game %d/%d", i+1, cfg.NumGames)
	}

	// Progress reporting in background.
	progressDone := make(chan struct{})

	go func() {
		defer close(progressDone)

		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				snap := collector.Snapshot()
				active := int64(
					cfg.NumGames,
				) - snap.GamesCompleted - snap.GamesTimedOut - snap.GamesFatal
				log.Printf(
					"[progress] active=%d completed=%d fatal=%d moves=%d errors=%d",
					active,
					snap.GamesCompleted,
					snap.GamesFatal,
					snap.TotalMoves,
					snap.TotalErrors,
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
