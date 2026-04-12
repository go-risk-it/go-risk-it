package render_test

import (
	"strings"
	"testing"

	"github.com/go-risk-it/go-risk-it/cmd/archdiagram/model"
	. "github.com/go-risk-it/go-risk-it/cmd/archdiagram/render"
)

func TestRenderPackageTables_Basic(t *testing.T) {
	t.Parallel()

	pkgs := []model.GoPackage{
		// Wiring root.
		{
			ImportPath: model.ModulePrefix + "kernel",
			GoFiles:    []string{"kernel.go"},
		},
		// Full-tier package (has service.go, >2 GoFiles).
		{
			ImportPath: model.ModulePrefix + "game/internal/logic/board",
			GoFiles:    []string{"service.go", "graph.go", "continents.go", "doc.go"},
		},
		// Lightweight package (no service.go).
		{
			ImportPath: model.ModulePrefix + "kernel/config",
			GoFiles:    []string{"config.go", "doc.go"},
		},
	}

	archModel := &model.ArchModel{
		Packages: map[string]*model.PackageInfo{
			"game/internal/logic/board": {
				Suffix:  "game/internal/logic/board",
				Layer:   "Logic",
				GoFiles: []string{"service.go", "graph.go", "continents.go", "doc.go"},
			},
			"kernel/config": {
				Suffix:  "kernel/config",
				Layer:   "Kernel",
				GoFiles: []string{"config.go", "doc.go"},
			},
		},
		Subsystems: map[string]*model.SubsystemInfo{},
		Layers:     map[string]*model.LayerInfo{},
	}

	result := RenderPackageTables(archModel, pkgs)

	// Wiring roots table.
	if !strings.Contains(result, "### Wiring roots") {
		t.Error("expected wiring roots heading")
	}

	if !strings.Contains(result, "| `kernel` | `kernel.go` |") {
		t.Error("expected kernel wiring root entry")
	}

	// Full-tier table.
	if !strings.Contains(result, "### Full-tier packages") {
		t.Error("expected full-tier heading")
	}

	if !strings.Contains(result, "| `game/internal/logic/board` | logic | 4 |") {
		t.Error("expected game/internal/logic/board full-tier entry")
	}

	// Lightweight table.
	if !strings.Contains(result, "### Lightweight-tier packages") {
		t.Error("expected lightweight-tier heading")
	}

	if !strings.Contains(result, "| `kernel/config` | kernel |") {
		t.Error("expected kernel/config lightweight entry")
	}

	// Note at the end.
	if !strings.Contains(result, "Note: Packages with") {
		t.Error("expected trailing note")
	}
}

func TestRenderPackageTables_Deterministic(t *testing.T) {
	t.Parallel()

	pkgs := []model.GoPackage{
		{ImportPath: model.ModulePrefix + "kernel", GoFiles: []string{"kernel.go"}},
		{
			ImportPath: model.ModulePrefix + "game/internal/logic/board",
			GoFiles:    []string{"service.go", "graph.go", "doc.go"},
		},
		{
			ImportPath: model.ModulePrefix + "kernel/config",
			GoFiles:    []string{"config.go"},
		},
	}

	archModel := &model.ArchModel{
		Packages: map[string]*model.PackageInfo{
			"game/internal/logic/board": {
				Suffix:  "game/internal/logic/board",
				Layer:   "Logic",
				GoFiles: []string{"service.go", "graph.go", "doc.go"},
			},
			"kernel/config": {
				Suffix:  "kernel/config",
				Layer:   "Kernel",
				GoFiles: []string{"config.go"},
			},
		},
		Subsystems: map[string]*model.SubsystemInfo{},
		Layers:     map[string]*model.LayerInfo{},
	}

	r1 := RenderPackageTables(archModel, pkgs)
	r2 := RenderPackageTables(archModel, pkgs)

	if r1 != r2 {
		t.Error("RenderPackageTables is not deterministic")
	}
}

func TestRenderPackageTables_EmptyModel(t *testing.T) {
	t.Parallel()

	archModel := &model.ArchModel{
		Packages:   map[string]*model.PackageInfo{},
		Subsystems: map[string]*model.SubsystemInfo{},
		Layers:     map[string]*model.LayerInfo{},
	}

	result := RenderPackageTables(archModel, nil)

	// Should still have all three headings.
	if !strings.Contains(result, "### Wiring roots") {
		t.Error("expected wiring roots heading")
	}

	if !strings.Contains(result, "### Full-tier packages") {
		t.Error("expected full-tier heading")
	}

	if !strings.Contains(result, "### Lightweight-tier packages") {
		t.Error("expected lightweight-tier heading")
	}
}

func TestIsFullTier(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		info   *model.PackageInfo
		wantFT bool
	}{
		{
			name:   "full-tier: has service.go and >2 files",
			info:   &model.PackageInfo{GoFiles: []string{"service.go", "types.go", "doc.go"}},
			wantFT: true,
		},
		{
			name:   "lightweight: no service.go",
			info:   &model.PackageInfo{GoFiles: []string{"config.go", "doc.go", "types.go"}},
			wantFT: false,
		},
		{
			name:   "lightweight: service.go but only 2 files",
			info:   &model.PackageInfo{GoFiles: []string{"service.go", "doc.go"}},
			wantFT: false,
		},
		{
			name:   "lightweight: empty GoFiles",
			info:   &model.PackageInfo{GoFiles: nil},
			wantFT: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := IsFullTier(testCase.info)
			if got != testCase.wantFT {
				t.Errorf("IsFullTier() = %v, want %v", got, testCase.wantFT)
			}
		})
	}
}

func TestRenderPackageTables_SortedAlphabetically(t *testing.T) {
	t.Parallel()

	pkgs := []model.GoPackage{}

	archModel := &model.ArchModel{
		Packages: map[string]*model.PackageInfo{
			"web/middleware": {
				Suffix:  "web/middleware",
				Layer:   "Web",
				GoFiles: []string{"middleware.go"},
			},
			"game/api": {
				Suffix:  "game/api",
				Layer:   "API",
				GoFiles: []string{"types.go"},
			},
			"kernel/config": {
				Suffix:  "kernel/config",
				Layer:   "Kernel",
				GoFiles: []string{"config.go"},
			},
		},
		Subsystems: map[string]*model.SubsystemInfo{},
		Layers:     map[string]*model.LayerInfo{},
	}

	result := RenderPackageTables(archModel, pkgs)

	// All are lightweight; verify alphabetical order.
	gamePos := strings.Index(result, "| `game/api`")
	kernelPos := strings.Index(result, "| `kernel/config`")
	webPos := strings.Index(result, "| `web/middleware`")

	if gamePos < 0 || kernelPos < 0 || webPos < 0 {
		t.Fatal("missing expected lightweight entries")
	}

	if gamePos >= kernelPos || kernelPos >= webPos {
		t.Errorf(
			"entries not in alphabetical order: game=%d, kernel=%d, web=%d",
			gamePos,
			kernelPos,
			webPos,
		)
	}
}
