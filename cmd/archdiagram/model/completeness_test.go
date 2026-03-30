package model_test

import (
	"encoding/json"
	"os/exec"
	"sort"
	"strings"
	"testing"

	"github.com/go-risk-it/go-risk-it/cmd/archdiagram/docparser"
	"github.com/go-risk-it/go-risk-it/cmd/archdiagram/model"
)

// completenessLayers defines the architectural layer taxonomy matching the live
// doc.go spec. This must stay in sync with the layers in main.go and the doc.go
// files. If a new layer is added to doc.go files, add it here.
//
//nolint:gochecknoglobals // test-only layer definition
var completenessLayers = map[string]*model.LayerInfo{
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

// packagesWithoutDocGo lists non-excluded packages that intentionally have no
// doc.go file. These are excluded from layer and summary assertions because
// the classifier returns empty strings for them, causing BuildModel to skip them.
//
//nolint:gochecknoglobals // test-only exclusion set
var packagesWithoutDocGo = map[string]bool{
	"game/consumers/converter": true,
}

// loadLivePackages runs `go list -json ./internal/...` against the real codebase
// and returns parsed GoPackage values. Requires the codebase to compile.
func loadLivePackages(t *testing.T) []model.GoPackage {
	t.Helper()

	cmd := exec.CommandContext(t.Context(), "go", "list", "-json", "./internal/...")
	cmd.Dir = "../../.." // from cmd/archdiagram/model/ to project root
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("go list failed (codebase must compile): %v", err)
	}

	var pkgs []model.GoPackage

	dec := json.NewDecoder(strings.NewReader(string(out)))
	for dec.More() {
		var pkg model.GoPackage
		if err := dec.Decode(&pkg); err != nil {
			t.Fatalf("failed to decode go list output: %v", err)
		}

		pkgs = append(pkgs, pkg)
	}

	return pkgs
}

// docGoClassifier returns a model.Classifier that reads doc.go files from the
// live codebase. It resolves package suffixes to filesystem directories using
// the GoPackage Dir field from go list output.
func docGoClassifier(t *testing.T, pkgs []model.GoPackage) model.Classifier {
	t.Helper()

	dirMap := make(map[string]string) // suffix -> directory
	for _, pkg := range pkgs {
		if !strings.HasPrefix(pkg.ImportPath, model.ModulePrefix) {
			continue
		}

		suffix := model.PackageSuffix(pkg.ImportPath)
		dirMap[suffix] = pkg.Dir
	}

	return func(suffix string) (string, string) {
		dir, ok := dirMap[suffix]
		if !ok {
			return "", ""
		}

		layer, summary, err := docparser.ParseLayerAndSummary(dir)
		if err != nil {
			t.Errorf("ParseLayerAndSummary(%s): %v", suffix, err)

			return "", ""
		}

		return layer, summary
	}
}

// buildLiveModel runs the full pipeline against the real codebase:
// go list -> parse doc.go -> BuildModel -> GroupPackages.
func buildLiveModel(t *testing.T) (*model.ArchModel, []model.GoPackage) {
	t.Helper()

	pkgs := loadLivePackages(t)
	classify := docGoClassifier(t, pkgs)
	archModel := model.BuildModel(pkgs, classify, completenessLayers)
	model.GroupPackages(archModel)

	return archModel, pkgs
}

// nonExcludedSuffixes returns all package suffixes from go list output that
// are not excluded by model.IsExcluded and not in the packagesWithoutDocGo set.
func nonExcludedSuffixes(pkgs []model.GoPackage) []string {
	var result []string

	for _, pkg := range pkgs {
		if !strings.HasPrefix(pkg.ImportPath, model.ModulePrefix) {
			continue
		}

		suffix := model.PackageSuffix(pkg.ImportPath)
		if model.IsExcluded(suffix) {
			continue
		}

		if packagesWithoutDocGo[suffix] {
			continue
		}

		result = append(result, suffix)
	}

	sort.Strings(result)

	return result
}

