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
		"game/internal/logic/board",
		"game/internal/logic/card",
		"game/internal/logic/phase",
		"game/internal/logic/player",
	)

	GroupPackages(archModel)

	sub, ok := archModel.Subsystems["game_services"]
	if !ok {
		t.Fatal("expected subsystem game_services to exist")
	}

	for _, suffix := range []string{
		"game/internal/logic/board",
		"game/internal/logic/card",
		"game/internal/logic/phase",
		"game/internal/logic/player",
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
		"game/internal/logic/board",
		"game/internal/logic/move/attack",
		"game/internal/logic/move/deploy",
		"game/internal/logic/move/orchestration",
	)

	GroupPackages(archModel)

	moveSub := archModel.Subsystems["move_pipeline"]
	if moveSub == nil {
		t.Fatal("expected subsystem move_pipeline to exist")

		return
	}

	for _, suffix := range []string{
		"game/internal/logic/move/attack",
		"game/internal/logic/move/deploy",
		"game/internal/logic/move/orchestration",
	} {
		if !slices.Contains(moveSub.Packages, suffix) {
			t.Errorf("expected %q in move_pipeline, got: %v", suffix, moveSub.Packages)
		}
	}

	// board should be in game_services, not move_pipeline
	gameSub := archModel.Subsystems["game_services"]
	if gameSub == nil {
		t.Fatal("expected subsystem game_services to exist")

		return
	}

	if !slices.Contains(gameSub.Packages, "game/internal/logic/board") {
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
			name:      "game/consumers routes to game_consumers",
			suffix:    "game/consumers",
			wantSubID: "game_consumers",
			wantLabel: "Game Consumers",
		},
		{
			name:      "game/internal/config routes to game_support",
			suffix:    "game/internal/config",
			wantSubID: "game_support",
			wantLabel: "Game Support",
		},
		{
			name:      "game/ctx routes to game_services",
			suffix:    "game/ctx",
			wantSubID: "game_services",
			wantLabel: "Game Services",
		},
		{
			name:      "game/internal/handlers routes to game_support",
			suffix:    "game/internal/handlers",
			wantSubID: "game_support",
			wantLabel: "Game Support",
		},
		{
			name:      "game/internal/rand routes to game_services",
			suffix:    "game/internal/rand",
			wantSubID: "game_services",
			wantLabel: "Game Services",
		},
		{
			name:      "game/internal/snapshot routes to game_support",
			suffix:    "game/internal/snapshot",
			wantSubID: "game_support",
			wantLabel: "Game Support",
		},
		{
			name:      "game/tracing routes to game_services",
			suffix:    "game/tracing",
			wantSubID: "game_services",
			wantLabel: "Game Services",
		},
		{
			name:      "kernel/bus routes to kernel_bus",
			suffix:    "kernel/bus",
			wantSubID: "kernel_bus",
			wantLabel: "Event Bus",
		},
		{
			name:      "kernel/config routes to kernel_config",
			suffix:    "kernel/config",
			wantSubID: "kernel_config",
			wantLabel: "Config",
		},
		{
			name:      "kernel/logger routes to kernel_observability",
			suffix:    "kernel/logger",
			wantSubID: "kernel_observability",
			wantLabel: "Observability",
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
			name:      "lobby/consumers routes to lobby_consumers",
			suffix:    "lobby/consumers",
			wantSubID: "lobby_consumers",
			wantLabel: "Lobby Consumers",
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
	archModel := makeModel(
		"game/internal/config",
		"game/internal/snapshot",
		"game/internal/logic/board",
	)

	GroupPackages(archModel)

	support := archModel.Subsystems["game_support"]
	if support == nil {
		t.Fatal("expected subsystem game_support")

		return
	}

	if !slices.Contains(support.Packages, "game/internal/config") {
		t.Errorf("expected game/config in game_support")
	}

	if !slices.Contains(support.Packages, "game/internal/snapshot") {
		t.Errorf("expected game/snapshot in game_support")
	}

	// game/logic/board should NOT be in game_support
	if slices.Contains(support.Packages, "game/internal/logic/board") {
		t.Error("game/internal/logic/board should not be in game_support")
	}
}

func TestGroupPackages_NewPackage(t *testing.T) {
	t.Parallel()

	// A new package under game/logic/ should automatically land in game_services
	// without any map update.
	archModel := makeModel("game/internal/logic/foo", "game/internal/logic/board")

	GroupPackages(archModel)

	sub := archModel.Subsystems["game_services"]
	if sub == nil {
		t.Fatal("expected game_services to exist for synthetic package")

		return
	}

	if !slices.Contains(sub.Packages, "game/internal/logic/foo") {
		t.Errorf("expected game/logic/foo in game_services, got: %v", sub.Packages)
	}
}

