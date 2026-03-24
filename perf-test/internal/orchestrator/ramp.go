package orchestrator

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/annotations"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/metrics"
)

// RampConfig holds parameters for continuous ramp mode.
type RampConfig struct {
	GamesPerMinute int
	MaxGames       int
	ErrorThreshold float64
	GameTimeout    time.Duration
	NumPlayers     int
	Multiplier     float64       // Rate multiplier per step (e.g., 1.5 = 50% increase each step). 0 means constant.
	StepInterval   time.Duration // How often to apply the multiplier. 0 defaults to 1 minute.
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
func RunContinuousRamp(
	ctx context.Context,
	cfg RampConfig,
	runGame RunFunc,
	collector *metrics.Collector,
	annotator *annotations.Annotator,
) []GameResult {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Println("shutdown signal received, stopping ramp...")
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
		log.Printf(
			"[ramp] starting: %d games/min ×%.1f/%v, max %d, error threshold %.0f%%",
			cfg.GamesPerMinute,
			cfg.Multiplier,
			cfg.stepInterval(),
			cfg.MaxGames,
			cfg.ErrorThreshold*100,
		)
	} else {
		log.Printf("[ramp] starting: %d games/min, max %d, error threshold %.0f%%",
			cfg.GamesPerMinute, cfg.MaxGames, cfg.ErrorThreshold*100)
	}

	annotator.Annotate("ramp: started", "perf-test", "phase")

	// Progress reporter.
	progressDone := make(chan struct{})

	go func() {
		defer close(progressDone)

		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				snap := collector.Snapshot()
				total := snap.GamesCompleted + snap.GamesTimedOut + snap.GamesFatal
				active := int64(launched) - total

				log.Printf("[ramp] launched=%d active=%d completed=%d fatal=%d moves=%d errors=%d",
					launched, active, snap.GamesCompleted, snap.GamesFatal,
					snap.TotalMoves, snap.TotalErrors)
			case <-ctx.Done():
				return
			}
		}
	}()

	for launched < cfg.MaxGames {
		select {
		case <-ctx.Done():
			log.Printf("[ramp] cancelled after %d games", launched)

			goto wait
		default:
		}

		// Check error threshold.
		snap := collector.Snapshot()
		totalFinished := snap.GamesCompleted + snap.GamesTimedOut + snap.GamesFatal

		if totalFinished > 0 {
			errorRate := float64(snap.GamesFatal+snap.GamesTimedOut) / float64(totalFinished)
			if errorRate > cfg.ErrorThreshold {
				log.Printf(
					"[ramp] error threshold exceeded: %.1f%% > %.1f%%, stopping after %d games",
					errorRate*100,
					cfg.ErrorThreshold*100,
					launched,
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

			log.Printf("[ramp] step %d: escalating to %.0f games/min",
				minuteNumber, currentRate)

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
	log.Printf("[ramp] all %d games launched, waiting for completion...", launched)
	wg.Wait()
	cancel()
	<-progressDone

	annotator.Annotate("ramp: complete", "perf-test", "phase")

	return results
}