// TestCompleteness_AllPackagesClassified verifies that every non-excluded package
// gets a layer assignment from its doc.go file. Packages without doc.go
// (listed in packagesWithoutDocGo) are excluded from this check.
func TestCompleteness_AllPackagesClassified(t *testing.T) {
	t.Parallel()

	archModel, pkgs := buildLiveModel(t)
	expected := nonExcludedSuffixes(pkgs)

	var unclassified []string

	for _, suffix := range expected {
		if _, ok := archModel.Packages[suffix]; !ok {
			unclassified = append(unclassified, suffix)
		}
	}

	if len(unclassified) > 0 {
		t.Errorf(
			"%d packages have no layer assignment (missing or invalid doc.go # Layer):",
			len(unclassified))

		for _, s := range unclassified {
			t.Errorf("  %s", s)
		}
	}

	t.Logf("classified %d / %d non-excluded packages",
		len(archModel.Packages), len(expected))
}

// TestCompleteness_AllPackagesGrouped verifies that every package in the model
// is assigned to exactly one subsystem by GroupPackages.
func TestCompleteness_AllPackagesGrouped(t *testing.T) {
	t.Parallel()

	archModel, _ := buildLiveModel(t)

	// Collect all assigned suffixes across all subsystems.
	assigned := make(map[string]string) // suffix -> subsystem ID

	for subsystemID, sub := range archModel.Subsystems {
		for _, pkg := range sub.Packages {
			if prev, ok := assigned[pkg]; ok {
				t.Errorf("package %q assigned to both %q and %q",
					pkg, prev, subsystemID)
			}

			assigned[pkg] = subsystemID
		}
	}

	var ungrouped []string

	for suffix := range archModel.Packages {
		if _, ok := assigned[suffix]; !ok {
			ungrouped = append(ungrouped, suffix)
		}
	}

	sort.Strings(ungrouped)

	if len(ungrouped) > 0 {
		t.Errorf("%d packages not assigned to any subsystem:", len(ungrouped))

		for _, s := range ungrouped {
			t.Errorf("  %s", s)
		}
	}

	t.Logf("grouped %d / %d packages into %d subsystems",
		len(assigned), len(archModel.Packages), len(archModel.Subsystems))
}

// TestCompleteness_ModelHasSummaries verifies that the model carries non-empty
// summaries for packages that have doc.go files with # Layer headings.
func TestCompleteness_ModelHasSummaries(t *testing.T) {
	t.Parallel()

	archModel, _ := buildLiveModel(t)

	withSummary := 0

	var missingSummary []string

	for suffix, pkg := range archModel.Packages {
		if pkg.Summary != "" {
			withSummary++
		} else {
			missingSummary = append(missingSummary, suffix)
		}
	}

	sort.Strings(missingSummary)

	// At least 70 packages should have summaries (the ~72 non-excluded packages
	// with doc.go). Allow a small buffer for edge cases.
	const minSummaries = 60
	if withSummary < minSummaries {
		t.Errorf("only %d packages have summaries (minimum: %d)",
			withSummary, minSummaries)
	}

	t.Logf("summaries: %d / %d packages have non-empty summaries",
		withSummary, len(archModel.Packages))

	if len(missingSummary) > 0 {
		t.Logf("packages without summaries:")

		for _, s := range missingSummary {
			t.Logf("  %s", s)
		}
	}
}

// TestCompleteness_EdgesExist verifies that the model has a meaningful number
// of cross-layer dependency edges. A healthy architecture with multiple layers
// will naturally have several cross-layer edges.
func TestCompleteness_EdgesExist(t *testing.T) {
	t.Parallel()

	archModel, _ := buildLiveModel(t)

	const minEdges = 5
	if len(archModel.Edges) < minEdges {
		t.Errorf("only %d cross-layer edges (minimum: %d)",
			len(archModel.Edges), minEdges)
	}

	t.Logf("cross-layer edges: %d", len(archModel.Edges))

	// Log the actual edges for diagnostic purposes.
	for _, e := range archModel.Edges {
		t.Logf("  %s -> %s", e.From, e.To)
	}
}