func TestGroupPackages_NewMovePackage(t *testing.T) {
	t.Parallel()

	// A new package under game/logic/move/ should land in move_pipeline.
	archModel := makeModel("game/internal/logic/move/newmove")

	GroupPackages(archModel)

	sub := archModel.Subsystems["move_pipeline"]
	if sub == nil {
		t.Fatal("expected move_pipeline to exist for new move package")

		return
	}

	if !slices.Contains(sub.Packages, "game/internal/logic/move/newmove") {
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

		return
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
			suffix:    "game/internal/logic/move/attack",
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
	// lobby/ctx has no root ("lobby" is not a root for this prefix) and no override.
	archModel := makeModel("lobby/ctx")

	GroupPackages(archModel)

	// Standalone: suffix slashes→underscores as ID, last segment title-cased as label.
	sub, ok := archModel.Subsystems["lobby_ctx"]
	if !ok {
		t.Fatalf("expected standalone subsystem lobby_ctx, got: %v",
			subsystemIDs(archModel))
	}

	if !slices.Contains(sub.Packages, "lobby/ctx") {
		t.Errorf("expected lobby/ctx in lobby_ctx packages")
	}

	if sub.Label != "Ctx" {
		t.Errorf("label = %q, want %q", sub.Label, "Ctx")
	}
}

func TestGroupPackages_MultipleStandalonesMerge(t *testing.T) {
	t.Parallel()

	// game/ctx is now overridden to game_services, but lobby/ctx is still standalone.
	// Test that two different standalone packages get separate subsystems.
	archModel := makeModel("lobby/ctx", "lobby/events")

	GroupPackages(archModel)

	if _, ok := archModel.Subsystems["lobby_ctx"]; !ok {
		t.Errorf("expected lobby_ctx standalone, got: %v",
			subsystemIDs(archModel))
	}

	if _, ok := archModel.Subsystems["lobby_events"]; !ok {
		t.Errorf("expected lobby_events standalone, got: %v",
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
		"game/internal/config",
		"game/ctx",
		"game/internal/data/db",
		"game/events",
		"game/internal/handlers",
		"game/internal/logic/board",
		"game/internal/logic/move/attack",
		"game/consumers",
		"game/routes",
		"game/internal/snapshot",
		"game/ws",
		"kernel/config",
		"kernel/bus",
		"lobby/api/messaging",
		"lobby/data/db",
		"lobby/events",
		"lobby/logic/creation",
		"lobby/consumers",
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
		"lobby/ctx",                  // standalone
		"game/internal/logic/board",  // root match
		"game/consumers",             // override
		"web/middleware",             // override
		"kernel/config",              // override → kernel_config
		"game/internal/logic/move/x", // longest-prefix
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
		"game/internal/logic/player",
		"game/internal/logic/board",
		"game/internal/logic/card",
		"game/internal/logic/phase",
	)

	GroupPackages(archModel)

	sub := archModel.Subsystems["game_services"]
	if sub == nil {
		t.Fatal("expected game_services")

		return
	}

	if !sort.StringsAreSorted(sub.Packages) {
		t.Errorf("packages not sorted: %v", sub.Packages)
	}
}

func TestGroupPackages_KernelPackages(t *testing.T) {
	t.Parallel()

	// Kernel packages now split into meaningful sub-subsystems via overrides.
	archModel := makeModel(
		"kernel/config",
		"kernel/bus",
		"kernel/metrics",
		"kernel/observe",
		"kernel/slog",
		"kernel/logger",
	)

	GroupPackages(archModel)

	// kernel/config → kernel_config
	configSub := archModel.Subsystems["kernel_config"]
	if configSub == nil {
		t.Fatal("expected kernel_config subsystem")

		return
	}

	if !slices.Contains(configSub.Packages, "kernel/config") {
		t.Errorf("expected kernel/config in kernel_config, got: %v", configSub.Packages)
	}

	if configSub.Label != "Config" {
		t.Errorf("kernel_config label = %q, want %q", configSub.Label, "Config")
	}

	// kernel/bus → kernel_bus
	busSub := archModel.Subsystems["kernel_bus"]
	if busSub == nil {
		t.Fatal("expected kernel_bus subsystem")

		return
	}

	if !slices.Contains(busSub.Packages, "kernel/bus") {
		t.Errorf("expected kernel/bus in kernel_bus, got: %v", busSub.Packages)
	}

	if busSub.Label != "Event Bus" {
		t.Errorf("kernel_bus label = %q, want %q", busSub.Label, "Event Bus")
	}

	// kernel/metrics, kernel/observe, kernel/slog, kernel/logger → kernel_observability
	obsSub := archModel.Subsystems["kernel_observability"]
	if obsSub == nil {
		t.Fatal("expected kernel_observability subsystem")

		return
	}

	if len(obsSub.Packages) != 4 {
		t.Errorf("expected 4 packages in kernel_observability, got %d: %v",
			len(obsSub.Packages), obsSub.Packages)
	}

	if obsSub.Label != "Observability" {
		t.Errorf("kernel_observability label = %q, want %q", obsSub.Label, "Observability")
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

		return
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

		return
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
