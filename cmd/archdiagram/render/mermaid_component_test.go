package render_test

import (
	"strings"
	"testing"

	"github.com/go-risk-it/go-risk-it/cmd/archdiagram/model"
	. "github.com/go-risk-it/go-risk-it/cmd/archdiagram/render"
)

func TestRenderComponentArch_Basic(t *testing.T) {
	t.Parallel()

	archModel := &model.ArchModel{
		Packages: map[string]*model.PackageInfo{
			"game/internal/logic/board": {Suffix: "game/internal/logic/board", Layer: "Logic"},
			"web/game/rest":             {Suffix: "web/game/rest", Layer: "Web"},
		},
		Subsystems: map[string]*model.SubsystemInfo{
			"game_services": {
				ID:       "game_services",
				Label:    "Game Services",
				Layer:    "Logic",
				Packages: []string{"game/internal/logic/board"},
			},
			"middleware": {
				ID:       "middleware",
				Label:    "Middleware",
				Layer:    "Web",
				Packages: []string{"web/game/rest"},
			},
		},
		Layers: map[string]*model.LayerInfo{
			"Logic": {Name: "Logic", Color: "#E8F5E9", Order: 0},
			"Web":   {Name: "Web", Color: "#E3F2FD", Order: 1},
		},
		Edges: []model.Edge{
			{From: "Web", To: "Logic"},
		},
	}

	result := RenderComponentArch(archModel)

	// Must start with mermaid fence.
	if !strings.HasPrefix(result, "```mermaid\n") {
		t.Errorf("expected mermaid fence prefix, got: %s", result[:40])
	}

	// Must end with closing fence.
	if !strings.HasSuffix(result, "```") {
		t.Error("expected closing mermaid fence")
	}

	// Must contain graph LR.
	if !strings.Contains(result, "graph LR") {
		t.Error("expected graph LR directive")
	}

	// Must contain subgraphs.
	if !strings.Contains(result, "subgraph") {
		t.Error("expected subgraph blocks")
	}

	// Must contain subsystem nodes.
	if !strings.Contains(result, "Game Services") {
		t.Error("expected Game Services node")
	}

	if !strings.Contains(result, "Middleware") {
		t.Error("expected Middleware node")
	}

	// Must contain style directives.
	if !strings.Contains(result, "style") {
		t.Error("expected style directives")
	}
}

func TestRenderComponentArch_Deterministic(t *testing.T) {
	t.Parallel()

	archModel := &model.ArchModel{
		Packages: map[string]*model.PackageInfo{
			"a": {Suffix: "a", Layer: "Logic"},
			"b": {Suffix: "b", Layer: "Web"},
			"c": {Suffix: "c", Layer: "Data"},
		},
		Subsystems: map[string]*model.SubsystemInfo{
			"sub_a": {ID: "sub_a", Label: "Sub A", Layer: "Logic", Packages: []string{"a"}},
			"sub_b": {ID: "sub_b", Label: "Sub B", Layer: "Web", Packages: []string{"b"}},
			"sub_c": {ID: "sub_c", Label: "Sub C", Layer: "Data", Packages: []string{"c"}},
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

	result1 := RenderComponentArch(archModel)
	result2 := RenderComponentArch(archModel)

	if result1 != result2 {
		t.Error("RenderComponentArch is not deterministic")
	}
}

func TestRenderComponentArch_EmptyModel(t *testing.T) {
	t.Parallel()

	archModel := &model.ArchModel{
		Packages:   map[string]*model.PackageInfo{},
		Subsystems: map[string]*model.SubsystemInfo{},
		Layers:     map[string]*model.LayerInfo{},
	}

	result := RenderComponentArch(archModel)

	if !strings.Contains(result, "graph LR") {
		t.Error("expected graph LR even with empty model")
	}
}

func TestRenderComponentArch_LayerOrder(t *testing.T) {
	t.Parallel()

	archModel := &model.ArchModel{
		Packages: map[string]*model.PackageInfo{},
		Subsystems: map[string]*model.SubsystemInfo{
			"z": {ID: "z", Label: "Z", Layer: "Web", Packages: []string{}},
			"a": {ID: "a", Label: "A", Layer: "API", Packages: []string{}},
		},
		Layers: map[string]*model.LayerInfo{
			"API": {Name: "API", Color: "#E8EAF6", Order: 0},
			"Web": {Name: "Web", Color: "#E3F2FD", Order: 8},
		},
	}

	result := RenderComponentArch(archModel)

	apiPos := strings.Index(result, "API")
	webPos := strings.Index(result, "Web")

	if apiPos < 0 || webPos < 0 {
		t.Fatal("missing layer subgraphs")
	}

	if apiPos >= webPos {
		t.Errorf("API (pos %d) should appear before Web (pos %d)", apiPos, webPos)
	}
}
