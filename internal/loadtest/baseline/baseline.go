package baseline

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/loadtest/fileutil"
)

// Baseline captures a complete performance snapshot for a git commit.
type Baseline struct {
	CommitSHA      string          `json:"commitSha"`
	Timestamp      time.Time       `json:"timestamp"`
	TestParams     TestParams      `json:"testParams"`
	Metrics        MetricsSnapshot `json:"metrics"`
	Environment    Environment     `json:"environment"`
	BreakingPoints []BreakingPoint `json:"breakingPoints,omitempty"`
	Insights       []Insight       `json:"insights,omitempty"`
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
	SLOName       string  `json:"sloName"`
	BreaksAtGames int     `json:"breaksAtGames"`
	LastGoodValue float64 `json:"lastGoodValue"`
	BreakValue    float64 `json:"breakValue"`
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

	if err := os.WriteFile(path, data, 0o600); err != nil {
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

// SaveNumbered writes baseline as JSON to dir/NNN-<slug>-<commit>.json with auto-incrementing
// sequence number. Returns the path of the written file.
func SaveNumbered(dir, slug string, baselineData Baseline) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create dir: %w", err)
	}

	seq, err := fileutil.NextSequenceNumber(dir)
	if err != nil {
		return "", fmt.Errorf("sequence number: %w", err)
	}

	slug = fileutil.SanitizeSlug(slug)
	filename := fmt.Sprintf("%03d-%s-%s.json", seq, slug, baselineData.CommitSHA)
	path := filepath.Join(dir, filename)

	data, err := json.MarshalIndent(baselineData, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal baseline: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return "", fmt.Errorf("write baseline: %w", err)
	}

	return path, nil
}
