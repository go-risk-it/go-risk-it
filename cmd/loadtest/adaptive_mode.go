package main

import (
	"context"
	"os"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/baseline"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/journal"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/journal/session"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/orchestrator"
	"go.opentelemetry.io/otel/attribute"
)

//nolint:funlen // sequential adaptive orchestration
func (a *App) runAdaptive(ctx context.Context) error {
	// Detect branch and read session's last ceiling.
	_, branch := getGitInfo()

	initialCeiling := 0
	if branch != "" && branch != "main" && branch != "master" {
		store := session.NewStore("perf-journal/sessions")

		sess, err := store.Load(branch)
		if err == nil && sess.LastCeiling() > 0 {
			initialCeiling = sess.LastCeiling()
			observe.Info(ctx, "using session ceiling as starting point",
				attribute.Int("ceiling", initialCeiling),
			)
		}
	}

	// Build adaptive config.
	adaptiveCfg := orchestrator.AdaptiveConfig{
		InitialCeiling:   initialCeiling,
		AdditiveIncrease: a.cfg.Run.Adaptive.Increase,
		HoldDuration:     a.cfg.Run.Staircase.HoldDuration,
		NumPlayers:       orchestrator.DefaultNumPlayers,
		GameTimeout:      orchestrator.DefaultGameTimeout,
		StaggerDelay:     orchestrator.DefaultStaggerDelay,
		SLOs:             baseline.DefaultSLOs(),
		MaxSteps:         a.cfg.Run.Adaptive.MaxSteps,
		MaxGames:         a.cfg.Run.Adaptive.MaxGames,
	}

	// Build step executor.
	execCfg := orchestrator.StepExecutorConfig{
		NumPlayers:   adaptiveCfg.NumPlayers,
		GameTimeout:  adaptiveCfg.GameTimeout,
		StaggerDelay: adaptiveCfg.StaggerDelay,
		HoldDuration: adaptiveCfg.HoldDuration,
	}

	execDeps := a.buildStepExecutorDeps(adaptiveCfg.GameTimeout)
	executor := orchestrator.NewStepExecutor(execCfg, execDeps, adaptiveCfg.MaxSteps)

	// Run adaptive staircase.
	result := orchestrator.RunAdaptive(ctx, adaptiveCfg, executor)

	// Convert and analyze results.
	stepResults, _ := convertStepResults(result.Steps, adaptiveCfg.SLOs)
	ceiling := journal.FindSLOCeiling(stepResults)

	// Override ceiling games from the adaptive result (it knows the converged value).
	if result.Ceiling > 0 && ceiling.Games == 0 {
		ceiling.Games = result.Ceiling
	}

	insights := findCeilingInsights(stepResults, ceiling)
	commitSHA, _ := getGitInfo()

	// Collect steps list for config.
	configSteps := make([]int, len(result.Steps))
	for i, so := range result.Steps {
		configSteps[i] = so.TargetGames
	}

	// Build journal entry.
	entry := journal.Entry{
		CommitSHA: commitSHA,
		Timestamp: time.Now(),
		Branch:    branch,
		Config: journal.StaircaseParams{
			Mode:            "adaptive",
			Steps:           configSteps,
			HoldDurationSec: a.cfg.Run.Staircase.HoldDuration.Seconds(),
			NumPlayers:      adaptiveCfg.NumPlayers,
			GameTimeoutSec:  adaptiveCfg.GameTimeout.Seconds(),
		},
		SLOCeiling:  ceiling,
		Steps:       stepResults,
		Environment: baseline.CaptureEnvironment(),
		Insights:    insights,
	}

	// Print report.
	journal.PrintStaircaseReport(os.Stdout, entry)

	if len(insights) > 0 {
		baseline.PrintInsights(os.Stdout, insights)
	}

	observe.Info(ctx, "adaptive ceiling",
		attribute.Int("games", ceiling.Games),
		attribute.Bool("converged", result.Converged),
	)

	// Save and compare.
	a.handleJournalSaveAndCompare(entry, branch, commitSHA)

	return nil
}
