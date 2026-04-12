package docparser_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/go-risk-it/go-risk-it/cmd/archdiagram/docparser"
)

// goldenEntry is a single entry in the golden file mapping package suffix to
// expected parse results.
type goldenEntry struct {
	Layer   string `json:"layer"`
	Summary string `json:"summary"`
}

// testdataDir returns the absolute path to the testdata directory.
func testdataDir(t *testing.T) string {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to determine test file location")
	}

	return filepath.Join(filepath.Dir(filename), "testdata")
}

// internalDir returns the absolute path to the internal/ directory.
func internalDir(t *testing.T) string {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to determine test file location")
	}

	// cmd/archdiagram/docparser/parser_test.go → project root → internal/
	return filepath.Join(filepath.Dir(filename), "..", "..", "..", "internal")
}

// TestGoldenFile verifies that ParseLayerAndSummary produces the expected
// layer and summary for every doc.go file listed in the golden file. If any
// parse result changes, this test fails — providing regression protection.
func TestGoldenFile(t *testing.T) {
	t.Parallel()

	goldenPath := filepath.Join(testdataDir(t), "golden.json")

	data, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("failed to read golden file: %v", err)
	}

	var golden map[string]goldenEntry
	if err := json.Unmarshal(data, &golden); err != nil {
		t.Fatalf("failed to parse golden file: %v", err)
	}

	if len(golden) == 0 {
		t.Fatal("golden file is empty")
	}

	root := internalDir(t)

	for suffix, expected := range golden {
		t.Run(suffix, func(t *testing.T) {
			t.Parallel()

			dir := filepath.Join(root, suffix)

			layer, summary, err := docparser.ParseLayerAndSummary(dir)
			if err != nil {
				t.Fatalf("ParseLayerAndSummary(%q): unexpected error: %v", dir, err)
			}

			if layer != expected.Layer {
				t.Errorf("layer: got %q, want %q", layer, expected.Layer)
			}

			if summary != expected.Summary {
				t.Errorf("summary:\n  got:  %q\n  want: %q", summary, expected.Summary)
			}
		})
	}
}

// TestParseMissing verifies that a directory without doc.go returns empty
// strings and no error.
func TestParseMissing(t *testing.T) {
	t.Parallel()

	// Use a directory without doc.go to verify empty returns.
	dir := t.TempDir()

	layer, summary, err := docparser.ParseLayerAndSummary(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if layer != "" {
		t.Errorf("layer: got %q, want empty", layer)
	}

	if summary != "" {
		t.Errorf("summary: got %q, want empty", summary)
	}
}

// TestParseNoLayer verifies that a doc.go without a # Layer heading returns
// empty strings and no error. A doc.go without # Layer is pre-spec and we
// don't extract any metadata from it.
func TestParseNoLayer(t *testing.T) {
	t.Parallel()

	// Create a temporary directory with a doc.go that has no # Layer heading
	dir := t.TempDir()

	content := `// Package example provides example functionality.
//
// It does useful things.
//
// # Key Types
//
// [Foo] is the primary type.
package example
`

	if err := os.WriteFile(filepath.Join(dir, "doc.go"), []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write test doc.go: %v", err)
	}

	layer, summary, err := docparser.ParseLayerAndSummary(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if layer != "" {
		t.Errorf("layer: got %q, want empty", layer)
	}

	if summary != "" {
		t.Errorf("summary: got %q, want empty", summary)
	}
}

// TestParseIOError verifies that a nonexistent directory returns an error.
func TestParseIOError(t *testing.T) {
	t.Parallel()

	_, _, err := docparser.ParseLayerAndSummary("/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Fatal("expected error for nonexistent directory, got nil")
	}
}

// TestParseEmptyDocGo verifies that a doc.go with no package comment returns
// empty strings and no error.
func TestParseEmptyDocGo(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	content := `package example
`

	if err := os.WriteFile(filepath.Join(dir, "doc.go"), []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write test doc.go: %v", err)
	}

	layer, summary, err := docparser.ParseLayerAndSummary(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if layer != "" {
		t.Errorf("layer: got %q, want empty", layer)
	}

	if summary != "" {
		t.Errorf("summary: got %q, want empty", summary)
	}
}
