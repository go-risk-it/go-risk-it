package fileutil_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/loadtest/fileutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSanitizeSlug(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "spaces become hyphens", in: "hello world", want: "hello-world"},
		{name: "special chars stripped", in: "a@b#c!d", want: "abcd"},
		{name: "multiple hyphens collapsed", in: "a---b", want: "a-b"},
		{name: "leading/trailing hyphens trimmed", in: "-hello-", want: "hello"},
		{name: "empty string", in: "", want: ""},
		{name: "already clean", in: "my-slug", want: "my-slug"},
		{name: "uppercase lowered", in: "Hello-World", want: "hello-world"},
		{
			name: "mixed special and spaces",
			in:   "  Some / Complex!! Name  ",
			want: "some-complex-name",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, fileutil.SanitizeSlug(tc.in))
		})
	}
}

func TestNextSequenceNumber(t *testing.T) {
	t.Parallel()

	t.Run("empty dir", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		seq, err := fileutil.NextSequenceNumber(dir)
		require.NoError(t, err)
		assert.Equal(t, 0, seq)
	})

	t.Run("nonexistent dir returns 0", func(t *testing.T) {
		t.Parallel()

		seq, err := fileutil.NextSequenceNumber("/tmp/does-not-exist-fileutil-test")
		require.NoError(t, err)
		assert.Equal(t, 0, seq)
	})

	t.Run("existing numbered files", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		for _, name := range []string{
			"000-baseline-abc123.json",
			"001-staircase-def456.json",
			"002-adaptive-ghi789.json",
		} {
			require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte("{}"), 0o600))
		}

		seq, err := fileutil.NextSequenceNumber(dir)
		require.NoError(t, err)
		assert.Equal(t, 3, seq)
	})

	t.Run("gaps in sequence", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		for _, name := range []string{
			"000-first.json",
			"005-fifth.json",
			"003-third.json",
		} {
			require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte("{}"), 0o600))
		}

		seq, err := fileutil.NextSequenceNumber(dir)
		require.NoError(t, err)
		assert.Equal(t, 6, seq) // highest is 005, next is 006
	})

	t.Run("non-matching files ignored", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		for _, name := range []string{
			"README.md",
			"config.json",
			"002-valid.json",
		} {
			require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte("{}"), 0o600))
		}

		seq, err := fileutil.NextSequenceNumber(dir)
		require.NoError(t, err)
		assert.Equal(t, 3, seq)
	})

	t.Run("directories ignored", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		require.NoError(t, os.Mkdir(filepath.Join(dir, "005-subdir"), 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "001-file.json"), []byte("{}"), 0o600))

		seq, err := fileutil.NextSequenceNumber(dir)
		require.NoError(t, err)
		assert.Equal(t, 2, seq)
	})
}
