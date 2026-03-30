package main

import (
	"sort"
	"strings"
	"testing"

	"github.com/go-risk-it/go-risk-it/cmd/archdiagram/model"
)

const layerLogic = "Logic"

func TestSetSubsystemLayers(t *testing.T) {
	t.Parallel()

	archModel := &model.ArchModel{
		Packages: map[string]*model.PackageInfo{
			"game/logic/board": {Suffix: "game/logic/board", Layer: layerLogic},
			"game/logic/card":  {Suffix: "game/logic/card", Layer: layerLogic},
			"game/logic/phase": {Suffix: "game/logic/phase", Layer: layerLogic},
			"web/game/rest":    {Suffix: "web/game/rest", Layer: "Web"},
		},
		Subsystems: map[string]*model.SubsystemInfo{
			"game_services": {
				ID:       "game_services",
				Label:    "Game Services",
				Packages: []string{"game/logic/board", "game/logic/card", "game/logic/phase"},
			},
			"game_handlers": {
				ID:       "game_handlers",
				Label:    "Game Handlers",
				Packages: []string{"web/game/rest"},
			},
		},
		Layers: map[string]*model.LayerInfo{
			layerLogic: {Name: layerLogic, Color: "#E8F5E9", Order: 0},
			"Web":      {Name: "Web", Color: "#E3F2FD", Order: 1},
		},
	}

	setSubsystemLayers(archModel)

	if archModel.Subsystems["game_services"].Layer != layerLogic {
		t.Errorf("game_services layer = %q, want %q",
			archModel.Subsystems["game_services"].Layer, layerLogic)
	}

	if archModel.Subsystems["game_handlers"].Layer != "Web" {
		t.Errorf("game_handlers layer = %q, want %q",
			archModel.Subsystems["game_handlers"].Layer, "Web")
	}
}

func TestSetSubsystemLayers_MajorityVote(t *testing.T) {
	t.Parallel()

	// Subsystem with 2 Logic and 1 Web package — Logic wins.
	archModel := &model.ArchModel{
		Packages: map[string]*model.PackageInfo{
			"a": {Suffix: "a", Layer: layerLogic},
			"b": {Suffix: "b", Layer: layerLogic},
			"c": {Suffix: "c", Layer: "Web"},
		},
		Subsystems: map[string]*model.SubsystemInfo{
			"mixed": {
				ID:       "mixed",
				Label:    "Mixed",
				Packages: []string{"a", "b", "c"},
			},
		},
		Layers: map[string]*model.LayerInfo{
			layerLogic: {Name: layerLogic, Color: "#E8F5E9", Order: 0},
			"Web":      {Name: "Web", Color: "#E3F2FD", Order: 1},
		},
	}

	setSubsystemLayers(archModel)

	if archModel.Subsystems["mixed"].Layer != layerLogic {
		t.Errorf("mixed layer = %q, want %q",
			archModel.Subsystems["mixed"].Layer, layerLogic)
	}
}

func TestSetSubsystemLayers_TieBreaksAlphabetically(t *testing.T) {
	t.Parallel()

	// 1 Logic, 1 Web — alphabetically "Logic" < "Web" so Logic wins.
	archModel := &model.ArchModel{
		Packages: map[string]*model.PackageInfo{
			"x": {Suffix: "x", Layer: "Web"},
			"y": {Suffix: "y", Layer: layerLogic},
		},
		Subsystems: map[string]*model.SubsystemInfo{
			"tied": {
				ID:       "tied",
				Label:    "Tied",
				Packages: []string{"x", "y"},
			},
		},
		Layers: map[string]*model.LayerInfo{
			layerLogic: {Name: layerLogic, Color: "#E8F5E9", Order: 0},
			"Web":      {Name: "Web", Color: "#E3F2FD", Order: 1},
		},
	}

	setSubsystemLayers(archModel)

	if archModel.Subsystems["tied"].Layer != layerLogic {
		t.Errorf("tied layer = %q, want %q",
			archModel.Subsystems["tied"].Layer, layerLogic)
	}
}

func TestBuildSuffixToDirMap(t *testing.T) {
	t.Parallel()

	pkgs := []model.GoPackage{
		{
			ImportPath: model.ModulePrefix + "game/logic/board",
			Dir:        "/src/internal/game/logic/board",
		},
		{ImportPath: model.ModulePrefix + "config", Dir: "/src/internal/config"},
		{ImportPath: "fmt"}, // external — no module prefix
		{
			ImportPath: "github.com/other/module/internal/pkg",
			Dir:        "/other",
		}, // external — no module prefix
	}

	got := buildSuffixToDirMap(pkgs)

	if len(got) != 2 {
		t.Errorf("expected 2 entries, got %d", len(got))
	}

	if got["game/logic/board"] != "/src/internal/game/logic/board" {
		t.Errorf("game/logic/board = %q, want %q",
			got["game/logic/board"], "/src/internal/game/logic/board")
	}

	if got["config"] != "/src/internal/config" {
		t.Errorf("config = %q, want %q", got["config"], "/src/internal/config")
	}
}

