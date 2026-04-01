package journal_test

import (
	"bytes"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/loadtest/baseline"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/journal"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/resources"
	"github.com/stretchr/testify/assert"
)

func makeTestEntry(numSteps int, firstPasses bool) journal.Entry {
	entry := journal.Entry{
		SLOCeiling: journal.SLOCeiling{Games: 5, ThroughputMPS: 30, CompletionRate: 0.95},
		Steps:      make([]journal.StepResult, numSteps),
	}

	for i := range numSteps {
		entry.Steps[i] = journal.StepResult{
			TargetGames: (i + 1) * 5,
			Metrics: baseline.MetricsSnapshot{
				E2E:           baseline.LatencyProfile{P95: 0.1 + float64(i)*0.15},
				WSDelivery:    baseline.LatencyProfile{P95: 0.04 + float64(i)*0.08},
				ThroughputMPS: 30 + float64(i)*25,
			},
			ServerResources: resources.ServerResources{
				RiskIt: resources.ContainerStats{
					CPUPercent: 12 + float64(i)*30,
					MemoryMB:   180 + float64(i)*100,
				},
			},
		}

		if !firstPasses || i > 0 {
			entry.Steps[i].SLOEval = baseline.EvalResult{
				Violations: []baseline.Violation{
					{
						SLO:    baseline.SLO{Name: "E2E move latency p95"},
						Actual: 0.6,
					},
				},
			}
		}
	}

	if firstPasses && numSteps > 0 {
		entry.Steps[0].SLOEval = baseline.EvalResult{}
	}

	return entry
}

func TestPrintStaircaseReport_WithCeiling(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	entry := makeTestEntry(2, true)
	journal.PrintStaircaseReport(&buf, entry)
	output := buf.String()
	assert.Contains(t, output, "Ceiling:")
	assert.Contains(t, output, "PASS")
	assert.Contains(t, output, "FAIL")
}

func TestPrintStaircaseReport_ZeroCeiling(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	entry := journal.Entry{
		SLOCeiling: journal.SLOCeiling{},
		Steps: []journal.StepResult{
			{
				TargetGames: 5,
				SLOEval: baseline.EvalResult{
					Violations: []baseline.Violation{
						{
							SLO:    baseline.SLO{Name: "x"},
							Actual: 1.0,
						},
					},
				},
			},
		},
	}
	journal.PrintStaircaseReport(&buf, entry)
	assert.Contains(t, buf.String(), "No SLO ceiling found")
}

func TestPrintStaircaseReport_WithBreakingPoints(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	entry := makeTestEntry(2, true)
	entry.BreakingPoints = []baseline.BreakingPoint{
		{
			SLOName:       "E2E move latency p95",
			BreaksAtGames: 10,
			LastGoodValue: 0.32,
			BreakValue:    0.65,
		},
	}
	journal.PrintStaircaseReport(&buf, entry)
	assert.Contains(t, buf.String(), "Breaking points:")
	assert.Contains(t, buf.String(), "E2E move latency p95")
}

func TestPrintCeilingComparison_Improvement(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	before := journal.Entry{
		SLOCeiling: journal.SLOCeiling{Games: 40, ThroughputMPS: 125, CompletionRate: 0.95},
	}
	after := journal.Entry{
		SLOCeiling: journal.SLOCeiling{Games: 80, ThroughputMPS: 200, CompletionRate: 0.98},
	}
	journal.PrintCeilingComparison(&buf, before, after)
	output := buf.String()
	assert.Contains(t, output, "40 -> 80")
	assert.Contains(t, output, "+40")
}

func TestPrintCeilingComparison_BottleneckShift(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	before := journal.Entry{
		SLOCeiling: journal.SLOCeiling{Games: 40, ThroughputMPS: 125, CompletionRate: 0.95},
		Steps: []journal.StepResult{
			{
				SLOEval: baseline.EvalResult{Violations: []baseline.Violation{
					{SLO: baseline.SLO{Name: "E2E move latency p95"}, Actual: 0.6},
				}},
			},
		},
	}
	after := journal.Entry{
		SLOCeiling: journal.SLOCeiling{Games: 80, ThroughputMPS: 200, CompletionRate: 0.98},
		Steps: []journal.StepResult{
			{
				SLOEval: baseline.EvalResult{Violations: []baseline.Violation{
					{SLO: baseline.SLO{Name: "Move failure rate"}, Actual: 0.2},
				}},
			},
		},
	}
	journal.PrintCeilingComparison(&buf, before, after)
	assert.Contains(t, buf.String(), "Bottleneck shift")
}
