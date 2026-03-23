// Package baseline provides performance history tracking: SLO evaluation,
// baseline save/load, comparison, and breaking point detection.
package baseline

// SLO defines a single service-level objective with a name, metric key,
// threshold, and unit.
type SLO struct {
	Name      string  `json:"name"`
	Metric    string  `json:"metric"`
	Threshold float64 `json:"threshold"`
	Unit      string  `json:"unit"`
	LowerBad  bool    `json:"lower_bad,omitempty"` // true if values below threshold are bad
}

// SLOSet groups SLOs into user-experience and boundary-health tiers.
type SLOSet struct {
	UserExperience []SLO `json:"user_experience"`
	BoundaryHealth []SLO `json:"boundary_health"`
}

// LatencyProfile holds percentile values for a latency distribution (in seconds).
type LatencyProfile struct {
	P50 float64 `json:"p50_s"`
	P95 float64 `json:"p95_s"`
	P99 float64 `json:"p99_s"`
	Max float64 `json:"max_s"`
}

// MetricsSnapshot holds key metrics captured at a point in time.
type MetricsSnapshot struct {
	// Core latency profiles.
	E2E        LatencyProfile `json:"e2e"`
	WSDelivery LatencyProfile `json:"ws_delivery"`

	// Throughput.
	ThroughputMPS     float64 `json:"throughput_moves_per_sec"`
	ThroughputPeakMPS float64 `json:"throughput_peak_moves_per_sec"`

	// Error rate.
	HTTPErrorRate float64 `json:"http_error_rate"`

	// Counters.
	TotalMoves     int64 `json:"total_moves"`
	TotalErrors    int64 `json:"total_errors"`
	GamesCompleted int64 `json:"games_completed"`
	GamesTimedOut  int64 `json:"games_timed_out"`
	GamesFatal     int64 `json:"games_fatal"`

	// Resilience.
	TotalRetries           int64 `json:"total_retries"`
	TotalConflicts         int64 `json:"total_conflicts"`
	TotalReconnects        int64 `json:"total_reconnects"`
	TotalReconnectFailures int64 `json:"total_reconnect_failures"`

	// Phase latency (per game phase).
	PhaseLatency map[string]LatencyProfile `json:"phase_latency,omitempty"`

	// REST per-action latency.
	RESTLatency map[string]LatencyProfile `json:"rest_latency,omitempty"`

	// Error breakdown by category.
	ErrorBreakdown map[string]int64 `json:"error_breakdown,omitempty"`

	// Phase flow.
	PhaseEntries map[string]int64 `json:"phase_entries,omitempty"`
	PhaseMoves   map[string]int64 `json:"phase_moves,omitempty"`

	// Test duration.
	DurationSec float64 `json:"duration_sec"`
}

// Violation records an SLO that was breached, along with the actual value.
type Violation struct {
	SLO    SLO     `json:"slo"`
	Actual float64 `json:"actual"`
}

// EvalResult holds the outcome of evaluating a MetricsSnapshot against an SLOSet.
type EvalResult struct {
	Violations []Violation `json:"violations"`
}

// AllPassing returns true if no SLOs were violated.
func (r EvalResult) AllPassing() bool {
	return len(r.Violations) == 0
}

// DefaultSLOs returns the standard SLO set.
func DefaultSLOs() SLOSet {
	return SLOSet{
		UserExperience: []SLO{
			{Name: "E2E move latency p95", Metric: "e2e_p95_s", Threshold: 0.5, Unit: "s"},
			{Name: "E2E move latency p99", Metric: "e2e_p99_s", Threshold: 1.0, Unit: "s"},
			{
				Name:      "WS delivery latency p95",
				Metric:    "ws_delivery_p95_s",
				Threshold: 0.2,
				Unit:      "s",
			},
			{
				Name:      "WS delivery latency p99",
				Metric:    "ws_delivery_p99_s",
				Threshold: 0.5,
				Unit:      "s",
			},
			{Name: "HTTP error rate", Metric: "http_error_rate", Threshold: 0.01, Unit: "ratio"},
		},
		BoundaryHealth: []SLO{},
	}
}

// Evaluate checks each SLO against the snapshot and returns violations.
func (s SLOSet) Evaluate(snap MetricsSnapshot) EvalResult {
	values := metricValues(snap)
	var violations []Violation

	for _, slo := range append(s.UserExperience, s.BoundaryHealth...) {
		actual, exists := values[slo.Metric]
		if !exists {
			continue
		}

		if actual > slo.Threshold {
			violations = append(violations, Violation{SLO: slo, Actual: actual})
		}
	}

	return EvalResult{Violations: violations}
}

// metricValues maps SLO metric keys to snapshot values.
func metricValues(snap MetricsSnapshot) map[string]float64 {
	return map[string]float64{
		"e2e_p95_s":         snap.E2E.P95,
		"e2e_p99_s":         snap.E2E.P99,
		"ws_delivery_p95_s": snap.WSDelivery.P95,
		"ws_delivery_p99_s": snap.WSDelivery.P99,
		"http_error_rate":   snap.HTTPErrorRate,
	}
}
