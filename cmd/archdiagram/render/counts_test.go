package render_test

import (
	"path/filepath"
	"runtime"
	"testing"

	. "github.com/go-risk-it/go-risk-it/cmd/archdiagram/render"
)

// repoRoot returns the repository root by navigating up from this test file.
func repoRoot(t *testing.T) string {
	t.Helper()

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to get caller info")
	}

	// This file is at cmd/archdiagram/render/counts_test.go,
	// so repo root is three directories up.
	return filepath.Join(filepath.Dir(thisFile), "..", "..", "..")
}

func TestCountArchRules(t *testing.T) {
	t.Parallel()

	root := repoRoot(t)

	count, err := CountArchRules(root)
	if err != nil {
		t.Fatalf("CountArchRules() error: %v", err)
	}

	if count < 1 {
		t.Errorf("CountArchRules() = %d, expected at least 1", count)
	}

	// Verify the count matches what we know about the codebase.
	// The current arch_test.go has 28 rules. Allow a range to avoid
	// breaking when rules are added.
	if count < 20 {
		t.Errorf("CountArchRules() = %d, expected >= 20 (sanity check)", count)
	}

	t.Logf("CountArchRules() = %d", count)
}

func TestCountArchRules_MissingFile(t *testing.T) {
	t.Parallel()

	_, err := CountArchRules("/nonexistent/path")
	if err == nil {
		t.Error("CountArchRules() expected error for missing file, got nil")
	}
}

func TestCountInvariants(t *testing.T) {
	t.Parallel()

	root := repoRoot(t)

	count, err := CountInvariants(root)
	if err != nil {
		t.Fatalf("CountInvariants() error: %v", err)
	}

	if count < 1 {
		t.Errorf("CountInvariants() = %d, expected at least 1", count)
	}

	// The current invariant.go has 12 invariants.
	if count < 10 {
		t.Errorf("CountInvariants() = %d, expected >= 10 (sanity check)", count)
	}

	t.Logf("CountInvariants() = %d", count)
}

func TestCountInvariants_MissingFile(t *testing.T) {
	t.Parallel()

	_, err := CountInvariants("/nonexistent/path")
	if err == nil {
		t.Error("CountInvariants() expected error for missing file, got nil")
	}
}