// TestCompleteness_ModelFieldsForRenderers verifies that the ArchModel has all
// fields populated that renderers need: Packages, Subsystems, Layers, and Edges.
// This ensures new renderers (D2, Mermaid, tree, tables) have a complete model.
func TestCompleteness_ModelFieldsForRenderers(t *testing.T) {
	t.Parallel()

	archModel, _ := buildLiveModel(t)

	if len(archModel.Packages) == 0 {
		t.Error("ArchModel.Packages is empty — renderers need package data")
	}

	if len(archModel.Subsystems) == 0 {
		t.Error("ArchModel.Subsystems is empty — renderers need subsystem groupings")
	}

	if len(archModel.Layers) == 0 {
		t.Error("ArchModel.Layers is empty — renderers need layer definitions")
	}

	if len(archModel.Edges) == 0 {
		t.Error("ArchModel.Edges is empty — renderers need cross-layer edges")
	}

	assertPackageFieldsPopulated(t, archModel)
	assertSubsystemFieldsPopulated(t, archModel)
	assertLayerReferencesValid(t, archModel)

	t.Logf("model fields: %d packages, %d subsystems, %d layers, %d edges",
		len(archModel.Packages), len(archModel.Subsystems),
		len(archModel.Layers), len(archModel.Edges))
}

// assertPackageFieldsPopulated verifies every package has required fields.
func assertPackageFieldsPopulated(t *testing.T, archModel *model.ArchModel) {
	t.Helper()

	for suffix, pkg := range archModel.Packages {
		if pkg.ImportPath == "" {
			t.Errorf("package %q has empty ImportPath", suffix)
		}

		if pkg.Suffix == "" {
			t.Errorf("package %q has empty Suffix", suffix)
		}

		if pkg.Layer == "" {
			t.Errorf("package %q has empty Layer", suffix)
		}
	}
}

// assertSubsystemFieldsPopulated verifies every subsystem has required fields.
func assertSubsystemFieldsPopulated(t *testing.T, archModel *model.ArchModel) {
	t.Helper()

	for subsystemID, sub := range archModel.Subsystems {
		if sub.ID == "" {
			t.Errorf("subsystem %q has empty ID", subsystemID)
		}

		if sub.Label == "" {
			t.Errorf("subsystem %q has empty Label", subsystemID)
		}

		if len(sub.Packages) == 0 {
			t.Errorf("subsystem %q has no packages", subsystemID)
		}
	}
}

// assertLayerReferencesValid verifies all layer names referenced by packages
// exist in the layers map.
func assertLayerReferencesValid(t *testing.T, archModel *model.ArchModel) {
	t.Helper()

	for suffix, pkg := range archModel.Packages {
		if _, ok := archModel.Layers[pkg.Layer]; !ok {
			t.Errorf("package %q references layer %q not in Layers map",
				suffix, pkg.Layer)
		}
	}
}

// TestCompleteness_LayersCoverAllDocGoLayers verifies that the completenessLayers
// map includes every layer name found in doc.go files across the codebase.
func TestCompleteness_LayersCoverAllDocGoLayers(t *testing.T) {
	t.Parallel()

	archModel, _ := buildLiveModel(t)

	// Collect all unique layer names from classified packages.
	layers := make(map[string]bool)
	for _, pkg := range archModel.Packages {
		layers[pkg.Layer] = true
	}

	var missingLayers []string

	for layer := range layers {
		if _, ok := completenessLayers[layer]; !ok {
			missingLayers = append(missingLayers, layer)
		}
	}

	sort.Strings(missingLayers)

	if len(missingLayers) > 0 {
		t.Errorf("layers found in doc.go but missing from completenessLayers:")

		for _, l := range missingLayers {
			t.Errorf("  %s", l)
		}
	}

	t.Logf(
		"layers in use: %d distinct layers across %d packages",
		len(layers),
		len(archModel.Packages),
	)
}

// TestCompleteness_NoDocGoExceptionsAreClassified verifies that packages listed
// in packagesWithoutDocGo are genuinely unclassifiable (no doc.go or no # Layer).
// If someone adds a doc.go to one of these packages, this test will catch it
// and prompt removal from the exception list.
func TestCompleteness_NoDocGoExceptionsAreClassified(t *testing.T) {
	t.Parallel()

	pkgs := loadLivePackages(t)
	classify := docGoClassifier(t, pkgs)

	for suffix := range packagesWithoutDocGo {
		layer, _ := classify(suffix)
		if layer != "" {
			t.Errorf(
				"package %q is in packagesWithoutDocGo but now has layer %q — remove from exception list",
				suffix,
				layer,
			)
		}
	}
}
