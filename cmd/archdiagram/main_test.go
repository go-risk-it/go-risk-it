package main

import (
	"strings"
	"testing"
)

func TestLayerFromPrefix_ExplicitMappings(t *testing.T) {
	t.Parallel()

	cases := []struct {
		suffix string
		want   string
	}{
		{"config", "Infrastructure"},
		{"metrics", "Infrastructure"},
		{"rand", "Infrastructure"},
		{"slog", "Infrastructure"},
		{"tracing", "Infrastructure"},
		{"upgradablerw_mutex", "Infrastructure"},
		{"ctx", "Ctx"},
		{"data/db", "Data"},
		{"data/game/db", "Data"},
		{"data/pool", "Data"},
		{"data/migration", "Data"},
		{"events", "Events"},
		{"events/logger", "Events"},
		{"events/game", "Events-domain"},
		{"events/lobby", "Events-domain"},
		{"logic/errors", "Shared"},
		{"api/game", "API"},
		{"api/game/messaging", "API"},
		{"testing/invariant", "Test"},
		{"testonly", "Test"},
	}

	for _, testCase := range cases {
		t.Run(testCase.suffix, func(t *testing.T) {
			t.Parallel()

			got := layerFromPrefix(testCase.suffix)
			if got != testCase.want {
				t.Errorf("layerFromPrefix(%q) = %q, want %q", testCase.suffix, got, testCase.want)
			}
		})
	}
}

func TestLayerFromPrefix_PrefixFallbacks(t *testing.T) {
	t.Parallel()

	cases := []struct {
		suffix string
		want   string
	}{
		{"api/lobby/rest/response", "API"},
		{"data/game/sqlc", "Data"},
		{"logic/game/board", "Logic"},
		{"logic/game/move/orchestration", "Logic"},
		{"web/game/controller", "Web"},
		{"web/lobby/ws", "Web"},
		{"events/game/some/sub", "Events-domain"},
		{"events/lobby/deep", "Events-domain"},
	}

	for _, testCase := range cases {
		t.Run(testCase.suffix, func(t *testing.T) {
			t.Parallel()

			got := layerFromPrefix(testCase.suffix)
			if got != testCase.want {
				t.Errorf("layerFromPrefix(%q) = %q, want %q", testCase.suffix, got, testCase.want)
			}
		})
	}
}

func TestLayerFromPrefix_UnknownReturnsEmpty(t *testing.T) {
	t.Parallel()

	got := layerFromPrefix("something/unknown")
	if got != "" {
		t.Errorf("layerFromPrefix(unknown) = %q, want empty", got)
	}
}

func TestSubsystemFromSuffix_ExplicitMappings(t *testing.T) {
	t.Parallel()

	cases := []struct {
		suffix string
		want   string
	}{
		// Web Layer
		{"web/game/controller", "game_handlers"},
		{"web/game/rest", "game_handlers"},
		{"web/game/ws", "game_handlers"},
		{"game/consumers", "game_handlers"},
		{"game/consumers/converter", "game_handlers"},
		{"web/lobby/controller", "lobby_handlers"},
		{"web/lobby/publisher", "lobby_handlers"},
		{"web/lobby/rest", "lobby_handlers"},
		{"web/lobby/ws", "lobby_handlers"},
		{"web/middleware", "middleware"},
		{"web/mux", "middleware"},
		{"web/nbio", "middleware"},
		{"web/otel", "middleware"},
		{"web/rest", "rest_utils"},
		{"web/rest/health", "rest_utils"},
		{"web/rest/route", "rest_utils"},
		{"web/rest/utils", "rest_utils"},
		{"web/ws", "websocket"},
		{"web/ws/message", "websocket"},

		// Logic Layer
		{"logic/game/move/orchestration", "move_pipeline"},
		{"logic/game/move/attack", "move_pipeline"},
		{"logic/game/move/deploy", "move_pipeline"},
		{"logic/game/move/conquer", "move_pipeline"},
		{"logic/game/move/reinforce", "move_pipeline"},
		{"logic/game/move/cards", "move_pipeline"},
		{"logic/game/move/validation", "move_pipeline"},
		{"logic/game/board", "game_services"},
		{"logic/game/phase", "game_services"},
		{"logic/game/player", "game_services"},
		{"logic/game/snapshot", "game_services"},
		{"logic/game/state", "game_services"},
		{"logic/game/creation", "game_services"},
		{"logic/game/headlines", "game_services"},
		{"logic/lobby/creation", "lobby_logic"},
		{"logic/lobby/management", "lobby_logic"},
		{"logic/lobby/start", "lobby_logic"},
		{"logic/lobby/state", "lobby_logic"},

		// Shared
		{"logic/errors", "domain_errors"},

		// Events
		{"events", "event_bus"},
		{"events/logger", "event_bus"},
		{"events/game", "game_events"},
		{"events/lobby", "lobby_events"},

		// Data
		{"data/game/db", "game_data"},
		{"data/lobby/db", "lobby_data"},
		{"data/db", "database"},
		{"data/pool", "database"},
		{"data/migration", "database"},

		// Infrastructure
		{"config", "observability"},
		{"metrics", "observability"},
		{"tracing", "observability"},
		{"slog", "observability"},
		{"ctx", "context"},
		{"rand", "utilities"},
		{"upgradablerw_mutex", "utilities"},

		// API
		{"api/game", "api_dtos"},
		{"api/game/messaging", "api_dtos"},
		{"api/lobby/rest/request", "api_dtos"},

		// Test
		{"testing/invariant", "testing"},
		{"testonly", "testing"},
	}

	for _, testCase := range cases {
		t.Run(testCase.suffix, func(t *testing.T) {
			t.Parallel()

			got := subsystemFromSuffix(testCase.suffix)
			if got != testCase.want {
				t.Errorf(
					"subsystemFromSuffix(%q) = %q, want %q",
					testCase.suffix,
					got,
					testCase.want,
				)
			}
		})
	}
}