func TestGenerateD2(t *testing.T) {
	t.Parallel()

	archModel := &model.ArchModel{
		Packages: map[string]*model.PackageInfo{
			"game/logic/board": {
				Suffix: "game/logic/board",
				Layer:  "Logic",
			},
		},
		Subsystems: map[string]*model.SubsystemInfo{
			"game_services": {
				ID:       "game_services",
				Label:    "Game Services",
				Layer:    "Logic",
				Packages: []string{"game/logic/board"},
			},
			"database": {
				ID:       "database",
				Label:    "Database",
				Layer:    "Data",
				Packages: []string{"data/db"},
			},
		},
		Layers: map[string]*model.LayerInfo{
			"Logic": {Name: "Logic", Color: "#E8F5E9", Order: 0},
			"Data":  {Name: "Data", Color: "#FFF3E0", Order: 1},
		},
		Edges: []model.Edge{
			{From: "Logic", To: "Data"},
		},
	}

	d2Output := generateD2(archModel)

	// Check title
	if !strings.Contains(d2Output, "title: go-risk-it Architecture") {
		t.Error("expected title in D2 output")
		t.Logf("D2 output:\n%s", d2Output)
	}

	// Check direction
	if !strings.Contains(d2Output, "direction: right") {
		t.Error("expected direction in D2 output")
	}

	// Check layer containers
	if !strings.Contains(d2Output, "logic: Logic {") {
		t.Error("expected Logic container")
		t.Logf("D2 output:\n%s", d2Output)
	}

	if !strings.Contains(d2Output, "data: Data {") {
		t.Error("expected Data container")
		t.Logf("D2 output:\n%s", d2Output)
	}

	// Check subsystem nodes
	if !strings.Contains(d2Output, `game_services: "Game Services"`) {
		t.Error("expected game_services subsystem node")
		t.Logf("D2 output:\n%s", d2Output)
	}

	if !strings.Contains(d2Output, `database: "Database"`) {
		t.Error("expected database subsystem node")
		t.Logf("D2 output:\n%s", d2Output)
	}

	// Check cross-layer edge
	if !strings.Contains(d2Output, "logic -> data") {
		t.Error("expected layer-level edge logic -> data")
		t.Logf("D2 output:\n%s", d2Output)
	}

	// No multi-line labels (old format)
	if strings.Contains(d2Output, "\\n") {
		t.Error("expected no multi-line package labels, but found \\n in output")
		t.Logf("D2 output:\n%s", d2Output)
	}
}

func TestLayerContainerID(t *testing.T) {
	t.Parallel()

	cases := []struct {
		layer string
		want  string
	}{
		{"Logic", "logic"},
		{"Web", "web"},
		{"Events-domain", "events_domain"},
		{"Game-domain", "game_domain"},
		{"Game-support", "game_support"},
		{"Lobby-domain", "lobby_domain"},
		{"Kernel", "kernel"},
		{"Data", "data"},
		{"API", "api"},
		{"Test", "test"},
	}

	for _, testCase := range cases {
		t.Run(testCase.layer, func(t *testing.T) {
			t.Parallel()

			got := layerContainerID(testCase.layer)
			if got != testCase.want {
				t.Errorf(
					"layerContainerID(%q) = %q, want %q",
					testCase.layer,
					got,
					testCase.want,
				)
			}
		})
	}
}

func TestLayerOrdering(t *testing.T) {
	t.Parallel()

	// Verify layers are emitted in order by checking D2 output.
	archModel := &model.ArchModel{
		Packages: map[string]*model.PackageInfo{},
		Subsystems: map[string]*model.SubsystemInfo{
			"api_dtos": {
				ID:       "api_dtos",
				Label:    "API DTOs",
				Layer:    "API",
				Packages: []string{"game/api"},
			},
			"kernel": {
				ID:       "kernel",
				Label:    "Kernel",
				Layer:    "Kernel",
				Packages: []string{"kernel/config"},
			},
			"game_services": {
				ID:       "game_services",
				Label:    "Game Services",
				Layer:    "Logic",
				Packages: []string{"game/logic/board"},
			},
			"game_handlers": {
				ID:       "game_handlers",
				Label:    "Game Handlers",
				Layer:    "Web",
				Packages: []string{"web/game/controller"},
			},
		},
		Layers: layers,
	}

	d2Output := generateD2(archModel)

	// Find positions of layer containers in output
	apiPos := strings.Index(d2Output, "api: API {")
	kernelPos := strings.Index(d2Output, "kernel: Kernel {")
	logicPos := strings.Index(d2Output, "logic: Logic {")
	webPos := strings.Index(d2Output, "web: Web {")

	if apiPos < 0 || kernelPos < 0 || logicPos < 0 || webPos < 0 {
		t.Fatalf("missing layer containers in output:\n%s", d2Output)
	}

	if apiPos >= kernelPos {
		t.Errorf("API (pos %d) should appear before Kernel (pos %d)", apiPos, kernelPos)
	}

	if kernelPos >= logicPos {
		t.Errorf("Kernel (pos %d) should appear before Logic (pos %d)", kernelPos, logicPos)
	}

	if logicPos >= webPos {
		t.Errorf("Logic (pos %d) should appear before Web (pos %d)", logicPos, webPos)
	}
}

