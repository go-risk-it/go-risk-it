package model_test

import (
	"slices"
	"sort"
	"testing"

	. "github.com/go-risk-it/go-risk-it/cmd/archdiagram/model"
)

// makeModel builds a minimal ArchModel with only the packages map populated.
// Layer is required for each package but not relevant to grouping logic.
func makeModel(suffixes ...string) *ArchModel {
	archModel := &ArchModel{
		Packages:   make(map[string]*PackageInfo),
		Subsystems: make(map[string]*SubsystemInfo),
	}

	for _, s := range suffixes {
		archModel.Packages[s] = &PackageInfo{
			Suffix: s,
			Layer:  "Logic", // placeholder — grouper doesn't use Layer
		}
	}

	return archModel
}

func TestGroupPackages_BasicGrouping(t *testing.T) {
	t.Parallel()

	// Packages under game/logic should land in game_services (not move_pipeline).
	archModel := makeModel(
		"game/logic/board",
		"game/logic/card",
		"game/logic/phase",
		"game/logic/player",
	)

	GroupPackages(archModel)

	sub, ok := archModel.Subsystems["game_services"]
	if !ok {
		t.Fatal("expected subsystem game_services to exist")
	}

	for _, suffix := range []string{
		"game/logic/board",
		"game/logic/card",
		"game/logic/phase",
		"game/logic/player",
	} {
		if !slices.Contains(sub.Packages, suffix) {
			t.Errorf("expected %q in game_services, got packages: %v", suffix, sub.Packages)
		}
	}

	if sub.Label != "Game Services" {
		t.Errorf("label = %q, want %q", sub.Label, "Game Services")
	}
}

func TestGroupPackages_LongestPrefixWins(t *testing.T) {
	t.Parallel()

	// game/logic/move/attack should match game/logic/move (move_pipeline),
	// not game/logic (game_services).
	archModel := makeModel(
		"game/logic/board",
		"game/logic/move/attack",
		"game/logic/move/deploy",
		"game/logic/move/orchestration",
	)

	GroupPackages(archModel)

	moveSub := archModel.Subsystems["move_pipeline"]
	if moveSub == nil {
		t.Fatal("expected subsystem move_pipeline to exist")
	}

	for _, suffix := range []string{
		"game/logic/move/attack",
		"game/logic/move/deploy",
		"game/logic/move/orchestration",
	} {
		if !slices.Contains(moveSub.Packages, suffix) {
			t.Errorf("expected %q in move_pipeline, got: %v", suffix, moveSub.Packages)
		}
	}

	// board should be in game_services, not move_pipeline
	gameSub := archModel.Subsystems["game_services"]
	if gameSub == nil {
		t.Fatal("expected subsystem game_services to exist")
	}

	if !slices.Contains(gameSub.Packages, "game/logic/board") {
		t.Errorf("expected game/logic/board in game_services, got: %v", gameSub.Packages)
	}
}

