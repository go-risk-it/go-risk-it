package journal_test

import (
	"testing"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/baseline"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/journal"
	"github.com/stretchr/testify/assert"
)

func TestCompareCeilings_Improvement(t *testing.T) {
	t.Parallel()

	before := journal.SLOCeiling{Games: 40, ThroughputMPS: 125, CompletionRate: 0.95}
	after := journal.SLOCeiling{Games: 80, ThroughputMPS: 200, CompletionRate: 0.98}
	delta := journal.CompareCeilings(before, after)
	assert.Equal(t, 40, delta.GamesDelta)
	assert.InDelta(t, 60.0, delta.ThroughputDeltaPct, 0.1)
	assert.InDelta(t, 3.16, delta.CompletionDeltaPct, 0.1) // (0.98-0.95)/0.95*100
}

func TestCompareCeilings_Regression(t *testing.T) {
	t.Parallel()

	before := journal.SLOCeiling{Games: 80, ThroughputMPS: 200, CompletionRate: 0.98}
	after := journal.SLOCeiling{Games: 40, ThroughputMPS: 125, CompletionRate: 0.95}
	delta := journal.CompareCeilings(before, after)
	assert.Equal(t, -40, delta.GamesDelta)
	assert.InDelta(t, -37.5, delta.ThroughputDeltaPct, 0.1)
}

func TestCompareCeilings_NoChange(t *testing.T) {
	t.Parallel()

	c := journal.SLOCeiling{Games: 40, ThroughputMPS: 125, CompletionRate: 0.95}
	delta := journal.CompareCeilings(c, c)
	assert.Equal(t, 0, delta.GamesDelta)
	assert.InDelta(t, 0.0, delta.ThroughputDeltaPct, 0.01)
}

func TestCompareCeilings_ZeroBefore(t *testing.T) {
	t.Parallel()

	before := journal.SLOCeiling{}
	after := journal.SLOCeiling{Games: 40, ThroughputMPS: 125, CompletionRate: 0.95}
	delta := journal.CompareCeilings(before, after)
	assert.Equal(t, 40, delta.GamesDelta)
	// No division by zero — percentage stays 0 when before is zero.
	assert.InDelta(t, 0.0, delta.ThroughputDeltaPct, 0.01)
}

func TestDetectBottleneckShift_Shifted(t *testing.T) {
	t.Parallel()

	before := journal.Entry{Steps: []journal.StepResult{
		{
			TargetGames: 40,
			SLOEval: baseline.EvalResult{Violations: []baseline.Violation{
				{SLO: baseline.SLO{Name: "E2E move latency p95"}, Actual: 0.6},
			}},
		},
	}}
	after := journal.Entry{Steps: []journal.StepResult{
		{
			TargetGames: 80,
			SLOEval: baseline.EvalResult{Violations: []baseline.Violation{
				{SLO: baseline.SLO{Name: "Move failure rate"}, Actual: 0.2},
			}},
		},
	}}
	shift := journal.DetectBottleneckShift(before, after)
	assert.True(t, shift.Shifted)
	assert.Equal(t, "E2E move latency p95", shift.Before)
	assert.Equal(t, "Move failure rate", shift.After)
}

func TestDetectBottleneckShift_Same(t *testing.T) {
	t.Parallel()

	makeEntry := func() journal.Entry {
		return journal.Entry{Steps: []journal.StepResult{
			{
				SLOEval: baseline.EvalResult{Violations: []baseline.Violation{
					{SLO: baseline.SLO{Name: "E2E move latency p95"}, Actual: 0.6},
				}},
			},
		}}
	}
	shift := journal.DetectBottleneckShift(makeEntry(), makeEntry())
	assert.False(t, shift.Shifted)
}

func TestDetectBottleneckShift_NoBreach(t *testing.T) {
	t.Parallel()

	entry := journal.Entry{Steps: []journal.StepResult{
		{SLOEval: baseline.EvalResult{}},
	}}
	shift := journal.DetectBottleneckShift(entry, entry)
	assert.False(t, shift.Shifted)
	assert.Empty(t, shift.Before)
	assert.Empty(t, shift.After)
}
