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

// MetricsSnapshot holds key metrics captured at a point in time.
type MetricsSnapshot struct {
	E2EP95         float64 `json:"e2e_p95_s"`
	WSDeliveryP95  float64 `json:"ws_delivery_p95_s"`
	DBTxnP95       float64 `json:"db_txn_p95_s"`
	DBPoolUtil     float64 `json:"db_pool_util"`
	WSBroadcastP95 float64 `json:"ws_broadcast_p95_s"`
	HTTPErrorRate  float64 `json:"http_error_rate"`
	ThroughputMPS  float64 `json:"throughput_moves_per_sec"`
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

// DefaultSLOs returns the standard SLO set matching the plan's SLO Reference.
func DefaultSLOs() SLOSet {
	return SLOSet{
		UserExperience: []SLO{
			{Name: "E2E move latency", Metric: "e2e_p95_s", Threshold: 0.5, Unit: "s"},
			{Name: "WS delivery latency", Metric: "ws_delivery_p95_s", Threshold: 0.2, Unit: "s"},
			{Name: "HTTP error rate", Metric: "http_error_rate", Threshold: 0.01, Unit: "ratio"},
		},
		BoundaryHealth: []SLO{
			{Name: "DB transaction latency", Metric: "db_txn_p95_s", Threshold: 0.05, Unit: "s"},
			{Name: "DB pool utilization", Metric: "db_pool_util", Threshold: 0.8, Unit: "ratio"},
			{
				Name:      "WS broadcast latency",
				Metric:    "ws_broadcast_p95_s",
				Threshold: 0.1,
				Unit:      "s",
			},
		},
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
		"e2e_p95_s":          snap.E2EP95,
		"ws_delivery_p95_s":  snap.WSDeliveryP95,
		"db_txn_p95_s":       snap.DBTxnP95,
		"db_pool_util":       snap.DBPoolUtil,
		"ws_broadcast_p95_s": snap.WSBroadcastP95,
		"http_error_rate":    snap.HTTPErrorRate,
	}
}
