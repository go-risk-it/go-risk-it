//go:build integration

package internal_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/loadtest/annotations"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/baseline"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/journal"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/metrics"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/orchestrator"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/resources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStaircasePipeline_EndToEnd runs a staircase with fake games,
// saves a journal entry, loads it back, runs a second staircase,
// and compares the two entries.
func TestStaircasePipeline_EndToEnd(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	slos := baseline.DefaultSLOs()
	steps := []int{2, 4}

	execCfg := orchestrator.StepExecutorConfig{
		NumPlayers:   2,
		GameTimeout:  time.Second,
		StaggerDelay: 5 * time.Millisecond,
		HoldDuration: 80 * time.Millisecond,
	}

	execDeps := orchestrator.StepExecutorDeps{
		RunnerFactory: func(c *metrics.Collector, _ orchestrator.GameObserver) orchestrator.RunFunc {
			return func(ctx context.Context, idx, players int) orchestrator.GameResult {
				c.RecordMove()
				c.RecordTimedMove()
				c.RecordE2E(5 * time.Millisecond)

				select {
				case <-ctx.Done():
				case <-time.After(15 * time.Millisecond):
				}

				return orchestrator.GameResult{GameIndex: idx, Moves: 10}
			}
		},
		NewCollector: metrics.NewCollector,
		CollectResources: func() resources.ServerResources {
			return resources.ServerResources{
				RiskIt: resources.ContainerStats{CPUPercent: 25.0, MemoryMB: 128},
			}
		},
		Annotator: annotations.NewAnnotator(""),
	}

	executor := orchestrator.NewStepExecutor(execCfg, execDeps, len(steps))

	// Run staircase.
	outputs := orchestrator.RunStaircase(
		context.Background(), steps, 1,
		executor, &orchestrator.NeverStop{}, annotations.NewAnnotator(""),
	)
	require.Len(t, outputs, 2, "expected 2 step outputs")

	// Convert to journal StepResults.
	stepResults := make([]journal.StepResult, len(outputs))

	for i, so := range outputs {
		metricsSnap := baseline.SnapshotToMetrics(so.Snapshot, so.Duration.Seconds())
		evalResult := slos.Evaluate(metricsSnap)

		stepResults[i] = journal.StepResult{
			TargetGames:     so.TargetGames,
			Metrics:         metricsSnap,
			SLOEval:         evalResult,
			ServerResources: so.ServerResources,
			DurationSec:     so.Duration.Seconds(),
		}
	}

	// Verify resource collection came through.
	assert.InDelta(t, 25.0, stepResults[0].ServerResources.RiskIt.CPUPercent, 0.01)

	// Build and save first journal entry.
	ceiling := journal.FindSLOCeiling(stepResults)
	entry1 := journal.Entry{
		CommitSHA:  "abc1234",
		Timestamp:  time.Now(),
		SLOCeiling: ceiling,
		Steps:      stepResults,
		Config: journal.StaircaseParams{
			Steps:           steps,
			HoldDurationSec: execCfg.HoldDuration.Seconds(),
			NumPlayers:      execCfg.NumPlayers,
			GameTimeoutSec:  execCfg.GameTimeout.Seconds(),
		},
	}

	path1, err := journal.SaveEntry(dir, "test-run", entry1)
	require.NoError(t, err)
	require.FileExists(t, path1)

	// Load it back and verify roundtrip.
	loaded, err := journal.LoadEntry(path1)
	require.NoError(t, err)
	assert.Equal(t, entry1.CommitSHA, loaded.CommitSHA)
	assert.Len(t, loaded.Steps, 2)
	assert.Equal(t, ceiling.Games, loaded.SLOCeiling.Games)

	// Save a second entry and verify sequencing.
	entry2 := entry1
	entry2.CommitSHA = "def5678"

	path2, err := journal.SaveEntry(dir, "test-run", entry2)
	require.NoError(t, err)
	assert.NotEqual(t, path1, path2)

	// List entries.
	paths, err := journal.ListEntries(dir)
	require.NoError(t, err)
	assert.Len(t, paths, 2)

	// Compare ceilings.
	delta := journal.CompareCeilings(entry1.SLOCeiling, entry2.SLOCeiling)
	assert.Equal(t, 0, delta.GamesDelta) // same config, should be identical
}

