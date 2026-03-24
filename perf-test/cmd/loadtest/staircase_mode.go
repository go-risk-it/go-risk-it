package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/baseline"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/chaos"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/journal"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/metrics"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/orchestrator"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/resources"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/runner"
)

//nolint:funlen,cyclop
func (a *App) runStaircase(ctx context.Context) error {
	staircaseCfg := a.cfg.staircaseCfg

	// Build staircase config from flags if not set by preset.
	if staircaseCfg == nil {
		steps := parseSteps(a.cfg.Staircase.Steps)
		if len(steps) == 0 {
			log.Fatal("staircase mode requires --preset or --steps flag")
		}

		staircaseCfg = &orchestrator.StaircaseConfig{
			Steps:             steps,
			HoldDuration:      a.cfg.Staircase.HoldDuration,
			NumPlayers:        orchestrator.DefaultNumPlayers,
			GameTimeout:       orchestrator.DefaultGameTimeout,
			StopOnBreach:      a.cfg.Staircase.StopOnBreach,
			StaggerDelay:      orchestrator.DefaultStaggerDelay,
			SLOs:              baseline.DefaultSLOs(),
			WarmUpCompletions: a.cfg.Staircase.WarmupCompletions,
			WarmUpDurationSec: a.cfg.Staircase.WarmupDuration,
		}
	}

	// CLI warm-up flags override preset values.
	if a.cfg.Staircase.WarmupCompletions > 0 {
		staircaseCfg.WarmUpCompletions = a.cfg.Staircase.WarmupCompletions
	}

	if a.cfg.Staircase.WarmupDuration > 0 {
		staircaseCfg.WarmUpDurationSec = a.cfg.Staircase.WarmupDuration
	}

	// Build step executor.
	execCfg := orchestrator.StepExecutorConfig{
		NumPlayers:        staircaseCfg.NumPlayers,
		GameTimeout:       staircaseCfg.GameTimeout,
		StaggerDelay:      staircaseCfg.StaggerDelay,
		HoldDuration:      staircaseCfg.HoldDuration,
		WarmUpCompletions: staircaseCfg.WarmUpCompletions,
		WarmUpDurationSec: staircaseCfg.WarmUpDurationSec,
	}

	execDeps := a.buildStepExecutorDeps(staircaseCfg.GameTimeout)
	executor := orchestrator.NewStepExecutor(execCfg, execDeps, len(staircaseCfg.Steps))

	// Build stop condition.
	var stopper orchestrator.StopCondition
	if staircaseCfg.StopOnBreach {
		stopper = &orchestrator.SLOStopCondition{SLOs: staircaseCfg.SLOs}
	} else {
		stopper = &orchestrator.NeverStop{}
	}

	// Run staircase.
	stepOutputs := orchestrator.RunStaircase(
		ctx, staircaseCfg.Steps, staircaseCfg.CooldownSec,
		executor, stopper, a.annotator,
	)

	// Convert and analyze results.
	stepResults, levelResults := convertStepResults(stepOutputs, staircaseCfg.SLOs)
	ceiling := journal.FindSLOCeiling(stepResults)
	breakingPoints := baseline.FindBreakingPoints(levelResults, staircaseCfg.SLOs)
	insights := findCeilingInsights(stepResults, ceiling)
	commitSHA, branch := getGitInfo()

	// Build journal entry.
	entry := journal.Entry{
		CommitSHA: commitSHA,
		Timestamp: time.Now(),
		Branch:    branch,
		Config: journal.StaircaseParams{
			Steps:             staircaseCfg.Steps,
			HoldDurationSec:   staircaseCfg.HoldDuration.Seconds(),
			NumPlayers:        staircaseCfg.NumPlayers,
			GameTimeoutSec:    staircaseCfg.GameTimeout.Seconds(),
			StopOnBreach:      staircaseCfg.StopOnBreach,
			WarmUpCompletions: staircaseCfg.WarmUpCompletions,
			WarmUpDurationSec: staircaseCfg.WarmUpDurationSec,
		},
		SLOCeiling:     ceiling,
		Steps:          stepResults,
		BreakingPoints: breakingPoints,
		Environment:    baseline.CaptureEnvironment(),
		Insights:       insights,
	}

	// Print report.
	journal.PrintStaircaseReport(os.Stdout, entry)

	if len(insights) > 0 {
		baseline.PrintInsights(os.Stdout, insights)
	}

	log.Printf("[staircase] ceiling=%d games", ceiling.Games)

	// Save and compare.
	a.handleJournalSaveAndCompare(entry, branch, commitSHA)

	return nil
}

// buildStepExecutorDeps creates the step executor dependencies.
func (a *App) buildStepExecutorDeps(gameTimeout time.Duration) orchestrator.StepExecutorDeps {
	return orchestrator.StepExecutorDeps{
		RunnerFactory: func(c *metrics.Collector, obs orchestrator.GameObserver) orchestrator.RunFunc {
			strat := a.strategy
			if a.cfg.Chaos.Enabled() {
				strat = chaos.WrapStrategy(strat, a.cfg.Chaos, c)
			}

			var injector *chaos.Injector
			if a.cfg.Chaos.DisconnectRate > 0 {
				injector = chaos.NewInjector(a.cfg.Chaos, c)
			}

			return runner.New(runner.Config{
				BaseURL:       a.cfg.Server.URL,
				WSURL:         a.cfg.Server.WSURL,
				AnonKey:       a.cfg.Server.AnonKey,
				Strategy:      strat,
				Timeout:       gameTimeout,
				Collector:     c,
				ThinkTime:     a.cfg.Game.ThinkTime,
				Timeouts:      runner.DefaultTimeouts(),
				ChaosInjector: injector,
				Observer:      obs,
			}).ToRunFunc()
		},
		NewCollector: metrics.NewCollector,
		CollectResources: func() resources.ServerResources {
			return resources.CollectServerResources(resources.DefaultStatsFunc)
		},
		Annotator:    a.annotator,
		OTelExporter: a.otel,
		DBStats:      a.dbStats,
	}
}

// buildStaircaseDeps creates StaircaseDeps for adaptive mode.
// Deprecated: use buildStepExecutorDeps + DefaultStepExecutor instead.
func (a *App) buildStaircaseDeps(gameTimeout time.Duration) orchestrator.StaircaseDeps {
	ed := a.buildStepExecutorDeps(gameTimeout)

	return orchestrator.StaircaseDeps{
		RunnerFactory:    ed.RunnerFactory,
		NewCollector:     ed.NewCollector,
		CollectResources: ed.CollectResources,
		Annotator:        ed.Annotator,
		OTelExporter:     ed.OTelExporter,
		DBStats:          ed.DBStats,
	}
}

// handleJournalSaveAndCompare saves and compares journal entries based on config.
func (a *App) handleJournalSaveAndCompare(entry journal.Entry, branch, commitSHA string) {
	if a.cfg.Journal.Save {
		slug := a.cfg.Journal.Name
		if slug == "" {
			slug = a.cfg.Preset
		}

		if slug == "" {
			slug = a.cfg.Mode
		}

		saveJournalEntry(entry, slug, branch, commitSHA, a.cfg.Journal.Hypothesis)
	}

	if a.cfg.Journal.Compare != "" {
		compareJournalEntries(a.cfg.Journal.Compare, entry)
	}
}
