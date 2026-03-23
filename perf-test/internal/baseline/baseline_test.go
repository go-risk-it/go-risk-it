package baseline_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/baseline"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBaseline_SaveAndLoad(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	baselineData := baseline.Baseline{
		CommitSHA: "abc1234",
		Timestamp: time.Now().Truncate(time.Second),
		TestParams: baseline.TestParams{
			Preset:  "medium",
			Players: 4,
			Games:   10,
			Mode:    "batch",
		},
		Metrics: baseline.MetricsSnapshot{
			E2E:        baseline.LatencyProfile{P95: 0.312},
			WSDelivery: baseline.LatencyProfile{P95: 0.1},
		},
	}

	path, err := baseline.Save(dir, baselineData)
	require.NoError(t, err)
	assert.FileExists(t, path)

	loaded, err := baseline.Load(path)
	require.NoError(t, err)
	assert.Equal(t, baselineData.CommitSHA, loaded.CommitSHA)
	assert.InDelta(t, baselineData.Metrics.E2E.P95, loaded.Metrics.E2E.P95, 0.001)
	assert.InDelta(t, baselineData.Metrics.WSDelivery.P95, loaded.Metrics.WSDelivery.P95, 0.001)
	assert.Equal(t, baselineData.TestParams.Preset, loaded.TestParams.Preset)
}

func TestBaseline_Load_NonExistentFile(t *testing.T) {
	t.Parallel()

	_, err := baseline.Load("/nonexistent/path.json")
	require.Error(t, err)
}

func TestSaveNumbered_SequentialNaming(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	b := baseline.Baseline{
		CommitSHA: "abc1234",
		Timestamp: time.Now().Truncate(time.Second),
		TestParams: baseline.TestParams{
			Preset: "light",
			Games:  1,
			Mode:   "batch",
		},
	}

	// First save should be 000.
	path1, err := baseline.SaveNumbered(dir, "origin", b)
	require.NoError(t, err)
	assert.Equal(t, "000-origin-abc1234.json", filepath.Base(path1))

	// Second save should be 001.
	path2, err := baseline.SaveNumbered(dir, "tuning", b)
	require.NoError(t, err)
	assert.Equal(t, "001-tuning-abc1234.json", filepath.Base(path2))

	// Verify roundtrip.
	loaded, err := baseline.Load(path1)
	require.NoError(t, err)
	assert.Equal(t, "abc1234", loaded.CommitSHA)
}

func TestSaveNumbered_SlugSanitization(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	b := baseline.Baseline{CommitSHA: "def5678"}

	path, err := baseline.SaveNumbered(dir, "My Cool Test!!", b)
	require.NoError(t, err)
	assert.Equal(t, "000-my-cool-test-def5678.json", filepath.Base(path))
}