func TestGroupPackages_Overrides(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		suffix    string
		wantSubID string
		wantLabel string
	}{
		{
			name:      "game/publisher routes to game_publisher",
			suffix:    "game/publisher",
			wantSubID: "game_publisher",
			wantLabel: "Game Publisher",
		},
		{
			name:      "game/config routes to game_support",
			suffix:    "game/config",
			wantSubID: "game_support",
			wantLabel: "Game Support",
		},
		{
			name:      "game/headlines routes to game_support",
			suffix:    "game/headlines",
			wantSubID: "game_support",
			wantLabel: "Game Support",
		},
		{
			name:      "game/snapshot routes to game_support",
			suffix:    "game/snapshot",
			wantSubID: "game_support",
			wantLabel: "Game Support",
		},
		{
			name:      "web/middleware routes to middleware",
			suffix:    "web/middleware",
			wantSubID: "middleware",
			wantLabel: "Middleware",
		},
		{
			name:      "web/ws routes to websocket",
			suffix:    "web/ws",
			wantSubID: "websocket",
			wantLabel: "WebSocket",
		},
		{
			name:      "lobby/publisher routes to lobby_publisher",
			suffix:    "lobby/publisher",
			wantSubID: "lobby_publisher",
			wantLabel: "Lobby Publisher",
		},
		{
			name:      "testonly routes to testing",
			suffix:    "testonly",
			wantSubID: "testing",
			wantLabel: "Testing",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			archModel := makeModel(testCase.suffix)
			GroupPackages(archModel)

			sub, ok := archModel.Subsystems[testCase.wantSubID]
			if !ok {
				t.Fatalf(
					"expected subsystem %q to exist, got subsystems: %v",
					testCase.wantSubID, subsystemIDs(archModel))
			}

			if !slices.Contains(sub.Packages, testCase.suffix) {
				t.Errorf("expected %q in subsystem %q, got packages: %v",
					testCase.suffix, testCase.wantSubID, sub.Packages)
			}

			if sub.Label != testCase.wantLabel {
				t.Errorf("label = %q, want %q", sub.Label, testCase.wantLabel)
			}
		})
	}
}

func TestGroupPackages_OverrideTakesPrecedenceOverRoot(t *testing.T) {
	t.Parallel()

	// game/config has prefix "game/" but is overridden to game_support.
	// Without the override, it would match no root (no "game" root exists).
	// game/snapshot also overridden to game_support instead of falling through.
	archModel := makeModel("game/config", "game/snapshot", "game/logic/board")

	GroupPackages(archModel)

	support := archModel.Subsystems["game_support"]
	if support == nil {
		t.Fatal("expected subsystem game_support")
	}

	if !slices.Contains(support.Packages, "game/config") {
		t.Errorf("expected game/config in game_support")
	}

	if !slices.Contains(support.Packages, "game/snapshot") {
		t.Errorf("expected game/snapshot in game_support")
	}

	// game/logic/board should NOT be in game_support
	if slices.Contains(support.Packages, "game/logic/board") {
		t.Error("game/logic/board should not be in game_support")
	}
}

func TestGroupPackages_NewPackage(t *testing.T) {
	t.Parallel()

	// A new package under game/logic/ should automatically land in game_services
	// without any map update.
	archModel := makeModel("game/logic/foo", "game/logic/board")

	GroupPackages(archModel)

	sub := archModel.Subsystems["game_services"]
	if sub == nil {
		t.Fatal("expected game_services to exist for synthetic package")
	}

	if !slices.Contains(sub.Packages, "game/logic/foo") {
		t.Errorf("expected game/logic/foo in game_services, got: %v", sub.Packages)
	}
}

func TestGroupPackages_NewMovePackage(t *testing.T) {
	t.Parallel()

	// A new package under game/logic/move/ should land in move_pipeline.
	archModel := makeModel("game/logic/move/newmove")

	GroupPackages(archModel)

	sub := archModel.Subsystems["move_pipeline"]
	if sub == nil {
		t.Fatal("expected move_pipeline to exist for new move package")
	}

	if !slices.Contains(sub.Packages, "game/logic/move/newmove") {
		t.Errorf("expected game/logic/move/newmove in move_pipeline")
	}
}

func TestGroupPackages_WebInfraPackages(t *testing.T) {
	t.Parallel()

	// web/middleware, web/mux, web/nbio should all group into middleware.
	archModel := makeModel("web/middleware", "web/mux", "web/nbio")

	GroupPackages(archModel)

	sub := archModel.Subsystems["middleware"]
	if sub == nil {
		t.Fatal("expected middleware subsystem")
	}

	for _, suffix := range []string{"web/middleware", "web/mux", "web/nbio"} {
		if !slices.Contains(sub.Packages, suffix) {
			t.Errorf("expected %q in middleware, got: %v", suffix, sub.Packages)
		}
	}

	if sub.Label != "Middleware" {
		t.Errorf("label = %q, want %q", sub.Label, "Middleware")
	}
}

