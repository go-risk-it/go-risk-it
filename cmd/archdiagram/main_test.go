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
			"game/internal/logic/board": {Suffix: "game/internal/logic/board", Layer: layerLogic},
			"game/internal/logic/card":  {Suffix: "game/internal/logic/card", Layer: layerLogic},
			"game/internal/logic/phase": {Suffix: "game/internal/logic/phase", Layer: layerLogic},
			"web/game/rest":             {Suffix: "web/game/rest", Layer: "Web"},
		},
		Subsystems: map[string]*model.SubsystemInfo{
			"game_services": {
				ID:       "game_services",
				Label:    "Game Services",
				Packages: []string{"game/internal/logic/board", "game/internal/logic/card", "game/internal/logic/phase"},
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
			ImportPath: model.ModulePrefix + "game/internal/logic/board",
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

	if got["game/internal/logic/board"] != "/src/internal/game/logic/board" {
		t.Errorf("game/internal/logic/board = %q, want %q",
			got["game/internal/logic/board"], "/src/internal/game/logic/board")
	}

	if got["config"] != "/src/internal/config" {
		t.Errorf("config = %q, want %q", got["config"], "/src/internal/config")
	}
}

func TestGenerateD2(t *testing.T) {
	t.Parallel()

	archModel := &model.ArchModel{
		Packages: map[string]*model.PackageInfo{
			"game/internal/logic/board": {
				Suffix: "game/internal/logic/board",
				Layer:  "Logic",
			},
		},
		Subsystems: map[string]*model.SubsystemInfo{
			"game_services": {
				ID:       "game_services",
				Label:    "Game Services",
				Layer:    "Logic",
				Packages: []string{"game/internal/logic/board"},
			},
			"lobby_logic": {
				ID:       "lobby_logic",
				Label:    "Lobby Logic",
				Layer:    "Logic",
				Packages: []string{"lobby/logic/management"},
			},
			"database": {
				ID:       "database",
				Label:    "Database",
				Layer:    "Data",
				Packages: []string{"data/db"},
			},
		},
		Layers: layers,
		Edges: []model.Edge{
			{From: "Logic", To: "Data"},
		},
	}

	d2Output := generateD2(archModel)

	assertD2Contains(t, d2Output, "title: go-risk-it Architecture", "title")
	assertD2Contains(t, d2Output, "direction: down", "direction")
	assertD2Contains(t, d2Output, `logic: "⚙️ Logic"`, "Logic visual container with emoji")
	assertD2Contains(t, d2Output, `data: "💾 Data"`, "Data visual container with emoji")
	assertD2Contains(t, d2Output, `game: "Game"`, "Game sub-container")
	assertD2Contains(t, d2Output, `lobby: "Lobby"`, "Lobby sub-container")
	assertD2Contains(t, d2Output, `game_services: "Game Services"`, "game_services node")
	assertD2Contains(t, d2Output, `database: "Database"`, "database node")
	assertD2Contains(t, d2Output, "logic -> data", "cross-layer edge")
	assertD2Contains(t, d2Output, "style.border-radius: 12", "top-level border-radius")
	assertD2Contains(t, d2Output, "style.border-radius: 8", "sub-container border-radius")
	assertD2Contains(t, d2Output, "style.border-radius: 6", "leaf node border-radius")
	assertD2Contains(t, d2Output, `style.stroke: "#2E7D32"`, "Logic stroke color")
	assertD2Contains(t, d2Output, `style.fill: "#BBDEFB"`, "Game sub-container fill")
	assertD2NotContains(t, d2Output, "test:", "Test container")
	assertD2NotContains(t, d2Output, `\n`, "multi-line labels")

	for _, excluded := range []string{"Game-domain", "Lobby-domain", "Game-support", "Events-domain"} {
		assertD2NotContains(t, d2Output, excluded, excluded+" container")
	}
}

func assertD2Contains(t *testing.T, d2Output, expected, description string) {
	t.Helper()

	if !strings.Contains(d2Output, expected) {
		t.Errorf("expected %s in D2 output", description)
		t.Logf("D2 output:\n%s", d2Output)
	}
}

func assertD2NotContains(t *testing.T, d2Output, unexpected, description string) {
	t.Helper()

	if strings.Contains(d2Output, unexpected) {
		t.Errorf("expected no %s in D2 output", description)
		t.Logf("D2 output:\n%s", d2Output)
	}
}

func TestContainerID(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		want string
	}{
		{"Logic", "logic"},
		{"Web", "web"},
		{"Events", "events"},
		{"Kernel", "kernel"},
		{"Data", "data"},
		{"API", "api"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := containerID(testCase.name)
			if got != testCase.want {
				t.Errorf(
					"containerID(%q) = %q, want %q",
					testCase.name,
					got,
					testCase.want,
				)
			}
		})
	}
}

