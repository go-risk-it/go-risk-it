package render_test

import (
	"fmt"
	"strings"
	"testing"

	. "github.com/go-risk-it/go-risk-it/cmd/archdiagram/render"
)

func TestRenderStats_ContainsCounts(t *testing.T) {
	t.Parallel()

	result := RenderStats(28, 12)

	if !strings.Contains(result, "28 architecture rules") {
		t.Error("expected '28 architecture rules' in output")
	}

	if !strings.Contains(result, "12 game-state invariants") {
		t.Error("expected '12 game-state invariants' in output")
	}
}

func TestRenderStats_DifferentCounts(t *testing.T) {
	t.Parallel()

	result := RenderStats(42, 15)

	if !strings.Contains(result, "42 architecture rules") {
		t.Error("expected '42 architecture rules' in output")
	}

	if !strings.Contains(result, "15 game-state invariants") {
		t.Error("expected '15 game-state invariants' in output")
	}
}

func TestRenderStats_AllBullets(t *testing.T) {
	t.Parallel()

	result := RenderStats(1, 1)

	expectedBullets := []string{
		"**Living Architecture**",
		"**Property-Based Testing**",
		"**Event-Driven Architecture**",
		"**Type-Safe Move Pipeline**",
		"**Auto-Generated Docs**",
	}

	for _, bullet := range expectedBullets {
		if !strings.Contains(result, bullet) {
			t.Errorf("missing bullet: %s", bullet)
		}
	}
}

func TestRenderStats_Deterministic(t *testing.T) {
	t.Parallel()

	r1 := RenderStats(28, 12)
	r2 := RenderStats(28, 12)

	if r1 != r2 {
		t.Error("RenderStats is not deterministic")
	}
}

func TestRenderStats_CountFormats(t *testing.T) {
	t.Parallel()

	tests := []struct {
		rules      int
		invariants int
		wantRule   string
		wantInv    string
	}{
		{1, 1, "1 architecture rules", "1 game-state invariants"},
		{100, 50, "100 architecture rules", "50 game-state invariants"},
	}

	for _, testCase := range tests {
		t.Run(
			fmt.Sprintf("rules=%d,inv=%d", testCase.rules, testCase.invariants),
			func(t *testing.T) {
				t.Parallel()

				result := RenderStats(testCase.rules, testCase.invariants)

				if !strings.Contains(result, testCase.wantRule) {
					t.Errorf("expected %q in output", testCase.wantRule)
				}

				if !strings.Contains(result, testCase.wantInv) {
					t.Errorf("expected %q in output", testCase.wantInv)
				}
			},
		)
	}
}
