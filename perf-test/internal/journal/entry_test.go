package journal_test

import (
	"testing"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/baseline"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/journal"
	"github.com/stretchr/testify/assert"
)

func TestFindSLOCeiling_AllPassing(t *testing.T) {
	t.Parallel()

	steps := []journal.StepResult{
		{
			TargetGames: 5,
			Metrics:     baseline.MetricsSnapshot{ThroughputMPS: 30, GamesCompleted: 5},
			SLOEval:     baseline.EvalResult{},
		},
		{
			TargetGames: 10,
			Metrics: baseline.MetricsSnapshot{
				ThroughputMPS:  60,
				GamesCompleted: 9,
				GamesTimedOut:  1,
			},
			SLOEval: baseline.EvalResult{},
		},
	}
	ceiling := journal.FindSLOCeiling(steps)
	assert.Equal(t, 10, ceiling.Games)
	assert.InDelta(t, 60.0, ceiling.ThroughputMPS, 0.01)
	assert.InDelta(t, 0.9, ceiling.CompletionRate, 0.01) // 9/10
}

func TestFindSLOCeiling_BreachAtSecondStep(t *testing.T) {
	t.Parallel()

	steps := []journal.StepResult{
		{
			TargetGames: 5,
			Metrics:     baseline.MetricsSnapshot{ThroughputMPS: 30, GamesCompleted: 5},
			SLOEval:     baseline.EvalResult{},
		},
		{
			TargetGames: 10,
			Metrics:     baseline.MetricsSnapshot{ThroughputMPS: 60},
			SLOEval: baseline.EvalResult{
				Violations: []baseline.Violation{
					{SLO: baseline.SLO{Name: "E2E move latency p95"}, Actual: 0.6},
				},
			},
		},
	}
	ceiling := journal.FindSLOCeiling(steps)
	assert.Equal(t, 5, ceiling.Games)
	assert.InDelta(t, 30.0, ceiling.ThroughputMPS, 0.01)
}

func TestFindSLOCeiling_FirstStepFails(t *testing.T) {
	t.Parallel()

	steps := []journal.StepResult{
		{
			TargetGames: 5,
			SLOEval: baseline.EvalResult{
				Violations: []baseline.Violation{
					{SLO: baseline.SLO{Name: "HTTP error rate"}, Actual: 0.5},
				},
			},
		},
	}
	ceiling := journal.FindSLOCeiling(steps)
	assert.Equal(t, 0, ceiling.Games)
}

func TestFindSLOCeiling_NoSteps(t *testing.T) {
	t.Parallel()

	ceiling := journal.FindSLOCeiling(nil)
	assert.Equal(t, 0, ceiling.Games)
}

func TestEntry_PassingSteps(t *testing.T) {
	t.Parallel()

	entry := journal.Entry{
		Steps: []journal.StepResult{
			{TargetGames: 5, SLOEval: baseline.EvalResult{}},
			{
				TargetGames: 10,
				SLOEval: baseline.EvalResult{
					Violations: []baseline.Violation{{SLO: baseline.SLO{Name: "x"}, Actual: 1.0}},
				},
			},
			{TargetGames: 20, SLOEval: baseline.EvalResult{}},
		},
	}
	passing := entry.PassingSteps()
	assert.Len(t, passing, 2)
	assert.Equal(t, 5, passing[0].TargetGames)
	assert.Equal(t, 20, passing[1].TargetGames)
}

func TestEntry_BreachStep_Found(t *testing.T) {
	t.Parallel()

	entry := journal.Entry{
		Steps: []journal.StepResult{
			{TargetGames: 5, SLOEval: baseline.EvalResult{}},
			{
				TargetGames: 10,
				SLOEval: baseline.EvalResult{
					Violations: []baseline.Violation{{SLO: baseline.SLO{Name: "x"}, Actual: 1.0}},
				},
			},
		},
	}
	breach := entry.BreachStep()
	assert.NotNil(t, breach)
	assert.Equal(t, 10, breach.TargetGames)
}

func TestEntry_BreachStep_NoBreach(t *testing.T) {
	t.Parallel()

	entry := journal.Entry{
		Steps: []journal.StepResult{
			{TargetGames: 5, SLOEval: baseline.EvalResult{}},
			{TargetGames: 10, SLOEval: baseline.EvalResult{}},
		},
	}
	assert.Nil(t, entry.BreachStep())
}