func TestLayerOrdering(t *testing.T) {
	t.Parallel()

	// Verify visual containers are emitted in order by checking D2 output.
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
				Packages: []string{"game/internal/logic/board"},
			},
			"lobby_logic": {
				ID:       "lobby_logic",
				Label:    "Lobby Logic",
				Layer:    "Logic",
				Packages: []string{"lobby/logic/management"},
			},
			"game_consumers": {
				ID:       "game_consumers",
				Label:    "Game Consumers",
				Layer:    "Web",
				Packages: []string{"game/consumers"},
			},
			"lobby_consumers": {
				ID:       "lobby_consumers",
				Label:    "Lobby Consumers",
				Layer:    "Web",
				Packages: []string{"lobby/consumers"},
			},
		},
		Layers: layers,
	}

	d2Output := generateD2(archModel)

	// Find positions of visual containers in output.
	apiPos := strings.Index(d2Output, `api: "📋 API"`)
	webPos := strings.Index(d2Output, `web: "🌐 Web"`)
	logicPos := strings.Index(d2Output, `logic: "⚙️ Logic"`)
	kernelPos := strings.Index(d2Output, `kernel: "🔧 Kernel"`)

	if apiPos < 0 || webPos < 0 || logicPos < 0 || kernelPos < 0 {
		t.Fatalf("missing visual containers in output:\n%s", d2Output)
	}

	if apiPos >= webPos {
		t.Errorf("API (pos %d) should appear before Web (pos %d)", apiPos, webPos)
	}

	if webPos >= logicPos {
		t.Errorf("Web (pos %d) should appear before Logic (pos %d)", webPos, logicPos)
	}

	if logicPos >= kernelPos {
		t.Errorf("Logic (pos %d) should appear before Kernel (pos %d)", logicPos, kernelPos)
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

	if !strings.Contains(d2Output, `kernel: "🔧 Kernel"`) {
		t.Error("expected Kernel visual container")
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
				Packages: []string{"game/internal/logic/zebra"},
			},
			"alpha": {
				ID:       "alpha",
				Label:    "Alpha",
				Layer:    "Logic",
				Packages: []string{"game/internal/logic/alpha"},
			},
		},
		Layers: map[string]*model.LayerInfo{
			"Logic": {Name: "Logic", Color: "#E8F5E9", Order: 0},
		},
	}

	d2Output := generateD2(archModel)

	alphaPos := strings.Index(d2Output, `alpha: "Alpha" {`)
	zebraPos := strings.Index(d2Output, `zebra: "Zebra" {`)

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

func TestGenerateD2_TestLayerExcluded(t *testing.T) {
	t.Parallel()

	archModel := &model.ArchModel{
		Packages: map[string]*model.PackageInfo{},
		Subsystems: map[string]*model.SubsystemInfo{
			"testing": {
				ID:       "testing",
				Label:    "Testing",
				Layer:    "Test",
				Packages: []string{"testonly"},
			},
			"kernel": {
				ID:       "kernel",
				Label:    "Kernel",
				Layer:    "Kernel",
				Packages: []string{"kernel/config"},
			},
		},
		Layers: layers,
		Edges: []model.Edge{
			{From: "Test", To: "Kernel"},
		},
	}

	d2Output := generateD2(archModel)

	// Test container should not appear.
	if strings.Contains(d2Output, `"Testing"`) {
		t.Error("expected no Testing subsystem in D2 output")
		t.Logf("D2 output:\n%s", d2Output)
	}

	// Edge from Test should be dropped.
	if strings.Contains(d2Output, "test ->") {
		t.Error("expected no edges from Test container")
		t.Logf("D2 output:\n%s", d2Output)
	}
}