func TestLayersMap(t *testing.T) {
	t.Parallel()

	// Verify we have exactly 10 layers with unique orders.
	if len(layers) != 10 {
		t.Errorf("expected 10 layers, got %d", len(layers))
	}

	orders := make(map[int]string)

	for name, info := range layers {
		if prev, ok := orders[info.Order]; ok {
			t.Errorf("duplicate order %d: %q and %q", info.Order, prev, name)
		}

		orders[info.Order] = name

		if info.Name != name {
			t.Errorf("layer %q has Name %q (should match key)", name, info.Name)
		}
	}

	// Verify orders are 0..9
	for i := range 10 {
		if _, ok := orders[i]; !ok {
			t.Errorf("missing layer with order %d", i)
		}
	}
}

func TestMakeClassifier_NoDir(t *testing.T) {
	t.Parallel()

	// When a suffix has no entry in the dir map, classifier returns empty.
	classify := makeClassifier(map[string]string{})

	layer, summary := classify("nonexistent/package")

	if layer != "" {
		t.Errorf("expected empty layer, got %q", layer)
	}

	if summary != "" {
		t.Errorf("expected empty summary, got %q", summary)
	}
}

func TestGenerateD2_NoEdges(t *testing.T) {
	t.Parallel()

	archModel := &model.ArchModel{
		Packages: map[string]*model.PackageInfo{},
		Subsystems: map[string]*model.SubsystemInfo{
			"kernel": {
				ID:       "kernel",
				Label:    "Kernel",
				Layer:    "Kernel",
				Packages: []string{"kernel/config"},
			},
		},
		Layers: map[string]*model.LayerInfo{
			"Kernel": {Name: "Kernel", Color: "#F3E5F5", Order: 1},
		},
	}

	d2Output := generateD2(archModel)

	if strings.Contains(d2Output, "Cross-layer dependencies") {
		t.Error("expected no cross-layer dependencies comment when no edges")
	}

	if !strings.Contains(d2Output, "kernel: Kernel {") {
		t.Error("expected Kernel container")
		t.Logf("D2 output:\n%s", d2Output)
	}
}

func TestGenerateD2_SubsystemsSortedWithinLayer(t *testing.T) {
	t.Parallel()

	archModel := &model.ArchModel{
		Packages: map[string]*model.PackageInfo{},
		Subsystems: map[string]*model.SubsystemInfo{
			"zebra": {
				ID:       "zebra",
				Label:    "Zebra",
				Layer:    "Logic",
				Packages: []string{"game/logic/zebra"},
			},
			"alpha": {
				ID:       "alpha",
				Label:    "Alpha",
				Layer:    "Logic",
				Packages: []string{"game/logic/alpha"},
			},
		},
		Layers: map[string]*model.LayerInfo{
			"Logic": {Name: "Logic", Color: "#E8F5E9", Order: 0},
		},
	}

	d2Output := generateD2(archModel)

	alphaPos := strings.Index(d2Output, `alpha: "Alpha"`)
	zebraPos := strings.Index(d2Output, `zebra: "Zebra"`)

	if alphaPos < 0 || zebraPos < 0 {
		t.Fatalf("missing subsystem nodes:\n%s", d2Output)
	}

	if alphaPos >= zebraPos {
		t.Errorf("alpha (pos %d) should appear before zebra (pos %d)", alphaPos, zebraPos)
	}
}

func TestGenerateD2_EdgesSorted(t *testing.T) {
	t.Parallel()

	archModel := &model.ArchModel{
		Packages:   map[string]*model.PackageInfo{},
		Subsystems: map[string]*model.SubsystemInfo{},
		Layers: map[string]*model.LayerInfo{
			"Web":   {Name: "Web", Color: "#E3F2FD", Order: 0},
			"Logic": {Name: "Logic", Color: "#E8F5E9", Order: 1},
			"Data":  {Name: "Data", Color: "#FFF3E0", Order: 2},
		},
		Edges: []model.Edge{
			{From: "Web", To: "Logic"},
			{From: "Logic", To: "Data"},
			{From: "Web", To: "Data"},
		},
	}

	d2Output := generateD2(archModel)

	lines := strings.Split(d2Output, "\n")

	var edgeLines []string

	for _, line := range lines {
		if strings.Contains(line, " -> ") {
			edgeLines = append(edgeLines, line)
		}
	}

	if !sort.StringsAreSorted(edgeLines) {
		t.Errorf("edges not sorted: %v", edgeLines)
	}
}
