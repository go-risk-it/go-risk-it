package model_test

import (
	"slices"
	"testing"

	. "github.com/go-risk-it/go-risk-it/cmd/archdiagram/model"
)

const testLayerLogic = "Logic"

// testLayers provides a minimal layer map for test use.
func testLayers() map[string]*LayerInfo {
	return map[string]*LayerInfo{
		testLayerLogic:   {Name: testLayerLogic, Color: "#E8F5E9", Order: 0},
		"Web":            {Name: "Web", Color: "#E3F2FD", Order: 1},
		"Data":           {Name: "Data", Color: "#FFF3E0", Order: 2},
		"Infrastructure": {Name: "Infrastructure", Color: "#F3E5F5", Order: 3},
		"Events":         {Name: "Events", Color: "#FCE4EC", Order: 4},
	}
}

// testClassifier is a mock classifier that maps known suffixes to layers.
func testClassifier(suffix string) (string, string) {
	switch suffix {
	case "logic/game/board":
		return testLayerLogic, "Board management"
	case "logic/game/phase":
		return testLayerLogic, "Phase transitions"
	case "logic/game/move/attack":
		return testLayerLogic, "Attack moves"
	case "web/game/controller":
		return "Web", "Game HTTP handlers"
	case "web/game/rest":
		return "Web", "REST utilities"
	case "data/game/db":
		return "Data", "Game data access"
	case "config":
		return "Infrastructure", "Configuration"
	case "events":
		return "Events", "Event bus"
	default:
		return "", ""
	}
}

func TestBuildModel_ExcludesWiringRoots(t *testing.T) {
	t.Parallel()

	pkgs := []GoPackage{
		// Wiring roots — should be excluded
		{ImportPath: ModulePrefix},                             // empty suffix ""
		{ImportPath: ModulePrefix + "logic"},                   // wiring root
		{ImportPath: ModulePrefix + "logic/game"},              // wiring root
		{ImportPath: ModulePrefix + "logic/game/move"},         // wiring root
		{ImportPath: ModulePrefix + "logic/game/move/service"}, // wiring root
		{ImportPath: ModulePrefix + "data"},                    // wiring root
		{ImportPath: ModulePrefix + "data/game"},               // wiring root
		{ImportPath: ModulePrefix + "web"},                     // wiring root
		{ImportPath: ModulePrefix + "web/game"},                // wiring root
		{ImportPath: ModulePrefix + "web/lobby"},               // wiring root
		{ImportPath: ModulePrefix + "lobby/logic"},             // wiring root
		{ImportPath: ModulePrefix + "lobby/data"},              // wiring root
		{ImportPath: ModulePrefix + "data/game/sqlc"},          // sqlc — excluded
		{ImportPath: ModulePrefix + "lobby/data/sqlc"},         // sqlc — excluded
		{ImportPath: ModulePrefix + "something/mocks"},         // mocks — excluded
		{
			ImportPath: ModulePrefix + "logic/game/board",
		}, // real package — should be included
		{
			ImportPath: ModulePrefix + "web/game/controller",
		}, // real package — should be included
		{
			ImportPath: "fmt",
		}, // external — no module prefix, ignored
		{
			ImportPath: "github.com/other/module/internal/whatever",
		}, // external — no module prefix, ignored
	}

	archModel := BuildModel(pkgs, testClassifier, testLayers())

	// Only the two real packages should be in the model.
	if len(archModel.Packages) != 2 {
		t.Errorf("expected 2 packages, got %d", len(archModel.Packages))

		for suffix := range archModel.Packages {
			t.Logf("  found: %s", suffix)
		}
	}

	if _, ok := archModel.Packages["logic/game/board"]; !ok {
		t.Error("expected logic/game/board to be included")
	}

	if _, ok := archModel.Packages["web/game/controller"]; !ok {
		t.Error("expected web/game/controller to be included")
	}

	// Verify excluded packages are absent.
	excluded := []string{
		"", "logic", "logic/game", "logic/game/move", "logic/game/move/service",
		"data", "data/game", "web", "web/game", "web/lobby",
		"lobby/logic", "lobby/data",
		"data/game/sqlc", "lobby/data/sqlc", "something/mocks",
	}
	for _, suffix := range excluded {
		if _, ok := archModel.Packages[suffix]; ok {
			t.Errorf("expected %q to be excluded, but found it", suffix)
		}
	}
}

