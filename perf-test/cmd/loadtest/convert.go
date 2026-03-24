package main

import (
	"github.com/go-risk-it/go-risk-it/perf-test/internal/baseline"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/journal"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/orchestrator"
)

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
