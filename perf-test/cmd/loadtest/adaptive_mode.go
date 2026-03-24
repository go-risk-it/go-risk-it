package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/baseline"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/journal"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/journal/session"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/orchestrator"
)

//nolint:funlen,cyclop
func (a *App) runAdaptive(ctx context.Context) error {
	// Detect branch and read session's last ceiling.
	_, branch := getGitInfo()

	initialCeiling := 0
	if branch != "" && branch != "main" && branch != "master" {
		store := session.NewStore("perf-journal/sessions")

		sess, err := store.Load(branch)
		if err == nil && sess.LastCeiling() > 0 {
			initialCeiling = sess.LastCeiling()
			log.Printf("[adaptive] using session ceiling %d as starting point", initialCeiling)
		}
	}

	// Build adaptive config.
	adaptiveCfg := orchestrator.AdaptiveConfig{
		InitialCeiling:    initialCeiling,
		AdditiveIncrease:  a.cfg.Adaptive.Increase,
		HoldDuration:      a.cfg.Staircase.HoldDuration,
		NumPlayers:        orchestrator.DefaultNumPlayers,
		GameTimeout:       orchestrator.DefaultGameTimeout,
		StaggerDelay:      orchestrator.DefaultStaggerDelay,
		SLOs:              baseline.DefaultSLOs(),
		MaxSteps:          a.cfg.Adaptive.MaxSteps,
		MaxGames:          a.cfg.Adaptive.MaxGames,
		WarmUpCompletions: a.cfg.Staircase.WarmupCompletions,
		WarmUpDurationSec: a.cfg.Staircase.WarmupDuration,
	}

	// Build dependencies (same as staircase).
	deps := a.buildStaircaseDeps(adaptiveCfg.GameTimeout)

	// Run adaptive staircase.
	result := orchestrator.RunAdaptive(ctx, adaptiveCfg, deps)

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
			Mode:              "adaptive",
			Steps:             configSteps,
			HoldDurationSec:   a.cfg.Staircase.HoldDuration.Seconds(),
			NumPlayers:        adaptiveCfg.NumPlayers,
			GameTimeoutSec:    adaptiveCfg.GameTimeout.Seconds(),
			WarmUpCompletions: a.cfg.Staircase.WarmupCompletions,
			WarmUpDurationSec: a.cfg.Staircase.WarmupDuration,
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

	convergedStr := ""
	if result.Converged {
		convergedStr = " (converged)"
	}

	log.Printf("[adaptive] ceiling=%d games%s", ceiling.Games, convergedStr)

	// Save and compare.
	a.handleJournalSaveAndCompare(entry, branch, commitSHA)

	return nil
}
