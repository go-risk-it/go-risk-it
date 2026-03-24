package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/baseline"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/metrics"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/orchestrator"
)

func (a *App) runBatch(ctx context.Context) error {
	batchCfg := orchestrator.Config{
		NumGames:    a.cfg.Run.Batch.NumGames,
		NumPlayers:  a.cfg.Game.NumPlayers,
		RampUp:      a.cfg.Run.Batch.RampUp,
		GameTimeout: a.cfg.Game.GameTimeout,
	}

	collector := a.newCollector(a.cfg.Game.GameTimeout)
	runGame := a.newRunFunc(collector, a.cfg.Game.GameTimeout)

	start := time.Now()
	results := orchestrator.Run(ctx, batchCfg, runGame, collector, a.annotator)
	totalDuration := time.Since(start)

	return a.printBatchReport(collector, results, totalDuration)
}

func (a *App) runRamp(ctx context.Context) error {
	rampCfg := a.cfg.Run.rampCfg
	if rampCfg == nil {
		rampCfg = &orchestrator.RampConfig{
			GamesPerMinute: a.cfg.Run.Ramp.Rate,
			MaxGames:       a.cfg.Run.Ramp.MaxGames,
			ErrorThreshold: a.cfg.Run.Ramp.ErrorThreshold,
			GameTimeout:    a.cfg.Game.GameTimeout,
			NumPlayers:     a.cfg.Game.NumPlayers,
			Multiplier:     a.cfg.Run.Ramp.Multiplier,
		}
	}

	estimated := estimateRampDuration(rampCfg)
	maxDuration := estimated + rampCfg.GameTimeout
	collector := a.newCollector(maxDuration)
	runGame := a.newRunFunc(collector, rampCfg.GameTimeout)

	start := time.Now()
	results := orchestrator.RunContinuousRamp(ctx, *rampCfg, runGame, collector, a.annotator)
	totalDuration := time.Since(start)

	return a.printBatchReport(collector, results, totalDuration)
}

// printBatchReport prints the report for batch and ramp modes.
func (a *App) printBatchReport(
	collector *metrics.Collector,
	results []orchestrator.GameResult,
	totalDuration time.Duration,
) error {
	fatalErrors := 0
	reportResults := make([]metrics.GameResult, len(results))

	for i, r := range results {
		if r.FatalError != nil {
			fatalErrors++
		}

		reportResults[i] = metrics.GameResult{
			GameIndex:  r.GameIndex,
			Duration:   r.Duration,
			Moves:      r.Moves,
			Errors:     r.Errors,
			Winner:     r.Winner,
			TimedOut:   r.TimedOut,
			FatalError: r.FatalError,
		}
	}

	snap := collector.Snapshot()

	switch a.cfg.Report.Output.Format {
	case "json":
		if err := metrics.PrintJSON(os.Stdout, snap, totalDuration, fatalErrors, reportResults); err != nil {
			return fmt.Errorf("json report: %w", err)
		}
	case "text":
		metrics.PrintReport(os.Stdout, snap, totalDuration, fatalErrors, reportResults)
	default:
		return fmt.Errorf("unknown output format: %q", a.cfg.Report.Output.Format)
	}

	metricsSnap := baseline.SnapshotToMetrics(snap, totalDuration.Seconds())
	insights := baseline.Analyze(metricsSnap)
	baseline.PrintInsights(os.Stdout, insights)

	if err := handleBaseline(a.cfg, metricsSnap, insights); err != nil {
		log.Printf("baseline: %v", err)
	}

	if fatalErrors > 0 {
		return fmt.Errorf("%d game(s) had fatal errors", fatalErrors)
	}

	return nil
}
