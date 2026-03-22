package orchestrator

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/metrics"
)

// RampConfig holds parameters for continuous ramp mode.
type RampConfig struct {
	GamesPerMinute int
	MaxGames       int
	ErrorThreshold float64
	GameTimeout    time.Duration
	NumPlayers     int
	Multiplier     float64 // Rate multiplier per minute (e.g., 2.0 = double each minute). 0 means constant.
}

// RunContinuousRamp spawns games at a steady rate, stopping when MaxGames is reached
// or error rate exceeds ErrorThreshold.
func RunContinuousRamp(
	cfg RampConfig,
	runner *GameRunner,
	collector *metrics.Collector,
) []GameResult {
	ctx, cancel := context.WithCancel(context.Background())
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
			"[ramp] starting: %d games/min ×%.1f/min, max %d, error threshold %.0f%%",
			cfg.GamesPerMinute, cfg.Multiplier, cfg.MaxGames, cfg.ErrorThreshold*100,
		)
	} else {
		log.Printf("[ramp] starting: %d games/min, max %d, error threshold %.0f%%",
			cfg.GamesPerMinute, cfg.MaxGames, cfg.ErrorThreshold*100)
	}

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

				break
			}
		}

		wg.Add(1)

		idx := launched

		go func() {
			defer wg.Done()

			result := runner.Run(ctx, idx, cfg.NumPlayers)

			mu.Lock()
			results = append(results, result)
			mu.Unlock()
		}()

		launched++

		// Check if we need to escalate rate.
		if cfg.Multiplier > 0 && time.Since(minuteStart) >= time.Minute {
			minuteNumber++
			currentRate *= cfg.Multiplier
			interval = time.Minute / time.Duration(currentRate)

			log.Printf("[ramp] minute %d: escalating to %.0f games/min",
				minuteNumber, currentRate)

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

	return results
}