func TestBuildModel_ClassifiesLayers(t *testing.T) {
	t.Parallel()

	pkgs := []GoPackage{
		{ImportPath: ModulePrefix + "logic/game/board"},
		{ImportPath: ModulePrefix + "logic/game/phase"},
		{ImportPath: ModulePrefix + "web/game/controller"},
		{ImportPath: ModulePrefix + "data/game/db"},
		{ImportPath: ModulePrefix + "config"},
		{ImportPath: ModulePrefix + "events"},
	}

	archModel := BuildModel(pkgs, testClassifier, testLayers())

	cases := []struct {
		suffix    string
		wantLayer string
		wantSumm  string
	}{
		{"logic/game/board", "Logic", "Board management"},
		{"logic/game/phase", "Logic", "Phase transitions"},
		{"web/game/controller", "Web", "Game HTTP handlers"},
		{"data/game/db", "Data", "Game data access"},
		{"config", "Infrastructure", "Configuration"},
		{"events", "Events", "Event bus"},
	}

	for _, testCase := range cases {
		t.Run(testCase.suffix, func(t *testing.T) {
			t.Parallel()

			pkg, ok := archModel.Packages[testCase.suffix]
			if !ok {
				t.Fatalf("package %q not found in model", testCase.suffix)
			}

			if pkg.Layer != testCase.wantLayer {
				t.Errorf("layer = %q, want %q", pkg.Layer, testCase.wantLayer)
			}

			if pkg.Summary != testCase.wantSumm {
				t.Errorf("summary = %q, want %q", pkg.Summary, testCase.wantSumm)
			}
		})
	}
}

func TestBuildModel_SkipsUnclassifiedPackages(t *testing.T) {
	t.Parallel()

	pkgs := []GoPackage{
		{ImportPath: ModulePrefix + "logic/game/board"},               // classified
		{ImportPath: ModulePrefix + "something/new/and/unclassified"}, // not in classifier
	}

	archModel := BuildModel(pkgs, testClassifier, testLayers())

	if len(archModel.Packages) != 1 {
		t.Errorf("expected 1 package, got %d", len(archModel.Packages))
	}

	if _, ok := archModel.Packages["logic/game/board"]; !ok {
		t.Error("expected logic/game/board to be included")
	}

	if _, ok := archModel.Packages["something/new/and/unclassified"]; ok {
		t.Error("expected unclassified package to be excluded")
	}
}

func TestBuildModel_CrossLayerEdges(t *testing.T) {
	t.Parallel()

	pkgs := []GoPackage{
		{
			ImportPath: ModulePrefix + "web/game/controller",
			Imports:    []string{ModulePrefix + "logic/game/board", "fmt"},
		},
		{
			ImportPath: ModulePrefix + "logic/game/board",
			Imports:    []string{ModulePrefix + "data/game/db"},
		},
		{
			ImportPath: ModulePrefix + "data/game/db",
		},
	}

	archModel := BuildModel(pkgs, testClassifier, testLayers())

	// Expect two cross-layer edges: Web→Logic, Logic→Data
	if len(archModel.Edges) != 2 {
		t.Errorf("expected 2 cross-layer edges, got %d", len(archModel.Edges))

		for _, e := range archModel.Edges {
			t.Logf("  edge: %s -> %s", e.From, e.To)
		}
	}

	webToLogic := Edge{From: "Web", To: "Logic"}
	logicToData := Edge{From: "Logic", To: "Data"}

	if !slices.Contains(archModel.Edges, webToLogic) {
		t.Errorf("expected edge Web -> Logic, not found")
	}

	if !slices.Contains(archModel.Edges, logicToData) {
		t.Errorf("expected edge Logic -> Data, not found")
	}
}

func TestBuildModel_SameLayerEdgesExcluded(t *testing.T) {
	t.Parallel()

	pkgs := []GoPackage{
		{
			ImportPath: ModulePrefix + "logic/game/board",
			Imports:    []string{ModulePrefix + "logic/game/phase"},
		},
		{
			ImportPath: ModulePrefix + "logic/game/phase",
		},
	}

	archModel := BuildModel(pkgs, testClassifier, testLayers())

	if len(archModel.Edges) != 0 {
		t.Errorf("expected 0 edges for same-layer imports, got %d", len(archModel.Edges))

		for _, e := range archModel.Edges {
			t.Logf("  edge: %s -> %s", e.From, e.To)
		}
	}
}

func TestBuildModel_DeduplicatesEdges(t *testing.T) {
	t.Parallel()

	// Two Web packages both import a Logic package — should produce one edge.
	pkgs := []GoPackage{
		{
			ImportPath: ModulePrefix + "web/game/controller",
			Imports:    []string{ModulePrefix + "logic/game/board"},
		},
		{
			ImportPath: ModulePrefix + "web/game/rest",
			Imports:    []string{ModulePrefix + "logic/game/phase"},
		},
		{ImportPath: ModulePrefix + "logic/game/board"},
		{ImportPath: ModulePrefix + "logic/game/phase"},
	}

	archModel := BuildModel(pkgs, testClassifier, testLayers())

	if len(archModel.Edges) != 1 {
		t.Errorf("expected 1 deduplicated edge, got %d", len(archModel.Edges))

		for _, e := range archModel.Edges {
			t.Logf("  edge: %s -> %s", e.From, e.To)
		}
	}

	if len(archModel.Edges) > 0 {
		if archModel.Edges[0] != (Edge{From: "Web", To: "Logic"}) {
			t.Errorf("expected Web -> Logic, got %s -> %s",
				archModel.Edges[0].From, archModel.Edges[0].To)
		}
	}
}