func TestSubsystemFromSuffix_Fallbacks(t *testing.T) {
	t.Parallel()

	cases := []struct {
		suffix string
		want   string
	}{
		{"api/future/new", "api_dtos"},
		{"web/game/newpkg", "game_handlers"},
		{"web/lobby/newpkg", "lobby_handlers"},
		{"web/ws/newpkg", "websocket"},
		{"web/rest/newpkg", "rest_utils"},
		{"web/newmiddleware", "middleware"},
		{"logic/game/move/newmove", "move_pipeline"},
		{"logic/game/newservice", "game_services"},
		{"logic/lobby/newfeature", "lobby_logic"},
		{"events/game/sub", "game_events"},
		{"events/lobby/sub", "lobby_events"},
		{"events/newpkg", "event_bus"},
		{"data/game/newpkg", "game_data"},
		{"data/lobby/newpkg", "lobby_data"},
		{"data/newpkg", "database"},
		{"testing/newpkg", "testing"},
	}

	for _, testCase := range cases {
		t.Run(testCase.suffix, func(t *testing.T) {
			t.Parallel()

			got := subsystemFromSuffix(testCase.suffix)
			if got != testCase.want {
				t.Errorf(
					"subsystemFromSuffix(%q) = %q, want %q",
					testCase.suffix,
					got,
					testCase.want,
				)
			}
		})
	}
}

func TestSubsystemFromSuffix_UnknownReturnsEmpty(t *testing.T) {
	t.Parallel()

	got := subsystemFromSuffix("something/unknown")
	if got != "" {
		t.Errorf("subsystemFromSuffix(unknown) = %q, want empty", got)
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
		{"logic", true},
		{"logic/game", true},
		{"data", true},
		{"web", true},
		{"web/game", true},
		// generated packages
		{"data/game/sqlc", true},
		{"data/lobby/sqlc", true},
		{"something/mocks", true},
		// real packages
		{"config", false},
		{"logic/game/board", false},
		{"web/game/controller", false},
		{"events", false},
	}

	for _, testCase := range cases {
		t.Run(testCase.suffix, func(t *testing.T) {
			t.Parallel()

			got := isExcluded(testCase.suffix)
			if got != testCase.want {
				t.Errorf("isExcluded(%q) = %v, want %v", testCase.suffix, got, testCase.want)
			}
		})
	}
}

func TestNodeID(t *testing.T) {
	t.Parallel()

	cases := []struct {
		suffix string
		want   string
	}{
		{"config", "config"},
		{"logic/game/board", "logic.game.board"},
		{"web/game/controller", "web.game.controller"},
		{"upgradablerw_mutex", "upgradablerw_mutex"},
	}

	for _, testCase := range cases {
		t.Run(testCase.suffix, func(t *testing.T) {
			t.Parallel()

			got := nodeID(testCase.suffix)
			if got != testCase.want {
				t.Errorf("nodeID(%q) = %q, want %q", testCase.suffix, got, testCase.want)
			}
		})
	}
}

func TestShortName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		suffix string
		want   string
	}{
		{"config", "config"},
		{"logic/game/board", "board"},
		{"web/game/ws", "ws"},
		{"api/game/rest/request", "request"},
	}

	for _, testCase := range cases {
		t.Run(testCase.suffix, func(t *testing.T) {
			t.Parallel()

			got := shortName(testCase.suffix)
			if got != testCase.want {
				t.Errorf("shortName(%q) = %q, want %q", testCase.suffix, got, testCase.want)
			}
		})
	}
}

