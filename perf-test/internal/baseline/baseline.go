package baseline

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Baseline captures a complete performance snapshot for a git commit.
type Baseline struct {
	CommitSHA      string          `json:"commit_sha"`
	Timestamp      time.Time       `json:"timestamp"`
	TestParams     TestParams      `json:"test_params"`
	Metrics        MetricsSnapshot `json:"metrics"`
	Environment    Environment     `json:"environment"`
	BreakingPoints []BreakingPoint `json:"breaking_points,omitempty"`
}

// TestParams describes the test configuration that produced this baseline.
type TestParams struct {
	Preset  string `json:"preset"`
	Players int    `json:"players"`
	Games   int    `json:"games"`
	Mode    string `json:"mode"`
}

// BreakingPoint records the concurrency level at which an SLO first fails.
type BreakingPoint struct {
	SLOName       string  `json:"slo_name"`
	BreaksAtGames int     `json:"breaks_at_games"`
	LastGoodValue float64 `json:"last_good_value"`
	BreakValue    float64 `json:"break_value"`
}

// Save writes baseline as JSON to dir/<commit>-<date>.json.
func Save(dir string, baselineData Baseline) (string, error) {
	dateStr := baselineData.Timestamp.Format("20060102-150405")
	filename := fmt.Sprintf("%s-%s.json", baselineData.CommitSHA, dateStr)
	path := filepath.Join(dir, filename)

	data, err := json.MarshalIndent(baselineData, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal baseline: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", fmt.Errorf("write baseline: %w", err)
	}

	return path, nil
}

// Load reads a baseline from a JSON file.
func Load(path string) (Baseline, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Baseline{}, fmt.Errorf("read baseline: %w", err)
	}

	var baselineData Baseline
	if err := json.Unmarshal(data, &baselineData); err != nil {
		return Baseline{}, fmt.Errorf("unmarshal baseline: %w", err)
	}

	return baselineData, nil
}
