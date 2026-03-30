package render_test

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/go-risk-it/go-risk-it/cmd/archdiagram/model"
	. "github.com/go-risk-it/go-risk-it/cmd/archdiagram/render"
)

// taxonomyRepoRoot returns the repository root by navigating up from this test file.
func taxonomyRepoRoot(t *testing.T) string {
	t.Helper()

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to get caller info")
	}

	return filepath.Join(filepath.Dir(thisFile), "..", "..", "..")
}

func TestCheckTaxonomyConsistency_LiveCodebase(t *testing.T) {
	t.Parallel()

	root := taxonomyRepoRoot(t)

	// Use the same layer set as main.go.
	layers := map[string]*model.LayerInfo{
		"API":           {Name: "API", Color: "#E8EAF6", Order: 0},
		"Kernel":        {Name: "Kernel", Color: "#F3E5F5", Order: 1},
		"Data":          {Name: "Data", Color: "#FFF3E0", Order: 2},
		"Events-domain": {Name: "Events-domain", Color: "#FCE4EC", Order: 3},
		"Game-domain":   {Name: "Game-domain", Color: "#E0F7FA", Order: 4},
		"Game-support":  {Name: "Game-support", Color: "#FFF9C4", Order: 5},
		"Lobby-domain":  {Name: "Lobby-domain", Color: "#E0F7FA", Order: 6},
		"Logic":         {Name: "Logic", Color: "#E8F5E9", Order: 7},
		"Web":           {Name: "Web", Color: "#E3F2FD", Order: 8},
		"Test":          {Name: "Test", Color: "#F5F5F5", Order: 9},
	}

	err := CheckTaxonomyConsistency(root, layers)
	if err != nil {
		t.Errorf("CheckTaxonomyConsistency() error: %v", err)
	}
}

func TestCheckTaxonomyConsistency_MissingInDoc(t *testing.T) {
	t.Parallel()

	root := taxonomyRepoRoot(t)

	// Add an extra layer that doesn't exist in doc-go-spec.md.
	layers := map[string]*model.LayerInfo{
		"API":           {Name: "API"},
		"Kernel":        {Name: "Kernel"},
		"Data":          {Name: "Data"},
		"Events-domain": {Name: "Events-domain"},
		"Game-domain":   {Name: "Game-domain"},
		"Game-support":  {Name: "Game-support"},
		"Lobby-domain":  {Name: "Lobby-domain"},
		"Logic":         {Name: "Logic"},
		"Web":           {Name: "Web"},
		"Test":          {Name: "Test"},
		"NewLayer":      {Name: "NewLayer"},
	}

	err := CheckTaxonomyConsistency(root, layers)
	if err == nil {
		t.Error("expected error for extra generator layer, got nil")
	}

	if !strings.Contains(err.Error(), "newlayer") {
		t.Errorf("error should mention 'newlayer', got: %v", err)
	}
}

func TestCheckTaxonomyConsistency_MissingInGenerator(t *testing.T) {
	t.Parallel()

	root := taxonomyRepoRoot(t)

	// Use a subset of layers — missing some that are in doc-go-spec.md.
	layers := map[string]*model.LayerInfo{
		"API":    {Name: "API"},
		"Kernel": {Name: "Kernel"},
	}

	err := CheckTaxonomyConsistency(root, layers)
	if err == nil {
		t.Error("expected error for missing generator layers, got nil")
	}

	if !strings.Contains(err.Error(), "in doc-go-spec.md but not in generator") {
		t.Errorf("error should mention doc layers missing from generator, got: %v", err)
	}
}

func TestCheckTaxonomyConsistency_MissingFile(t *testing.T) {
	t.Parallel()

	layers := map[string]*model.LayerInfo{
		"API": {Name: "API"},
	}

	err := CheckTaxonomyConsistency("/nonexistent/path", layers)
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestExtractTaxonomyLayers(t *testing.T) {
	t.Parallel()

	content := `# Doc.go Specification

## Layer Taxonomy

| Layer | Description | Import constraints |
|-------|-------------|-------------------|
| ` + "`api`" + ` | DTOs | May only import other api packages |
| ` + "`kernel`" + ` | Shared infrastructure | Never imports game |
| ` + "`wiring`" + ` | fx Module roots | One file only |

## Next Section
`

	layers := ExtractTaxonomyLayersForTest(content)

	if !layers["api"] {
		t.Error("expected 'api' in taxonomy layers")
	}

	if !layers["kernel"] {
		t.Error("expected 'kernel' in taxonomy layers")
	}

	if !layers["wiring"] {
		t.Error("expected 'wiring' in taxonomy layers")
	}

	if len(layers) != 3 {
		t.Errorf("expected 3 layers, got %d", len(layers))
	}
}

func TestCompareLayers_WiringExcluded(t *testing.T) {
	t.Parallel()

	// doc has "api" + "wiring", generator has "api" only.
	// wiring is excluded from comparison, so this should pass.
	docLayers := map[string]bool{"api": true, "wiring": true}
	genLayers := map[string]bool{"api": true}

	err := CompareLayersForTest(docLayers, genLayers)
	if err != nil {
		t.Errorf("expected no error (wiring excluded), got: %v", err)
	}
}

func TestCompareLayers_Match(t *testing.T) {
	t.Parallel()

	layers := map[string]bool{"api": true, "kernel": true, "logic": true}

	err := CompareLayersForTest(layers, layers)
	if err != nil {
		t.Errorf("expected no error for matching sets, got: %v", err)
	}
}

func TestCompareLayers_Mismatch(t *testing.T) {
	t.Parallel()

	doc := map[string]bool{"api": true, "kernel": true}
	gen := map[string]bool{"api": true, "logic": true}

	err := CompareLayersForTest(doc, gen)
	if err == nil {
		t.Error("expected error for mismatching sets, got nil")
	}

	errStr := err.Error()

	if !strings.Contains(errStr, "logic") {
		t.Errorf("error should mention 'logic', got: %v", err)
	}

	if !strings.Contains(errStr, "kernel") {
		t.Errorf("error should mention 'kernel', got: %v", err)
	}
}
