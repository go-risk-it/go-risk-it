package model

import (
	"sort"
	"strings"
)

// SubsystemRoots maps directory prefixes to subsystem IDs.
// Longest prefix match wins. All packages under a root inherit its group.
//
//nolint:gochecknoglobals // package-level lookup table for subsystem classification
var SubsystemRoots = map[string]string{
	"game/api":                 "api_dtos",
	"game/internal/data":       "game_data",
	"game/internal/logic":      "game_services",
	"game/internal/logic/move": "move_pipeline", // longest-prefix wins over game/internal/logic
	"game/consumers":           "game_consumers",
	"kernel/data":              "kernel_data",
	"lobby/api":                "api_dtos",
	"lobby/data":               "lobby_data",
	"lobby/logic":              "lobby_logic",
	"testing":                  "testing",
	"web/rest":                 "rest_utils",
}

// SubsystemOverrides maps individual package suffixes to subsystem IDs.
// Checked before roots. Use this for packages that don't follow their
// directory prefix's natural grouping.
//
// Edge cases documented:
//   - game/consumers → game_consumers (not a root child — consumers is a top-level game module)
//   - game/config, game/headlines, game/snapshot → game_support (cross-cutting game modules)
//   - game/ctx, game/rand, game/tracing → game_services (tiny game packages consolidated)
//   - game/routes, game/ws → game_consumers (web-facing game infra)
//   - kernel/* → kernel_{bus,config,ctx,errors,observability,utils} (meaningful sub-subsystems)
//   - web/middleware, web/mux, web/nbio → middleware (web infra, not matched by web/rest root)
//   - web/ws → websocket (shared WS infrastructure, distinct from game/ws or lobby/ws)
//   - lobby/consumers, lobby/routes, lobby/ws → lobby_consumers (web-facing lobby infra)
//   - testonly → testing (top-level test helper, no "testing" prefix)
//
//nolint:gochecknoglobals // package-level lookup table for subsystem overrides
var SubsystemOverrides = map[string]string{
	"game/internal/config":      "game_support",
	"game/ctx":                  "game_services",
	"game/events":               "game_services",
	"game/internal/handlers":    "game_support",
	"game/internal/rand":        "game_services",
	"game/internal/snapshot":    "game_support",
	"game/tracing":              "game_services",
	"game/routes":               "game_consumers",
	"game/ws":                   "game_consumers",
	"kernel/bus":                "kernel_bus",
	"kernel/config":             "kernel_config",
	"kernel/ctx":                "kernel_ctx",
	"kernel/errors":             "kernel_errors",
	"kernel/logger":             "kernel_observability",
	"kernel/metrics":            "kernel_observability",
	"kernel/observe":            "kernel_observability",
	"kernel/otelsetup":          "kernel_observability",
	"kernel/slog":               "kernel_observability",
	"kernel/upgradablerw_mutex": "kernel_utils",
	"lobby/consumers":           "lobby_consumers",
	"lobby/routes":              "lobby_consumers",
	"lobby/ws":                  "lobby_consumers",
	"web/middleware":            "middleware",
	"web/mux":                   "middleware",
	"web/nbio":                  "middleware",
	"web/ws":                    "websocket",
	"testonly":                  "testing",
}

// SubsystemLabels maps subsystem IDs to human-readable display labels.
// Every ID that appears in SubsystemRoots or SubsystemOverrides must have
// an entry here. Standalone subsystems generate their own labels.
//
//nolint:gochecknoglobals // package-level lookup table for subsystem labels
var SubsystemLabels = map[string]string{
	"api_dtos":             "API DTOs",
	"game_data":            "Game Data",
	"game_consumers":       "Game Consumers",
	"game_services":        "Game Services",
	"game_support":         "Game Support",
	"kernel_bus":           "Event Bus",
	"kernel_config":        "Config",
	"kernel_ctx":           "Context",
	"kernel_data":          "Data",
	"kernel_errors":        "Errors",
	"kernel_observability": "Observability",
	"kernel_utils":         "Utils",
	"lobby_data":           "Lobby Data",
	"lobby_logic":          "Lobby Logic",
	"lobby_consumers":      "Lobby Consumers",
	"middleware":           "Middleware",
	"move_pipeline":        "Move Pipeline",
	"rest_utils":           "REST Utils",
	"testing":              "Testing",
	"websocket":            "WebSocket",
}

