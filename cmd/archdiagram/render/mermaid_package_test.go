package render_test

import (
	"strings"
	"testing"

	"github.com/go-risk-it/go-risk-it/cmd/archdiagram/model"
	. "github.com/go-risk-it/go-risk-it/cmd/archdiagram/render"
)

func TestRenderPackageArch_Basic(t *testing.T) {
	t.Parallel()

	archModel := &model.ArchModel{
		Packages: map[string]*model.PackageInfo{
			"game/internal/logic/board": {Suffix: "game/internal/logic/board", Layer: "Logic"},
			"web/game/rest":             {Suffix: "web/game/rest", Layer: "Web"},
			"data/game/db":              {Suffix: "data/game/db", Layer: "Data"},
		},
		Subsystems: map[string]*model.SubsystemInfo{
			"game_services": {
				ID:       "game_services",
				Label:    "Game Services",
				Layer:    "Logic",
				Packages: []string{"game/internal/logic/board"},
			},
			"rest_utils": {
				ID:       "rest_utils",
				Label:    "REST Utils",
				Layer:    "Web",
				Packages: []string{"web/game/rest"},
			},
			"game_data": {
				ID:       "game_data",
				Label:    "Game Data",
				Layer:    "Data",
				Packages: []string{"data/game/db"},
			},
		},
		Layers: map[string]*model.LayerInfo{
			"Logic": {Name: "Logic", Color: "#E8F5E9", Order: 0},
			"Web":   {Name: "Web", Color: "#E3F2FD", Order: 1},
			"Data":  {Name: "Data", Color: "#FFF3E0", Order: 2},
		},
		Edges: []model.Edge{
			{From: "Web", To: "Logic"},
			{From: "Logic", To: "Data"},
		},
	}

	result := RenderPackageArch(archModel)

	// Must contain graph TD.
	if !strings.Contains(result, "graph TD") {
		t.Error("expected graph TD directive")
	}

	// Must contain layer subgraphs.
	if !strings.Contains(result, "Logic Layer") {
		t.Error("expected Logic Layer subgraph")
	}

	if !strings.Contains(result, "Web Layer") {
		t.Error("expected Web Layer subgraph")
	}

	if !strings.Contains(result, "Data Layer") {
		t.Error("expected Data Layer subgraph")
	}

	// Must contain subsystem nodes.
	if !strings.Contains(result, "Game Services") {
		t.Error("expected Game Services node")
	}

	// Must contain edges.
	if !strings.Contains(result, "-->") {
		t.Error("expected layer-to-layer edges")
	}

	// Must be wrapped in mermaid fences.
	if !strings.HasPrefix(result, "```mermaid\n") {
		t.Error("expected mermaid fence prefix")
	}

	if !strings.HasSuffix(result, "```") {
		t.Error("expected mermaid fence suffix")
	}
}

func TestRenderPackageArch_Deterministic(t *testing.T) {
	t.Parallel()

	archModel := &model.ArchModel{
		Packages: map[string]*model.PackageInfo{
			"a": {Suffix: "a", Layer: "Logic"},
			"b": {Suffix: "b", Layer: "Web"},
		},
		Subsystems: map[string]*model.SubsystemInfo{
			"sub_a": {ID: "sub_a", Label: "Sub A", Layer: "Logic", Packages: []string{"a"}},
			"sub_b": {ID: "sub_b", Label: "Sub B", Layer: "Web", Packages: []string{"b"}},
		},
		Layers: map[string]*model.LayerInfo{
			"Logic": {Name: "Logic", Color: "#E8F5E9", Order: 0},
			"Web":   {Name: "Web", Color: "#E3F2FD", Order: 1},
		},
		Edges: []model.Edge{
			{From: "Web", To: "Logic"},
		},
	}

	r1 := RenderPackageArch(archModel)
	r2 := RenderPackageArch(archModel)

	if r1 != r2 {
		t.Error("RenderPackageArch is not deterministic")
	}
}

func TestRenderPackageArch_EmptyModel(t *testing.T) {
	t.Parallel()

	archModel := &model.ArchModel{
		Packages:   map[string]*model.PackageInfo{},
		Subsystems: map[string]*model.SubsystemInfo{},
		Layers:     map[string]*model.LayerInfo{},
	}

	result := RenderPackageArch(archModel)

	if !strings.Contains(result, "graph TD") {
		t.Error("expected graph TD even with empty model")
	}

	// Should not contain edges.
	if strings.Contains(result, "-->") {
		t.Error("expected no edges for empty model")
	}
}

func TestRenderPackageArch_LayerOrder(t *testing.T) {
	t.Parallel()

	archModel := &model.ArchModel{
		Packages: map[string]*model.PackageInfo{},
		Subsystems: map[string]*model.SubsystemInfo{
			"z": {ID: "z", Label: "Z", Layer: "Data", Packages: []string{}},
			"a": {ID: "a", Label: "A", Layer: "API", Packages: []string{}},
		},
		Layers: map[string]*model.LayerInfo{
			"API":  {Name: "API", Color: "#E8EAF6", Order: 0},
			"Data": {Name: "Data", Color: "#FFF3E0", Order: 2},
		},
	}

	result := RenderPackageArch(archModel)

	apiPos := strings.Index(result, "API Layer")
	dataPos := strings.Index(result, "Data Layer")

	if apiPos < 0 || dataPos < 0 {
		t.Fatal("missing layer subgraphs")
	}

	if apiPos >= dataPos {
		t.Errorf("API (pos %d) should appear before Data (pos %d)", apiPos, dataPos)
	}
}