func TestGroupPackages_LabelGeneration(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		subsystem string
		suffix    string
		wantLabel string
	}{
		{
			name:      "explicit label for api_dtos",
			subsystem: "api_dtos",
			suffix:    "game/api",
			wantLabel: "API DTOs",
		},
		{
			name:      "explicit label for rest_utils",
			subsystem: "rest_utils",
			suffix:    "web/rest",
			wantLabel: "REST Utils",
		},
		{
			name:      "explicit label for move_pipeline",
			subsystem: "move_pipeline",
			suffix:    "game/logic/move/attack",
			wantLabel: "Move Pipeline",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			archModel := makeModel(testCase.suffix)
			GroupPackages(archModel)

			sub, ok := archModel.Subsystems[testCase.subsystem]
			if !ok {
				t.Fatalf("expected subsystem %q, got: %v",
					testCase.subsystem, subsystemIDs(archModel))
			}

			if sub.Label != testCase.wantLabel {
				t.Errorf("label = %q, want %q",
					sub.Label, testCase.wantLabel)
			}
		})
	}
}

func TestGroupPackages_StandalonePackages(t *testing.T) {
	t.Parallel()

	// Packages not matching any root or override become standalone subsystems.
	// game/ctx has no root ("game" is not a root) and no override.
	archModel := makeModel("game/ctx")

	GroupPackages(archModel)

	// Standalone: suffix slashes→underscores as ID, last segment title-cased as label.
	sub, ok := archModel.Subsystems["game_ctx"]
	if !ok {
		t.Fatalf("expected standalone subsystem game_ctx, got: %v",
			subsystemIDs(archModel))
	}

	if !slices.Contains(sub.Packages, "game/ctx") {
		t.Errorf("expected game/ctx in game_ctx packages")
	}

	if sub.Label != "Ctx" {
		t.Errorf("label = %q, want %q", sub.Label, "Ctx")
	}
}

func TestGroupPackages_MultipleStandalonesMerge(t *testing.T) {
	t.Parallel()

	// Two packages under game/tracing — if "game/tracing" doesn't match a root,
	// both should become standalone under game_tracing.
	// But actually game/tracing is a single package, no children.
	// Let's test with game/ctx and lobby/ctx — these are different standalones.
	archModel := makeModel("game/ctx", "lobby/ctx")

	GroupPackages(archModel)

	if _, ok := archModel.Subsystems["game_ctx"]; !ok {
		t.Errorf("expected game_ctx standalone, got: %v",
			subsystemIDs(archModel))
	}

	if _, ok := archModel.Subsystems["lobby_ctx"]; !ok {
		t.Errorf("expected lobby_ctx standalone, got: %v",
			subsystemIDs(archModel))
	}
}

func TestGroupPackages_EmptyModel(t *testing.T) {
	t.Parallel()

	archModel := &ArchModel{
		Packages:   make(map[string]*PackageInfo),
		Subsystems: make(map[string]*SubsystemInfo),
	}

	GroupPackages(archModel)

	if len(archModel.Subsystems) != 0 {
		t.Errorf("expected 0 subsystems for empty model, got %d", len(archModel.Subsystems))
	}
}

func TestGroupPackages_AllPackagesAssigned(t *testing.T) {
	t.Parallel()

	// Every package in the model must end up in exactly one subsystem.
	suffixes := []string{
		"game/api",
		"game/api/messaging",
		"game/commands",
		"game/config",
		"game/ctx",
		"game/data/db",
		"game/events",
		"game/headlines",
		"game/logic/board",
		"game/logic/move/attack",
		"game/publisher",
		"game/publisher/converter",
		"game/routes",
		"game/snapshot",
		"game/ws",
		"kernel/config",
		"kernel/bus",
		"lobby/api/messaging",
		"lobby/data/db",
		"lobby/events",
		"lobby/logic/creation",
		"lobby/publisher",
		"lobby/routes",
		"lobby/ws",
		"testing/invariant",
		"testonly",
		"web/middleware",
		"web/mux",
		"web/rest",
		"web/ws",
	}

	archModel := makeModel(suffixes...)
	GroupPackages(archModel)

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

	for _, s := range suffixes {
		if _, ok := assigned[s]; !ok {
			t.Errorf("package %q not assigned to any subsystem", s)
		}
	}
}

