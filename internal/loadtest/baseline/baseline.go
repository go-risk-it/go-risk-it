package baseline

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
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

// sanitizeSlug normalizes a name for use in filenames: lowercase, hyphens, no special chars.
func sanitizeSlug(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")

	// Strip anything that isn't alphanumeric or hyphen.
	re := regexp.MustCompile(`[^a-z0-9-]`)
	s = re.ReplaceAllString(s, "")

	// Collapse multiple hyphens.
	re = regexp.MustCompile(`-+`)
	s = re.ReplaceAllString(s, "-")

	return strings.Trim(s, "-")
}

// nextSequenceNumber scans dir for files matching NNN-* and returns the next number.
func nextSequenceNumber(dir string) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, nil // directory doesn't exist yet, start at 0
	}

	highest := -1
	re := regexp.MustCompile(`^(\d{3})-`)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		matches := re.FindStringSubmatch(entry.Name())
		if matches == nil {
			continue
		}

		n, err := strconv.Atoi(matches[1])
		if err != nil {
			continue
		}

		if n > highest {
			highest = n
		}
	}

	return highest + 1, nil
}

// SaveNumbered writes baseline as JSON to dir/NNN-<slug>-<commit>.json with auto-incrementing
// sequence number. Returns the path of the written file.
func SaveNumbered(dir, slug string, baselineData Baseline) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create dir: %w", err)
	}

	seq, err := nextSequenceNumber(dir)
	if err != nil {
		return "", fmt.Errorf("sequence number: %w", err)
	}

	slug = sanitizeSlug(slug)
	filename := fmt.Sprintf("%03d-%s-%s.json", seq, slug, baselineData.CommitSHA)
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
