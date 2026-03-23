package journal_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/journal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveEntry_CreatesFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	entry := journal.Entry{CommitSHA: "abc123", Timestamp: time.Now()}
	path, err := journal.SaveEntry(dir, "test-run", entry)
	require.NoError(t, err)
	assert.FileExists(t, path)
	assert.Contains(t, filepath.Base(path), "000-test-run-abc123")
}

func TestSaveEntry_SequentialNumbering(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	entry := journal.Entry{CommitSHA: "abc123"}

	path1, err := journal.SaveEntry(dir, "run", entry)
	require.NoError(t, err)
	assert.Contains(t, filepath.Base(path1), "000-")

	path2, err := journal.SaveEntry(dir, "run", entry)
	require.NoError(t, err)
	assert.Contains(t, filepath.Base(path2), "001-")
}

func TestSaveEntry_SlugSanitization(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	entry := journal.Entry{CommitSHA: "abc123"}

	path, err := journal.SaveEntry(dir, "My Test!", entry)
	require.NoError(t, err)
	assert.Contains(t, filepath.Base(path), "000-my-test-abc123")
}

func TestLoadEntry_Roundtrip(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	original := journal.Entry{
		CommitSHA: "def456",
		Timestamp: time.Date(2026, 3, 23, 10, 0, 0, 0, time.UTC),
		Config: journal.StaircaseParams{
			Steps:           []int{5, 10, 20},
			HoldDurationSec: 60,
			NumPlayers:      4,
		},
		SLOCeiling: journal.SLOCeiling{Games: 10, ThroughputMPS: 50.5},
	}

	path, err := journal.SaveEntry(dir, "roundtrip", original)
	require.NoError(t, err)

	loaded, err := journal.LoadEntry(path)
	require.NoError(t, err)
	assert.Equal(t, original.CommitSHA, loaded.CommitSHA)
	assert.Equal(t, original.Config.Steps, loaded.Config.Steps)
	assert.Equal(t, original.SLOCeiling.Games, loaded.SLOCeiling.Games)
	assert.InDelta(t, original.SLOCeiling.ThroughputMPS, loaded.SLOCeiling.ThroughputMPS, 0.01)
}

func TestLoadEntry_NonExistent(t *testing.T) {
	t.Parallel()

	_, err := journal.LoadEntry("/nonexistent/path.json")
	require.Error(t, err)
}

func TestListEntries_Empty(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	entries, err := journal.ListEntries(dir)
	require.NoError(t, err)
	assert.Empty(t, entries)
}

func TestListEntries_Sorted(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	entry := journal.Entry{CommitSHA: "abc123"}

	_, err := journal.SaveEntry(dir, "first", entry)
	require.NoError(t, err)

	_, err = journal.SaveEntry(dir, "second", entry)
	require.NoError(t, err)

	_, err = journal.SaveEntry(dir, "third", entry)
	require.NoError(t, err)

	paths, err := journal.ListEntries(dir)
	require.NoError(t, err)
	assert.Len(t, paths, 3)
	assert.Contains(t, filepath.Base(paths[0]), "000-")
	assert.Contains(t, filepath.Base(paths[1]), "001-")
	assert.Contains(t, filepath.Base(paths[2]), "002-")
}

func TestLatestEntry_ReturnsHighest(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	for i, sha := range []string{"aaa", "bbb", "ccc"} {
		entry := journal.Entry{
			CommitSHA: sha,
			SLOCeiling: journal.SLOCeiling{
				Games: (i + 1) * 10,
			},
		}

		_, err := journal.SaveEntry(dir, "run", entry)
		require.NoError(t, err)
	}

	latest, err := journal.LatestEntry(dir)
	require.NoError(t, err)
	assert.Equal(t, "ccc", latest.CommitSHA)
	assert.Equal(t, 30, latest.SLOCeiling.Games)
}

func TestLatestEntry_EmptyDir(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	_, err := journal.LatestEntry(dir)
	require.Error(t, err)
}