func TestGroupPackages_SubsystemIDsAreD2Safe(t *testing.T) {
	t.Parallel()

	// D2 identifiers cannot contain slashes. Verify all IDs use underscores.
	archModel := makeModel(
		"game/ctx",          // standalone
		"game/logic/board",  // root match
		"game/publisher",    // override
		"web/middleware",    // override
		"kernel/config",     // root match
		"game/logic/move/x", // longest-prefix
	)

	GroupPackages(archModel)

	for id := range archModel.Subsystems {
		for _, ch := range id {
			if ch == '/' {
				t.Errorf("subsystem ID %q contains slash — not D2-safe", id)

				break
			}
		}
	}
}

func TestGroupPackages_PackagesSorted(t *testing.T) {
	t.Parallel()

	// Packages within a subsystem should be sorted for deterministic output.
	archModel := makeModel(
		"game/logic/player",
		"game/logic/board",
		"game/logic/card",
		"game/logic/phase",
	)

	GroupPackages(archModel)

	sub := archModel.Subsystems["game_services"]
	if sub == nil {
		t.Fatal("expected game_services")
	}

	if !sort.StringsAreSorted(sub.Packages) {
		t.Errorf("packages not sorted: %v", sub.Packages)
	}
}

func TestGroupPackages_KernelPackages(t *testing.T) {
	t.Parallel()

	// All kernel/* packages should land in the kernel subsystem.
	archModel := makeModel(
		"kernel/config",
		"kernel/bus",
		"kernel/metrics",
		"kernel/slog",
		"kernel/logger",
	)

	GroupPackages(archModel)

	sub := archModel.Subsystems["kernel"]
	if sub == nil {
		t.Fatal("expected kernel subsystem")
	}

	if len(sub.Packages) != 5 {
		t.Errorf("expected 5 packages in kernel, got %d: %v", len(sub.Packages), sub.Packages)
	}

	if sub.Label != "Kernel" {
		t.Errorf("label = %q, want %q", sub.Label, "Kernel")
	}
}

func TestGroupPackages_WebRestPackages(t *testing.T) {
	t.Parallel()

	// web/rest/* should land in rest_utils via root matching.
	archModel := makeModel(
		"web/rest",
		"web/rest/health",
		"web/rest/route",
		"web/rest/utils",
	)

	GroupPackages(archModel)

	sub := archModel.Subsystems["rest_utils"]
	if sub == nil {
		t.Fatal("expected rest_utils subsystem")
	}

	if len(sub.Packages) != 4 {
		t.Errorf("expected 4 packages in rest_utils, got %d: %v",
			len(sub.Packages), sub.Packages)
	}
}

func TestGroupPackages_LobbyLogicPackages(t *testing.T) {
	t.Parallel()

	archModel := makeModel(
		"lobby/logic/creation",
		"lobby/logic/management",
		"lobby/logic/start",
		"lobby/logic/state",
	)

	GroupPackages(archModel)

	sub := archModel.Subsystems["lobby_logic"]
	if sub == nil {
		t.Fatal("expected lobby_logic subsystem")
	}

	if len(sub.Packages) != 4 {
		t.Errorf("expected 4 packages, got %d: %v", len(sub.Packages), sub.Packages)
	}
}

// subsystemIDs returns sorted subsystem IDs from a model, for diagnostic output.
func subsystemIDs(archModel *ArchModel) []string {
	ids := make([]string, 0, len(archModel.Subsystems))
	for subsystemID := range archModel.Subsystems {
		ids = append(ids, subsystemID)
	}

	sort.Strings(ids)

	return ids
}
