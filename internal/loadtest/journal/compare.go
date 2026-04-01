package journal

import "github.com/go-risk-it/go-risk-it/internal/loadtest/baseline"

// CeilingDelta describes the change in SLO ceiling between two journal entries.
type CeilingDelta struct {
	GamesBefore        int
	GamesAfter         int
	GamesDelta         int     // positive = improvement
	ThroughputDeltaPct float64 // percentage change
	CompletionDeltaPct float64
}

// BottleneckShift detects when the first-breaking SLO changed between runs.
type BottleneckShift struct {
	Before  string // SLO name that broke first in before entry (empty if none)
	After   string // SLO name that broke first in after entry (empty if none)
	Shifted bool   // true if Before != After (and both are non-empty)
}

// CompareCeilings computes the delta between two SLO ceilings.
func CompareCeilings(before, after SLOCeiling) CeilingDelta {
	delta := CeilingDelta{
		GamesBefore: before.Games,
		GamesAfter:  after.Games,
		GamesDelta:  after.Games - before.Games,
	}

	if before.ThroughputMPS > 0 {
		delta.ThroughputDeltaPct = (after.ThroughputMPS - before.ThroughputMPS) /
			before.ThroughputMPS * 100
	}

	if before.CompletionRate > 0 {
		delta.CompletionDeltaPct = (after.CompletionRate - before.CompletionRate) /
			before.CompletionRate * 100
	}

	return delta
}

// DetectBottleneckShift finds the first-breaking SLO in each entry and checks
// if it changed.
func DetectBottleneckShift(before, after Entry) BottleneckShift {
	shift := BottleneckShift{
		Before: firstViolationName(before.Steps),
		After:  firstViolationName(after.Steps),
	}

	shift.Shifted = shift.Before != "" && shift.After != "" &&
		shift.Before != shift.After

	return shift
}

// firstViolationName returns the name of the first SLO violation found across
// all steps.
func firstViolationName(steps []StepResult) string {
	for _, step := range steps {
		for _, v := range step.SLOEval.Violations {
			return extractSLOName(v)
		}
	}

	return ""
}

// extractSLOName returns the Name field from a Violation's SLO.
func extractSLOName(v baseline.Violation) string {
	return v.SLO.Name
}