// sortedRoots holds SubsystemRoots keys sorted longest-first for prefix matching.
// Computed once at init time.
//
//nolint:gochecknoglobals // derived from SubsystemRoots at init
var sortedRoots []string

//nolint:gochecknoinits // pre-sorts roots for longest-prefix matching
func init() {
	sortedRoots = make([]string, 0, len(SubsystemRoots))
	for prefix := range SubsystemRoots {
		sortedRoots = append(sortedRoots, prefix)
	}

	sort.Slice(sortedRoots, func(i, j int) bool {
		return len(sortedRoots[i]) > len(sortedRoots[j])
	})
}

// GroupPackages assigns packages from m.Packages to subsystems, populating m.Subsystems.
//
// Resolution order per package suffix:
//  1. SubsystemOverrides — exact match on the full suffix.
//  2. SubsystemRoots — longest prefix match (suffix equals or starts with root+"/").
//  3. Standalone — suffix with slashes replaced by underscores; label is the last
//     path segment with first letter uppercased.
func GroupPackages(archModel *ArchModel) {
	// Collect suffixes and sort for deterministic iteration.
	suffixes := make([]string, 0, len(archModel.Packages))
	for suffix := range archModel.Packages {
		suffixes = append(suffixes, suffix)
	}

	sort.Strings(suffixes)

	for _, suffix := range suffixes {
		subsystemID := resolveSubsystem(suffix)
		ensureSubsystem(archModel, subsystemID)
		archModel.Subsystems[subsystemID].Packages = append(
			archModel.Subsystems[subsystemID].Packages, suffix,
		)
	}

	// Sort packages within each subsystem for deterministic output.
	for _, sub := range archModel.Subsystems {
		sort.Strings(sub.Packages)
	}
}

// resolveSubsystem determines the subsystem ID for a package suffix.
// Checks overrides first, then longest-prefix root match, then falls back to standalone.
func resolveSubsystem(suffix string) string {
	// 1. Exact override match.
	if id, ok := SubsystemOverrides[suffix]; ok {
		return id
	}

	// 2. Longest prefix match against roots.
	for _, root := range sortedRoots {
		if suffix == root || strings.HasPrefix(suffix, root+"/") {
			return SubsystemRoots[root]
		}
	}

	// 3. Standalone: slashes → underscores.
	return strings.ReplaceAll(suffix, "/", "_")
}

// ensureSubsystem creates a SubsystemInfo entry if it doesn't exist yet.
func ensureSubsystem(archModel *ArchModel, subsystemID string) {
	if archModel.Subsystems[subsystemID] != nil {
		return
	}

	archModel.Subsystems[subsystemID] = &SubsystemInfo{
		ID:    subsystemID,
		Label: subsystemLabel(subsystemID),
	}
}

// subsystemLabel returns the display label for a subsystem ID.
// Uses the explicit label map for known IDs; generates a label from the
// last path-like segment for standalone subsystems.
func subsystemLabel(subsystemID string) string {
	if label, ok := SubsystemLabels[subsystemID]; ok {
		return label
	}

	// Standalone: use last segment of the original suffix (underscores as separators).
	// The ID was produced by replacing slashes with underscores, so split on "_"
	// and take the last segment, then title-case the first letter.
	parts := strings.Split(subsystemID, "_")
	last := parts[len(parts)-1]

	return titleCase(last)
}

// titleCase uppercases the first letter of s. Returns s unchanged if empty.
func titleCase(s string) string {
	if s == "" {
		return s
	}

	return strings.ToUpper(s[:1]) + s[1:]
}
