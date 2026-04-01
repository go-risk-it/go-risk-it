package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/loadtest/baseline"
)

// handleBaseline saves and/or compares baselines.
func handleBaseline(
	cfg *Config,
	metricsSnap baseline.MetricsSnapshot,
	insights []baseline.Insight,
) error {
	if !cfg.Report.Output.SaveBaseline && cfg.Report.Output.CompareFile == "" {
		return nil
	}

	commitSHA, _ := getGitInfo()
	currentBaseline := buildCurrentBaseline(cfg, metricsSnap, insights, commitSHA)

	if cfg.Report.Output.SaveBaseline {
		var path string
		var err error

		if cfg.Report.Output.BaselineName != "" {
			path, err = baseline.SaveNumbered(
				"perf-journal/baselines",
				cfg.Report.Output.BaselineName,
				currentBaseline,
			)
		} else {
			path, err = baseline.Save("baselines", currentBaseline)
		}

		if err != nil {
			log.Printf("failed to save baseline: %v", err)
		} else {
			log.Printf("baseline saved: %s", path)
		}
	}

	if cfg.Report.Output.CompareFile != "" {
		referenceBaseline, err := baseline.Load(cfg.Report.Output.CompareFile)
		if err != nil {
			return fmt.Errorf("load baseline for comparison: %w", err)
		}

		fmt.Println()
		baseline.PrintComparison(os.Stdout, referenceBaseline, currentBaseline)
	}

	return nil
}

// buildCurrentBaseline creates a Baseline from the current run.
// commitSHA is passed in to avoid duplicate git calls.
func buildCurrentBaseline(
	cfg *Config,
	metricsSnap baseline.MetricsSnapshot,
	insights []baseline.Insight,
	commitSHA string,
) baseline.Baseline {
	return baseline.Baseline{
		CommitSHA: commitSHA,
		Timestamp: time.Now(),
		TestParams: baseline.TestParams{
			Preset:  cfg.Run.Preset,
			Players: cfg.Game.NumPlayers,
			Games:   cfg.Run.Batch.NumGames,
			Mode:    cfg.Run.Mode,
		},
		Metrics:     metricsSnap,
		Environment: baseline.CaptureEnvironment(),
		Insights:    insights,
	}
}