func TestBuildModel_InternalDeps(t *testing.T) {
	t.Parallel()

	pkgs := []GoPackage{
		{
			ImportPath: ModulePrefix + "web/game/controller",
			Imports: []string{
				ModulePrefix + "logic/game/board",
				ModulePrefix + "data/game/db",
				"fmt",
				"net/http",
			},
		},
		{ImportPath: ModulePrefix + "logic/game/board"},
		{ImportPath: ModulePrefix + "data/game/db"},
	}

	archModel := BuildModel(pkgs, testClassifier, testLayers())

	pkg := archModel.Packages["web/game/controller"]
	if pkg == nil {
		t.Fatal("web/game/controller not found")
	}

	// Should have internal deps on the two internal packages, not fmt/net/http.
	if len(pkg.InternalDeps) != 2 {
		t.Errorf("expected 2 internal deps, got %d: %v", len(pkg.InternalDeps), pkg.InternalDeps)
	}

	if !slices.Contains(pkg.InternalDeps, "logic/game/board") {
		t.Error("expected internal dep on logic/game/board")
	}

	if !slices.Contains(pkg.InternalDeps, "data/game/db") {
		t.Error("expected internal dep on data/game/db")
	}
}

func TestBuildModel_LayersPassedThrough(t *testing.T) {
	t.Parallel()

	layers := testLayers()
	archModel := BuildModel(nil, testClassifier, layers)

	if len(archModel.Layers) != len(layers) {
		t.Errorf("expected %d layers, got %d", len(layers), len(archModel.Layers))
	}

	for name, info := range layers {
		got, ok := archModel.Layers[name]
		if !ok {
			t.Errorf("layer %q missing from model", name)

			continue
		}

		if got.Name != info.Name || got.Color != info.Color || got.Order != info.Order {
			t.Errorf("layer %q mismatch: got %+v, want %+v", name, got, info)
		}
	}
}

func TestBuildModel_EmptyInput(t *testing.T) {
	t.Parallel()

	archModel := BuildModel(nil, testClassifier, testLayers())

	if len(archModel.Packages) != 0 {
		t.Errorf("expected 0 packages, got %d", len(archModel.Packages))
	}

	if len(archModel.Edges) != 0 {
		t.Errorf("expected 0 edges, got %d", len(archModel.Edges))
	}

	if len(archModel.Subsystems) != 0 {
		t.Errorf("expected 0 subsystems, got %d", len(archModel.Subsystems))
	}
}

func TestBuildModel_EdgeToExcludedTarget(t *testing.T) {
	t.Parallel()

	// A real package imports a wiring root and an sqlc package.
	// Neither should produce edges (targets are excluded/unclassified).
	pkgs := []GoPackage{
		{
			ImportPath: ModulePrefix + "logic/game/board",
			Imports: []string{
				ModulePrefix + "logic",          // wiring root
				ModulePrefix + "data/game/sqlc", // sqlc
			},
		},
	}

	archModel := BuildModel(pkgs, testClassifier, testLayers())

	if len(archModel.Edges) != 0 {
		t.Errorf("expected 0 edges when targets are excluded, got %d", len(archModel.Edges))

		for _, e := range archModel.Edges {
			t.Logf("  edge: %s -> %s", e.From, e.To)
		}
	}
}

func TestIsExcluded(t *testing.T) {
	t.Parallel()

	cases := []struct {
		suffix string
		want   bool
	}{
		// wiring roots
		{"", true},
		{"kernel", true},
		{"game", true},
		{"game/logic", true},
		{"game/logic/move", true},
		{"game/logic/move/service", true},
		{"game/data", true},
		{"lobby", true},
		{"lobby/logic", true},
		{"lobby/data", true},
		{"web", true},
		// generated packages
		{"game/data/sqlc", true},
		{"lobby/data/sqlc", true},
		{"something/mocks", true},
		// real packages
		{"kernel/config", false},
		{"game/logic/board", false},
		{"game/consumers", false},
		{"game/events", false},
	}

	for _, testCase := range cases {
		t.Run(testCase.suffix, func(t *testing.T) {
			t.Parallel()

			got := IsExcluded(testCase.suffix)
			if got != testCase.want {
				t.Errorf("IsExcluded(%q) = %v, want %v",
					testCase.suffix, got, testCase.want)
			}
		})
	}
}

func TestPackageSuffix(t *testing.T) {
	t.Parallel()

	cases := []struct {
		importPath string
		want       string
	}{
		{ModulePrefix + "logic/game/board", "logic/game/board"},
		{ModulePrefix + "config", "config"},
		{ModulePrefix, ""},
		{"fmt", "fmt"}, // non-module import — unchanged
	}

	for _, testCase := range cases {
		t.Run(testCase.importPath, func(t *testing.T) {
			t.Parallel()

			got := PackageSuffix(testCase.importPath)
			if got != testCase.want {
				t.Errorf("PackageSuffix(%q) = %q, want %q",
					testCase.importPath, got, testCase.want)
			}
		})
	}
}
