package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/baseline"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/chaos"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/metrics"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/orchestrator"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/runner"
)

func (a *App) runBatch(ctx context.Context) error {
	_ = ctx // batch mode doesn't use context cancellation yet

	batchCfg := orchestrator.Config{
		NumGames:    a.cfg.NumGames,
		NumPlayers:  a.cfg.Game.NumPlayers,
		RampUp:      a.cfg.RampUp,
		GameTimeout: a.cfg.Game.GameTimeout,
	}

	maxDuration := a.cfg.Game.GameTimeout
	collector := metrics.NewCollector(maxDuration)

	if a.otel != nil {
		collector.SetOTelExporter(a.otel)
	}

	strategy := a.strategy
	if a.cfg.Chaos.Enabled() {
		strategy = chaos.WrapStrategy(strategy, a.cfg.Chaos, collector)
	}

	var injector *chaos.Injector
	if a.cfg.Chaos.DisconnectRate > 0 {
		injector = chaos.NewInjector(a.cfg.Chaos, collector)
	}

	runGame := runner.New(runner.Config{
		BaseURL:       a.cfg.Server.URL,
		WSURL:         a.cfg.Server.WSURL,
		AnonKey:       a.cfg.Server.AnonKey,
		Strategy:      strategy,
		Timeout:       a.cfg.Game.GameTimeout,
		Collector:     collector,
		ThinkTime:     a.cfg.Game.ThinkTime,
		Timeouts:      runner.DefaultTimeouts(),
		ChaosInjector: injector,
	}).ToRunFunc()

	start := time.Now()
	results := orchestrator.Run(batchCfg, runGame, collector, a.annotator)
	totalDuration := time.Since(start)

	return a.printBatchReport(collector, results, totalDuration)
}

func (a *App) runRamp(ctx context.Context) error {
	_ = ctx // ramp mode doesn't use context cancellation yet

	rampCfg := a.cfg.rampCfg
	if rampCfg == nil {
		rampCfg = &orchestrator.RampConfig{
			GamesPerMinute: a.cfg.Ramp.Rate,
			MaxGames:       a.cfg.Ramp.MaxGames,
			ErrorThreshold: a.cfg.Ramp.ErrorThreshold,
			GameTimeout:    a.cfg.Game.GameTimeout,
			NumPlayers:     a.cfg.Game.NumPlayers,
			Multiplier:     a.cfg.Ramp.Multiplier,
		}
	}

	estimated := estimateRampDuration(rampCfg)
	maxDuration := estimated + rampCfg.GameTimeout
	collector := metrics.NewCollector(maxDuration)

	if a.otel != nil {
		collector.SetOTelExporter(a.otel)
	}

	strategy := a.strategy
	if a.cfg.Chaos.Enabled() {
		strategy = chaos.WrapStrategy(strategy, a.cfg.Chaos, collector)
	}

	var injector *chaos.Injector
	if a.cfg.Chaos.DisconnectRate > 0 {
		injector = chaos.NewInjector(a.cfg.Chaos, collector)
	}

	runGame := runner.New(runner.Config{
		BaseURL:       a.cfg.Server.URL,
		WSURL:         a.cfg.Server.WSURL,
		AnonKey:       a.cfg.Server.AnonKey,
		Strategy:      strategy,
		Timeout:       rampCfg.GameTimeout,
		Collector:     collector,
		ThinkTime:     a.cfg.Game.ThinkTime,
		Timeouts:      runner.DefaultTimeouts(),
		ChaosInjector: injector,
	}).ToRunFunc()

	start := time.Now()
	results := orchestrator.RunContinuousRamp(*rampCfg, runGame, collector, a.annotator)
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

	switch a.cfg.Output.Format {
	case "json":
		if err := metrics.PrintJSON(os.Stdout, snap, totalDuration, fatalErrors, reportResults); err != nil {
			log.Fatalf("json report: %v", err)
		}
	case "text":
		metrics.PrintReport(os.Stdout, snap, totalDuration, fatalErrors, reportResults)
	default:
		log.Fatalf("unknown output format: %q", a.cfg.Output.Format)
	}

	metricsSnap := baseline.SnapshotToMetrics(snap, totalDuration.Seconds())
	insights := baseline.Analyze(metricsSnap)
	baseline.PrintInsights(os.Stdout, insights)

	handleBaseline(a.cfg, metricsSnap, insights)

	if fatalErrors > 0 {
		return fmt.Errorf("%d game(s) had fatal errors", fatalErrors)
	}

	return nil
}
