package journal

import (
	"time"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/baseline"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/health"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/resources"
)

// StaircaseParams records the staircase configuration used for a run.
type StaircaseParams struct {
	Mode              string  `json:"mode,omitempty"` // "staircase" or "adaptive"
	Steps             []int   `json:"steps"`
	HoldDurationSec   float64 `json:"hold_duration_sec"`
	NumPlayers        int     `json:"num_players"`
	GameTimeoutSec    float64 `json:"game_timeout_sec"`
	StopOnBreach      bool    `json:"stop_on_breach"`
	WarmUpCompletions int     `json:"warm_up_completions,omitempty"`
	WarmUpDurationSec int     `json:"warm_up_duration_sec,omitempty"`
}

// StepResult holds metrics and SLO evaluation for a single staircase step.
type StepResult struct {
	TargetGames        int                       `json:"target_games"`
	Metrics            baseline.MetricsSnapshot  `json:"metrics"`
	SLOEval            baseline.EvalResult       `json:"slo_eval"`
	ServerResources    resources.ServerResources `json:"server_resources"`
	DurationSec        float64                   `json:"duration_sec"`
	HealthDistribution *health.Distribution      `json:"health_distribution,omitempty"`
}

// SLOCeiling is the headline metric: max concurrent games where all SLOs hold.
type SLOCeiling struct {
	Games                int     `json:"games"`
	ThroughputMPS        float64 `json:"throughput_mps"`
	CompletionRate       float64 `json:"completion_rate"`
	EffectiveConcurrency int     `json:"effective_concurrency,omitempty"`
}

// Entry is a complete performance journal entry.
type Entry struct {
	CommitSHA      string                   `json:"commit_sha"`
	Timestamp      time.Time                `json:"timestamp"`
	Branch         string                   `json:"branch,omitempty"`
	SessionID      string                   `json:"session_id,omitempty"`
	Config         StaircaseParams          `json:"config"`
	SLOCeiling     SLOCeiling               `json:"slo_ceiling"`
	Steps          []StepResult             `json:"steps"`
	BreakingPoints []baseline.BreakingPoint `json:"breaking_points,omitempty"`
	Environment    baseline.Environment     `json:"environment"`
	Insights       []baseline.Insight       `json:"insights,omitempty"`
}

// FindSLOCeiling walks steps in order, returns the last step where AllPassing().
// If no step passes, returns zero SLOCeiling.
func FindSLOCeiling(steps []StepResult) SLOCeiling {
	var ceiling SLOCeiling

	for _, step := range steps {
		if !step.SLOEval.AllPassing() {
			break
		}

		total := step.Metrics.GamesCompleted + step.Metrics.GamesTimedOut + step.Metrics.GamesFatal

		var completionRate float64
		if total > 0 {
			completionRate = float64(step.Metrics.GamesCompleted) / float64(total)
		}

		ceiling = SLOCeiling{
			Games:          step.TargetGames,
			ThroughputMPS:  step.Metrics.ThroughputMPS,
			CompletionRate: completionRate,
		}

		if step.HealthDistribution != nil {
			ceiling.EffectiveConcurrency = step.HealthDistribution.EffectiveConcurrency()
		}
	}

	return ceiling
}

// PassingSteps returns only steps where all SLOs pass.
func (e Entry) PassingSteps() []StepResult {
	var passing []StepResult
	for _, step := range e.Steps {
		if step.SLOEval.AllPassing() {
			passing = append(passing, step)
		}
	}

	return passing
}

// BreachStep returns the first step with SLO violations, or nil.
func (e Entry) BreachStep() *StepResult {
	for i := range e.Steps {
		if !e.Steps[i].SLOEval.AllPassing() {
			return &e.Steps[i]
		}
	}

	return nil
}
