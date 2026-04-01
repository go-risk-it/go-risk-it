package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/baseline"
	"go.opentelemetry.io/otel/attribute"
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

		ctx := context.Background()

		if err != nil {
			observe.Error(ctx, err, "failed to save baseline")
		} else {
			observe.Info(ctx, "baseline saved", attribute.String("path", path))
		}
	}

	if cfg.Report.Output.CompareFile != "" {
		referenceBaseline, err := baseline.Load(cfg.Report.Output.CompareFile)
		if err != nil {
			return fmt.Errorf("load baseline for comparison: %w", err)
		}

		os.Stdout.WriteString("\n")
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