func TestGenerateD2_EdgeDeduplication(t *testing.T) {
	t.Parallel()

	// Game-domain -> Kernel and Logic -> Kernel both map to Logic -> Kernel
	// after visual mapping. Should produce only one edge.
	archModel := &model.ArchModel{
		Packages:   map[string]*model.PackageInfo{},
		Subsystems: map[string]*model.SubsystemInfo{},
		Layers:     layers,
		Edges: []model.Edge{
			{From: "Game-domain", To: "Kernel"},
			{From: "Logic", To: "Kernel"},
		},
	}

	d2Output := generateD2(archModel)

	edgeCount := strings.Count(d2Output, "logic -> kernel")
	if edgeCount != 1 {
		t.Errorf("expected 1 logic -> kernel edge, got %d", edgeCount)
		t.Logf("D2 output:\n%s", d2Output)
	}
}

func TestGenerateD2_SameVisualEdgeDropped(t *testing.T) {
	t.Parallel()

	// Game-domain -> Logic maps to Logic -> Logic (same visual).
	// Should produce no edge.
	archModel := &model.ArchModel{
		Packages:   map[string]*model.PackageInfo{},
		Subsystems: map[string]*model.SubsystemInfo{},
		Layers:     layers,
		Edges: []model.Edge{
			{From: "Game-domain", To: "Logic"},
		},
	}

	d2Output := generateD2(archModel)

	if strings.Contains(d2Output, "->") {
		t.Error("expected no edges when source and target map to same visual container")
		t.Logf("D2 output:\n%s", d2Output)
	}
}

func TestDetectModule(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		packages []string
		want     string
	}{
		{
			name:     "all game packages",
			packages: []string{"game/internal/logic/board", "game/internal/logic/card"},
			want:     "game",
		},
		{
			name:     "all lobby packages",
			packages: []string{"lobby/logic/management", "lobby/consumers"},
			want:     "lobby",
		},
		{
			name:     "mixed game and lobby",
			packages: []string{"game/api", "lobby/api"},
			want:     "",
		},
		{
			name:     "shared infrastructure",
			packages: []string{"kernel/config", "kernel/ctx"},
			want:     "",
		},
		{
			name:     "single game package",
			packages: []string{"game/ctx"},
			want:     "game",
		},
		{
			name:     "single lobby package",
			packages: []string{"lobby/ctx"},
			want:     "lobby",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			sub := &model.SubsystemInfo{
				ID:       "test_sub",
				Label:    "Test Sub",
				Packages: testCase.packages,
			}

			got := detectModule(sub)
			if got != testCase.want {
				t.Errorf("detectModule(%v) = %q, want %q",
					testCase.packages, got, testCase.want)
			}
		})
	}
}

func TestVisualContainerMapping(t *testing.T) {
	t.Parallel()

	// Every enforcement layer must map to a visual container or "".
	for layerName := range layers {
		visualName, ok := layerToVisual[layerName]
		if !ok {
			t.Errorf("layer %q has no entry in layerToVisual", layerName)

			continue
		}

		if visualName == "" {
			// Excluded (e.g., Test) — valid.
			continue
		}

		if _, ok := visualContainers[visualName]; !ok {
			t.Errorf("layer %q maps to visual %q which is not in visualContainers",
				layerName, visualName)
		}
	}

	// Verify we have exactly 6 visual containers.
	if len(visualContainers) != 6 {
		t.Errorf("expected 6 visual containers, got %d", len(visualContainers))
	}

	// Verify unique orders.
	orders := make(map[int]string)

	for name, info := range visualContainers {
		if prev, ok := orders[info.Order]; ok {
			t.Errorf("duplicate visual container order %d: %q and %q",
				info.Order, prev, name)
		}

		orders[info.Order] = name
	}
}

