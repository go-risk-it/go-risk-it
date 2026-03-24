package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/baseline"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/journal"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/journal/session"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/orchestrator"
)

// parseSteps parses a comma-separated list of step counts.
func parseSteps(s string) []int {
	if s == "" {
		return nil
	}

	parts := strings.Split(s, ",")
	steps := make([]int, 0, len(parts))

	for _, p := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil {
			log.Fatalf("invalid step count %q: %v", p, err)
		}

		steps = append(steps, n)
	}

	return steps
}

// estimateRampDuration estimates the total ramp duration for throughput bucket sizing.
func estimateRampDuration(rampCfg *orchestrator.RampConfig) time.Duration {
	if rampCfg.Multiplier <= 0 {
		return time.Duration(
			rampCfg.MaxGames/max(rampCfg.GamesPerMinute, 1),
		) * time.Minute
	}

	rate := float64(rampCfg.GamesPerMinute)
	total := 0
	steps := 0

	step := rampCfg.StepInterval
	if step <= 0 {
		step = time.Minute
	}

	for total < rampCfg.MaxGames {
		total += int(rate)
		steps++
		rate *= rampCfg.Multiplier
	}

	return time.Duration(steps) * step
}

// convertStepResults converts orchestrator StepOutputs to journal StepResults
// and baseline LevelResults.
func convertStepResults(
	outputs []orchestrator.StepOutput,
	slos baseline.SLOSet,
) ([]journal.StepResult, []baseline.LevelResult) {
	stepResults := make([]journal.StepResult, len(outputs))
	levelResults := make([]baseline.LevelResult, len(outputs))

	for i, so := range outputs {
		metricsSnap := baseline.SnapshotToMetrics(so.Snapshot, so.Duration.Seconds())
		evalResult := slos.Evaluate(metricsSnap)

		stepResults[i] = journal.StepResult{
			TargetGames:        so.TargetGames,
			Metrics:            metricsSnap,
			SLOEval:            evalResult,
			ServerResources:    so.ServerResources,
			DurationSec:        so.Duration.Seconds(),
			HealthDistribution: so.HealthDistribution,
			DBStats:            so.DBStats,
		}

		levelResults[i] = baseline.LevelResult{
			Games:   so.TargetGames,
			Metrics: metricsSnap,
		}
	}

	return stepResults, levelResults
}

// findCeilingInsights returns insights from the step that matches the ceiling,
// or the last step if no ceiling.
func findCeilingInsights(
	stepResults []journal.StepResult,
	ceiling journal.SLOCeiling,
) []baseline.Insight {
	if len(stepResults) == 0 {
		return nil
	}

	lastIdx := len(stepResults) - 1
	if ceiling.Games > 0 {
		for i, sr := range stepResults {
			if sr.TargetGames == ceiling.Games {
				lastIdx = i

				break
			}
		}
	}

	return baseline.Analyze(stepResults[lastIdx].Metrics)
}

// getGitInfo returns the current short commit SHA and branch name.
func getGitInfo() (commitSHA, branch string) {
	commitSHA = "unknown"

	out, err := exec.Command("git", "rev-parse", "--short", "HEAD").Output()
	if err == nil {
		commitSHA = strings.TrimSpace(string(out))
	}

	branchOut, err := exec.Command("git", "branch", "--show-current").Output()
	if err == nil {
		branch = strings.TrimSpace(string(branchOut))
	}

	return commitSHA, branch
}

// saveJournalEntry saves the journal entry and handles session integration.
func saveJournalEntry(
	entry journal.Entry,
	slug string,
	branch, commitSHA, hypothesis string,
) {
	path, err := journal.SaveEntry("perf-journal/entries", slug, entry)
	if err != nil {
		log.Printf("failed to save journal entry: %v", err)

		return
	}

	log.Printf("journal entry saved: %s", path)

	if branch != "" && branch != "main" && branch != "master" {
		handleSession(branch, commitSHA, path, entry.SLOCeiling.Games, hypothesis)
	}
}

// compareJournalEntries prints a comparison between two journal entries.
func compareJournalEntries(compareFile string, entry journal.Entry) {
	prevEntry, err := journal.LoadEntry(compareFile)
	if err != nil {
		log.Fatalf("failed to load journal entry: %v", err)
	}

	journal.PrintCeilingComparison(os.Stdout, prevEntry, entry)
}

// handleSession manages session state for branch-scoped optimization tracking.
func handleSession(branch, commitSHA, entryPath string, ceilingGames int, hypothesis string) {
	store := session.NewStore("perf-journal/sessions")

	sess, err := store.GetOrCreate(branch, "perf-journal/entries")
	if err != nil {
		log.Printf("[session] failed to create/load session: %v", err)

		return
	}

	// Compute delta vs session baseline ceiling.
	baselineCeiling := 0
	if len(sess.Runs) > 0 {
		baselineCeiling = sess.Runs[0].CeilingGames
	}

	ref := session.RunRef{
		EntryPath:    entryPath,
		CommitSHA:    commitSHA,
		CeilingGames: ceilingGames,
		CeilingDelta: ceilingGames - baselineCeiling,
		Hypothesis:   hypothesis,
		Timestamp:    time.Now(),
	}

	if err := store.AddRun(branch, ref); err != nil {
		log.Printf("[session] failed to add run: %v", err)

		return
	}

	runNum := len(sess.Runs) + 1
	log.Printf(
		"[session] %s: run #%d, ceiling=%d (delta=%+d from session start)",
		branch, runNum, ceilingGames, ref.CeilingDelta,
	)

	if hypothesis != "" {
		log.Printf("[session] hypothesis: %s", hypothesis)
	}
}

// handleBaseline saves and/or compares baselines.
func handleBaseline(
	cfg *Config,
	metricsSnap baseline.MetricsSnapshot,
	insights []baseline.Insight,
) {
	if !cfg.Output.SaveBaseline && cfg.Output.CompareFile == "" {
		return
	}

	currentBaseline := buildCurrentBaseline(cfg, metricsSnap, insights)

	if cfg.Output.SaveBaseline {
		var path string
		var err error

		if cfg.Output.BaselineName != "" {
			path, err = baseline.SaveNumbered(
				"perf-journal/baselines",
				cfg.Output.BaselineName,
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

	if cfg.Output.CompareFile != "" {
		referenceBaseline, err := baseline.Load(cfg.Output.CompareFile)
		if err != nil {
			log.Fatalf("failed to load baseline: %v", err)
		}

		fmt.Println()
		baseline.PrintComparison(os.Stdout, referenceBaseline, currentBaseline)
	}
}

// buildCurrentBaseline creates a Baseline from the current run.
func buildCurrentBaseline(
	cfg *Config,
	metricsSnap baseline.MetricsSnapshot,
	insights []baseline.Insight,
) baseline.Baseline {
	commitSHA := "unknown"

	out, err := exec.Command("git", "rev-parse", "--short", "HEAD").Output()
	if err == nil {
		commitSHA = strings.TrimSpace(string(out))
	}

	return baseline.Baseline{
		CommitSHA: commitSHA,
		Timestamp: time.Now(),
		TestParams: baseline.TestParams{
			Preset:  cfg.Preset,
			Players: cfg.Game.NumPlayers,
			Games:   cfg.NumGames,
			Mode:    cfg.Mode,
		},
		Metrics:     metricsSnap,
		Environment: baseline.CaptureEnvironment(),
		Insights:    insights,
	}
}