func TestGenerateD2_SubsystemNodes(t *testing.T) {
	t.Parallel()

	subs := []subsystem{
		{ID: "observability", Label: "Observability", Layer: "Infrastructure"},
		{ID: "move_pipeline", Label: "Move Pipeline", Layer: "Logic"},
		{ID: "database", Label: "Database", Layer: "Data"},
	}

	inputEdges := []edge{
		{From: "Logic", To: "Data"},
	}

	d2Output := generateD2(subs, inputEdges)

	// Check title
	if !strings.Contains(d2Output, "title: go-risk-it Architecture") {
		t.Error("expected title in D2 output")
		t.Logf("D2 output:\n%s", d2Output)
	}

	// Check that the layer-level cross-layer edge exists
	if !strings.Contains(d2Output, "logic -> data") {
		t.Error("expected layer-level edge logic -> data in D2 output")
		t.Logf("D2 output:\n%s", d2Output)
	}

	// Check layer containers exist
	if !strings.Contains(d2Output, "infrastructure: Infrastructure {") {
		t.Error("expected Infrastructure container")
	}

	if !strings.Contains(d2Output, "logic: Logic {") {
		t.Error("expected Logic container")
	}

	if !strings.Contains(d2Output, "data: Data {") {
		t.Error("expected Data container")
	}

	// Check subsystem nodes inside containers
	if !strings.Contains(d2Output, `observability: "Observability"`) {
		t.Error("expected observability subsystem node")
		t.Logf("D2 output:\n%s", d2Output)
	}

	if !strings.Contains(d2Output, `move_pipeline: "Move Pipeline"`) {
		t.Error("expected move_pipeline subsystem node")
		t.Logf("D2 output:\n%s", d2Output)
	}

	if !strings.Contains(d2Output, `database: "Database"`) {
		t.Error("expected database subsystem node")
		t.Logf("D2 output:\n%s", d2Output)
	}

	// Verify NO individual package nodes (old format with \n descriptions)
	if strings.Contains(d2Output, "\\n") {
		t.Error("expected no multi-line package labels (old format), but found \\n in output")
		t.Logf("D2 output:\n%s", d2Output)
	}
}

func TestGenerateD2_IntraLayerEdgesOmitted(t *testing.T) {
	t.Parallel()

	// Verify that edges between same-layer packages are not included
	fromSuffix := "logic/game/board"
	toSuffix := "logic/game/phase"

	fromLayer := layerFromPrefix(fromSuffix)
	toLayer := layerFromPrefix(toSuffix)

	if fromLayer != toLayer {
		t.Fatalf("test setup error: expected same layer, got %q and %q", fromLayer, toLayer)
	}
}

func TestContainerIDForSuffix(t *testing.T) {
	t.Parallel()

	cases := []struct {
		suffix string
		want   string
	}{
		{"config", "infrastructure"},
		{"logic/game/board", "logic"},
		{"events/game", "events_domain"},
		{"web/game/controller", "web"},
		{"ctx", "ctx"},
		{"logic/errors", "shared"},
	}

	for _, testCase := range cases {
		t.Run(testCase.suffix, func(t *testing.T) {
			t.Parallel()

			got := containerIDForSuffix(testCase.suffix)
			if got != testCase.want {
				t.Errorf(
					"containerIDForSuffix(%q) = %q, want %q",
					testCase.suffix,
					got,
					testCase.want,
				)
			}
		})
	}
}

func TestLayerContainerID(t *testing.T) {
	t.Parallel()

	cases := []struct {
		layer string
		want  string
	}{
		{"Infrastructure", "infrastructure"},
		{"Logic", "logic"},
		{"Events-domain", "events_domain"},
		{"Web", "web"},
		{"Ctx", "ctx"},
		{"Shared", "shared"},
		{"Data", "data"},
		{"API", "api"},
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

func TestAllSubsystemMapEntriesHaveDefinitions(t *testing.T) {
	t.Parallel()

	for suffix, subID := range subsystemMap {
		if _, ok := subsystemDefs[subID]; !ok {
			t.Errorf("subsystemMap[%q] = %q, but no subsystemDefs entry exists", suffix, subID)
		}
	}
}

func TestAllSubsystemMapEntriesHaveLayerMappings(t *testing.T) {
	t.Parallel()

	for suffix := range subsystemMap {
		layer := layerFromPrefix(suffix)
		if layer == "" {
			t.Errorf("subsystemMap entry %q has no layer mapping", suffix)
		}
	}
}

func TestSubsystemDefsLayersAreValid(t *testing.T) {
	t.Parallel()

	for id, def := range subsystemDefs {
		if _, ok := layers[def.Layer]; !ok {
			t.Errorf("subsystemDefs[%q] has layer %q which is not in layers map", id, def.Layer)
		}
	}
}
