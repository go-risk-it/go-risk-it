package render_test

import (
	"strings"
	"testing"

	"github.com/go-risk-it/go-risk-it/cmd/archdiagram/model"
	. "github.com/go-risk-it/go-risk-it/cmd/archdiagram/render"
)

func TestRenderProjectTree_Basic(t *testing.T) {
	t.Parallel()

	pkgs := []model.GoPackage{
		{ImportPath: model.ModulePrefix + "kernel/config"},
		{ImportPath: model.ModulePrefix + "kernel/ctx"},
		{ImportPath: model.ModulePrefix + "game/internal/logic/board"},
		{ImportPath: model.ModulePrefix + "game/internal/logic/phase"},
	}

	archModel := &model.ArchModel{
		Packages: map[string]*model.PackageInfo{
			"kernel/config": {
				Suffix:  "kernel/config",
				Summary: "Configuration management",
			},
			"kernel/ctx": {Suffix: "kernel/ctx", Summary: "Typed contexts"},
			"game/internal/logic/board": {
				Suffix:  "game/internal/logic/board",
				Summary: "Board topology",
			},
			"game/internal/logic/phase": {
				Suffix:  "game/internal/logic/phase",
				Summary: "Phase transitions",
			},
		},
		Subsystems: map[string]*model.SubsystemInfo{},
		Layers:     map[string]*model.LayerInfo{},
	}

	result := RenderProjectTree(archModel, pkgs)

	// Must start with code fence + internal/.
	if !strings.HasPrefix(result, "```\ninternal/\n") {
		t.Errorf("expected tree prefix, got: %s", result[:30])
	}

	// Must end with closing fence.
	if !strings.HasSuffix(result, "```") {
		t.Error("expected closing code fence")
	}

	// Must contain expected directories.
	if !strings.Contains(result, "kernel/") {
		t.Error("expected kernel/ directory")
	}

	if !strings.Contains(result, "config/") {
		t.Error("expected config/ directory")
	}

	// Must contain box-drawing characters.
	if !strings.Contains(result, "\u251C") && !strings.Contains(result, "\u2514") {
		t.Error("expected box-drawing characters")
	}

	// Must contain summaries.
	if !strings.Contains(result, "# Board topology") {
		t.Error("expected summary annotation for board")
	}
}

func TestRenderProjectTree_Deterministic(t *testing.T) {
	t.Parallel()

	pkgs := []model.GoPackage{
		{ImportPath: model.ModulePrefix + "web/middleware"},
		{ImportPath: model.ModulePrefix + "game/internal/logic/board"},
		{ImportPath: model.ModulePrefix + "kernel/config"},
	}

	archModel := &model.ArchModel{
		Packages: map[string]*model.PackageInfo{
			"web/middleware":            {Suffix: "web/middleware", Summary: "Auth middleware"},
			"game/internal/logic/board": {Suffix: "game/internal/logic/board", Summary: "Board"},
			"kernel/config":             {Suffix: "kernel/config", Summary: "Config"},
		},
		Subsystems: map[string]*model.SubsystemInfo{},
		Layers:     map[string]*model.LayerInfo{},
	}

	r1 := RenderProjectTree(archModel, pkgs)
	r2 := RenderProjectTree(archModel, pkgs)

	if r1 != r2 {
		t.Error("RenderProjectTree is not deterministic")
	}
}

func TestRenderProjectTree_ExcludesSqlcAndMocks(t *testing.T) {
	t.Parallel()

	pkgs := []model.GoPackage{
		{ImportPath: model.ModulePrefix + "game/internal/data/db"},
		{ImportPath: model.ModulePrefix + "game/internal/data/sqlc"},
		{ImportPath: model.ModulePrefix + "game/mocks"},
	}

	archModel := &model.ArchModel{
		Packages: map[string]*model.PackageInfo{
			"game/internal/data/db": {Suffix: "game/internal/data/db", Summary: "Game DB"},
		},
		Subsystems: map[string]*model.SubsystemInfo{},
		Layers:     map[string]*model.LayerInfo{},
	}

	result := RenderProjectTree(archModel, pkgs)

	if strings.Contains(result, "sqlc") {
		t.Error("expected sqlc to be excluded from tree")
	}

	if strings.Contains(result, "mocks") {
		t.Error("expected mocks to be excluded from tree")
	}
}

func TestRenderProjectTree_WiringRootsShowAsDirectories(t *testing.T) {
	t.Parallel()

	pkgs := []model.GoPackage{
		{ImportPath: model.ModulePrefix + "kernel"},
		{ImportPath: model.ModulePrefix + "kernel/config"},
	}

	archModel := &model.ArchModel{
		Packages: map[string]*model.PackageInfo{
			"kernel/config": {Suffix: "kernel/config", Summary: "Config management"},
		},
		Subsystems: map[string]*model.SubsystemInfo{},
		Layers:     map[string]*model.LayerInfo{},
	}

	result := RenderProjectTree(archModel, pkgs)

	// kernel/ should appear as a directory (it's a wiring root, no summary).
	if !strings.Contains(result, "kernel/") {
		t.Error("expected kernel/ directory node")
	}

	// kernel/ line should NOT have a summary annotation (wiring roots are not in model).
	lines := strings.SplitSeq(result, "\n")
	for line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "kernel/") && !strings.Contains(trimmed, "config") {
			if strings.Contains(line, "# ") {
				t.Errorf("wiring root 'kernel/' should not have summary annotation: %s", line)
			}
		}
	}
}

func TestRenderProjectTree_EmptyModel(t *testing.T) {
	t.Parallel()

	archModel := &model.ArchModel{
		Packages:   map[string]*model.PackageInfo{},
		Subsystems: map[string]*model.SubsystemInfo{},
		Layers:     map[string]*model.LayerInfo{},
	}

	result := RenderProjectTree(archModel, nil)

	want := "```\ninternal/\n```"
	if result != want {
		t.Errorf("expected %q for empty model, got %q", want, result)
	}
}