func TestVisualContainerNesting(t *testing.T) {
	t.Parallel()

	// Build a model where Logic has both game and lobby subsystems.
	archModel := &model.ArchModel{
		Packages: map[string]*model.PackageInfo{},
		Subsystems: map[string]*model.SubsystemInfo{
			"game_services": {
				ID:       "game_services",
				Label:    "Game Services",
				Layer:    "Logic",
				Packages: []string{"game/internal/logic/board", "game/internal/logic/card"},
			},
			"lobby_logic": {
				ID:       "lobby_logic",
				Label:    "Lobby Logic",
				Layer:    "Logic",
				Packages: []string{"lobby/logic/management"},
			},
		},
		Layers: layers,
	}

	d2Output := generateD2(archModel)

	// Logic container should have nested game and lobby sub-containers.
	if !strings.Contains(d2Output, `game: "Game" {`) {
		t.Error("expected Game sub-container in Logic")
		t.Logf("D2 output:\n%s", d2Output)
	}

	if !strings.Contains(d2Output, `lobby: "Lobby" {`) {
		t.Error("expected Lobby sub-container in Logic")
		t.Logf("D2 output:\n%s", d2Output)
	}

	// game_services should be inside the game sub-container.
	gameBlockStart := strings.Index(d2Output, `game: "Game" {`)
	gameServicesPos := strings.Index(d2Output, `game_services: "Game Services" {`)

	if gameBlockStart < 0 || gameServicesPos < 0 {
		t.Fatal("missing game block or game_services node")
	}

	if gameServicesPos <= gameBlockStart {
		t.Errorf("game_services (pos %d) should appear after Game block (pos %d)",
			gameServicesPos, gameBlockStart)
	}

	// lobby_logic should be inside the lobby sub-container.
	lobbyBlockStart := strings.Index(d2Output, `lobby: "Lobby" {`)
	lobbyLogicPos := strings.Index(d2Output, `lobby_logic: "Lobby Logic" {`)

	if lobbyBlockStart < 0 || lobbyLogicPos < 0 {
		t.Fatal("missing lobby block or lobby_logic node")
	}

	if lobbyLogicPos <= lobbyBlockStart {
		t.Errorf("lobby_logic (pos %d) should appear after Lobby block (pos %d)",
			lobbyLogicPos, lobbyBlockStart)
	}
}

func TestVisualContainerNesting_FlatWhenSingleModule(t *testing.T) {
	t.Parallel()

	// Kernel has only shared subsystems — should be flat (no nesting).
	archModel := &model.ArchModel{
		Packages: map[string]*model.PackageInfo{},
		Subsystems: map[string]*model.SubsystemInfo{
			"kernel": {
				ID:       "kernel",
				Label:    "Kernel",
				Layer:    "Kernel",
				Packages: []string{"kernel/config", "kernel/ctx"},
			},
		},
		Layers: layers,
	}

	d2Output := generateD2(archModel)

	// Should have the Kernel container with flat subsystem.
	if !strings.Contains(d2Output, `kernel: "🔧 Kernel"`) {
		t.Error("expected Kernel visual container")
		t.Logf("D2 output:\n%s", d2Output)
	}

	if !strings.Contains(d2Output, `kernel: "Kernel" {`) {
		t.Error("expected kernel subsystem node with block")
		t.Logf("D2 output:\n%s", d2Output)
	}

	// Should NOT have Game/Lobby sub-containers.
	// Count occurrences of "Game" in braces to distinguish from subsystem labels.
	if strings.Contains(d2Output, `"Game" {`) {
		t.Error("expected no Game sub-container for Kernel-only container")
		t.Logf("D2 output:\n%s", d2Output)
	}
}

func TestVisualContainerNesting_SharedSubContainer(t *testing.T) {
	t.Parallel()

	// Web has game, lobby, AND shared subsystems — shared ones get a sub-container.
	archModel := &model.ArchModel{
		Packages: map[string]*model.PackageInfo{},
		Subsystems: map[string]*model.SubsystemInfo{
			"game_consumers": {
				ID:       "game_consumers",
				Label:    "Game Consumers",
				Layer:    "Web",
				Packages: []string{"game/consumers", "game/routes", "game/ws"},
			},
			"lobby_consumers": {
				ID:       "lobby_consumers",
				Label:    "Lobby Consumers",
				Layer:    "Web",
				Packages: []string{"lobby/consumers", "lobby/routes", "lobby/ws"},
			},
			"middleware": {
				ID:       "middleware",
				Label:    "Middleware",
				Layer:    "Web",
				Packages: []string{"web/middleware", "web/mux", "web/nbio"},
			},
		},
		Layers: layers,
	}

	d2Output := generateD2(archModel)

	// Should have nested sub-containers including Shared.
	if !strings.Contains(d2Output, `game: "Game" {`) {
		t.Error("expected Game sub-container in Web")
		t.Logf("D2 output:\n%s", d2Output)
	}

	if !strings.Contains(d2Output, `lobby: "Lobby" {`) {
		t.Error("expected Lobby sub-container in Web")
		t.Logf("D2 output:\n%s", d2Output)
	}

	if !strings.Contains(d2Output, `shared: "Shared" {`) {
		t.Error("expected Shared sub-container in Web")
		t.Logf("D2 output:\n%s", d2Output)
	}

	if !strings.Contains(d2Output, `middleware: "Middleware" {`) {
		t.Error("expected middleware node inside Shared sub-container")
		t.Logf("D2 output:\n%s", d2Output)
	}
}
