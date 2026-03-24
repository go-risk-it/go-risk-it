package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/baseline"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/journal"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/metrics"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/orchestrator"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/resources"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/runner"
)

//nolint:funlen // orchestration function with sequential setup steps
func (a *App) runStaircase(ctx context.Context) error {
	staircaseCfg := a.cfg.Run.staircaseCfg

	// Build staircase config from flags if not set by preset.
	if staircaseCfg == nil {
		steps, err := parseSteps(a.cfg.Run.Staircase.Steps)
		if err != nil {
			return fmt.Errorf("parse steps: %w", err)
		}

		if len(steps) == 0 {
			return fmt.Errorf("staircase mode requires --preset or --steps flag")
		}

		staircaseCfg = &orchestrator.StaircaseConfig{
			Steps:             steps,
			HoldDuration:      a.cfg.Run.Staircase.HoldDuration,
			NumPlayers:        orchestrator.DefaultNumPlayers,
			GameTimeout:       orchestrator.DefaultGameTimeout,
			StopOnBreach:      a.cfg.Run.Staircase.StopOnBreach,
			StaggerDelay:      orchestrator.DefaultStaggerDelay,
			SLOs:              baseline.DefaultSLOs(),
			WarmUpCompletions: a.cfg.Run.Staircase.WarmupCompletions,
			WarmUpDurationSec: a.cfg.Run.Staircase.WarmupDuration,
		}
	}

	// CLI warm-up flags override preset values.
	if a.cfg.Run.Staircase.WarmupCompletions > 0 {
		staircaseCfg.WarmUpCompletions = a.cfg.Run.Staircase.WarmupCompletions
	}

	if a.cfg.Run.Staircase.WarmupDuration > 0 {
		staircaseCfg.WarmUpDurationSec = a.cfg.Run.Staircase.WarmupDuration
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
			strategy, injector := a.setupChaos(c)

			return runner.New(runner.Config{
				BaseURL:       a.cfg.Server.URL,
				WSURL:         a.cfg.Server.WSURL,
				AnonKey:       a.cfg.Server.AnonKey,
				Strategy:      strategy,
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

// handleJournalSaveAndCompare saves and compares journal entries based on config.
func (a *App) handleJournalSaveAndCompare(entry journal.Entry, branch, commitSHA string) {
	if a.cfg.Report.Journal.Save {
		slug := a.cfg.Report.Journal.Name
		if slug == "" {
			slug = a.cfg.Run.Preset
		}

		if slug == "" {
			slug = a.cfg.Run.Mode
		}

		saveJournalEntry(entry, slug, branch, commitSHA, a.cfg.Report.Journal.Hypothesis)
	}

	if a.cfg.Report.Journal.Compare != "" {
		if err := compareJournalEntries(a.cfg.Report.Journal.Compare, entry); err != nil {
			log.Printf("journal compare: %v", err)
		}
	}
}
