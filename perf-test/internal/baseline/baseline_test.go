package baseline_test

import (
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
			E2EP95:   0.312,
			DBTxnP95: 0.042,
		},
	}

	path, err := baseline.Save(dir, baselineData)
	require.NoError(t, err)
	assert.FileExists(t, path)

	loaded, err := baseline.Load(path)
	require.NoError(t, err)
	assert.Equal(t, baselineData.CommitSHA, loaded.CommitSHA)
	assert.InDelta(t, baselineData.Metrics.E2EP95, loaded.Metrics.E2EP95, 0.001)
	assert.InDelta(t, baselineData.Metrics.DBTxnP95, loaded.Metrics.DBTxnP95, 0.001)
	assert.Equal(t, baselineData.TestParams.Preset, loaded.TestParams.Preset)
}

func TestBaseline_Load_NonExistentFile(t *testing.T) {
	t.Parallel()

	_, err := baseline.Load("/nonexistent/path.json")
	require.Error(t, err)
}