// TestDockerStatsParsing tests parsing real docker stats JSON format.
func TestDockerStatsParsing(t *testing.T) {
	t.Parallel()

	// This mirrors real docker stats --no-stream --format json output.
	jsonLine := `{"Name":"go-risk-it-risk-it-1","CPUPerc":"12.50%","MemUsage":"45.2MiB / 8GiB"}`

	stats, err := resources.ParseDockerStats([]byte(jsonLine))
	require.NoError(t, err)
	assert.InDelta(t, 12.50, stats.CPUPercent, 0.01)
	assert.InDelta(t, 45.2, stats.MemoryMB, 0.1)
}

// TestLatestEntry_AfterMultipleSaves tests LatestEntry returns the highest.
func TestLatestEntry_AfterMultipleSaves(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	for i := 0; i < 3; i++ {
		entry := journal.Entry{
			CommitSHA: "commit" + string(rune('a'+i)),
			Timestamp: time.Now(),
			SLOCeiling: journal.SLOCeiling{
				Games: (i + 1) * 10,
			},
		}
		_, err := journal.SaveEntry(dir, "multi", entry)
		require.NoError(t, err)
	}

	latest, err := journal.LatestEntry(dir)
	require.NoError(t, err)
	assert.Equal(t, 30, latest.SLOCeiling.Games)
}

// TestStaircase_StopOnBreach_ProducesPartialOutput verifies that when
// SLO breach is detected, we still get the breach step in the output.
func TestStaircase_StopOnBreach_ProducesPartialOutput(t *testing.T) {
	t.Parallel()

	steps := []int{2, 4, 8}
	slos := baseline.SLOSet{
		UserExperience: []baseline.SLO{
			{
				Name:      "impossible",
				Metric:    "e2e_p95_s",
				Threshold: 0.000001,
				Unit:      "s",
			},
		},
	}

	execCfg := orchestrator.StepExecutorConfig{
		NumPlayers:   2,
		GameTimeout:  time.Second,
		StaggerDelay: 5 * time.Millisecond,
		HoldDuration: 60 * time.Millisecond,
	}

	execDeps := orchestrator.StepExecutorDeps{
		RunnerFactory: func(c *metrics.Collector, _ orchestrator.GameObserver) orchestrator.RunFunc {
			return func(ctx context.Context, idx, players int) orchestrator.GameResult {
				c.RecordMove()
				c.RecordTimedMove()
				c.RecordE2E(50 * time.Millisecond)

				select {
				case <-ctx.Done():
				case <-time.After(10 * time.Millisecond):
				}

				return orchestrator.GameResult{GameIndex: idx}
			}
		},
		NewCollector:     metrics.NewCollector,
		CollectResources: func() resources.ServerResources { return resources.ServerResources{} },
		Annotator:        annotations.NewAnnotator(""),
	}

	executor := orchestrator.NewStepExecutor(execCfg, execDeps, len(steps))
	stopper := &orchestrator.SLOStopCondition{SLOs: slos}

	outputs := orchestrator.RunStaircase(
		context.Background(), steps, 1,
		executor, stopper, annotations.NewAnnotator(""),
	)

	// Should stop at first step (SLO breach), but include that step.
	require.Len(t, outputs, 1)
	assert.Equal(t, 2, outputs[0].TargetGames)

	// Build step results and verify SLO evaluation.
	snap := baseline.SnapshotToMetrics(outputs[0].Snapshot, outputs[0].Duration.Seconds())
	eval := slos.Evaluate(snap)
	assert.False(t, eval.AllPassing())

	// Ceiling should be 0 (no passing steps).
	stepResults := []journal.StepResult{
		{
			TargetGames: 2,
			Metrics:     snap,
			SLOEval:     eval,
		},
	}
	ceiling := journal.FindSLOCeiling(stepResults)
	assert.Equal(t, 0, ceiling.Games)
}

// TestJournalEntryDir_CreateOnDemand verifies SaveEntry creates nested dirs.
func TestJournalEntryDir_CreateOnDemand(t *testing.T) {
	t.Parallel()

	dir := t.TempDir() + "/nested/deep/dir"
	entry := journal.Entry{CommitSHA: "test123", Timestamp: time.Now()}

	path, err := journal.SaveEntry(dir, "test", entry)
	require.NoError(t, err)

	_, err = os.Stat(path)
	require.NoError(t, err)
}
