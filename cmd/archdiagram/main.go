// Package main generates an architecture diagram from the go-risk-it internal packages.
//
// It reads all internal packages via `go list -json`, groups them into ~20 subsystem
// nodes within architectural layer containers, and produces a D2 diagram with
// cross-layer dependency edges. It then shells out to the `d2` CLI to render the SVG.
//
// Usage:
//
//	go run ./cmd/archdiagram/ [-output <dir>] [-d2 <path>]
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const modulePrefix = "github.com/go-risk-it/go-risk-it/internal/"

// goPackage is the subset of `go list -json` fields we need.
type goPackage struct {
	ImportPath string   `json:"ImportPath"` //nolint:tagliatelle // matches go list -json output
	Imports    []string `json:"Imports"`    //nolint:tagliatelle // matches go list -json output
	Dir        string   `json:"Dir"`        //nolint:tagliatelle // matches go list -json output
	GoFiles    []string `json:"GoFiles"`    //nolint:tagliatelle // matches go list -json output
}

// layerInfo describes a layer's display properties for the D2 diagram.
type layerInfo struct {
	Name  string // display name in the container
	Color string // fill color (hex)
	Order int    // sort order for consistent output
}

// layers defines the architectural layers and their visual properties.
// Colors are from the spec; order determines container placement in the D2 source.
var layers = map[string]layerInfo{
	"API":            {Name: "API", Color: "#E8EAF6", Order: 0},
	"Infrastructure": {Name: "Infrastructure", Color: "#F3E5F5", Order: 1},
	"Ctx":            {Name: "Ctx", Color: "#E0F7FA", Order: 2},
	"Data":           {Name: "Data", Color: "#FFF3E0", Order: 3},
	"Events":         {Name: "Events", Color: "#FCE4EC", Order: 4},
	"Events-domain":  {Name: "Events-domain", Color: "#FCE4EC", Order: 5},
	"Shared":         {Name: "Shared", Color: "#FFF9C4", Order: 6},
	"Logic":          {Name: "Logic", Color: "#E8F5E9", Order: 7},
	"Web":            {Name: "Web", Color: "#E3F2FD", Order: 8},
	"Test":           {Name: "Test", Color: "#F5F5F5", Order: 9},
}

// wiringRoots are single-file fx.Module aggregation packages excluded from the diagram.
// Matches the set in arch_test.go.
var wiringRoots = map[string]bool{
	"":                        true,
	"logic":                   true,
	"logic/game":              true,
	"logic/game/move":         true,
	"logic/game/move/service": true,
	"lobby/logic":             true,
	"data":                    true,
	"data/game":               true,
	"lobby/data":              true,
	"web":                     true,
	"web/game":                true,
	"web/lobby":               true,
}

// subsystem holds the display info for a subsystem group in the diagram.
type subsystem struct {
	ID    string // D2-safe identifier (e.g., "move_pipeline")
	Label string // display label (e.g., "Move Pipeline")
	Layer string // architectural layer key
}

// subsystemMap maps package suffixes to their subsystem ID.
// Each subsystem groups several related packages into a single node.
//
//nolint:gochecknoglobals // package-level lookup table for subsystem classification
var subsystemMap = map[string]string{
	// Web Layer — Game Handlers
	"web/game/controller": "game_handlers",
	"web/game/rest":       "game_handlers",
	"web/game/ws":         "game_handlers",

	// Game Publisher (event handlers → WS delivery)
	"game/publisher":           "game_handlers",
	"game/publisher/converter": "game_handlers",

	// Web Layer — Lobby Handlers
	"web/lobby/controller": "lobby_handlers",
	"lobby/publisher":      "lobby_handlers",
	"web/lobby/rest":       "lobby_handlers",
	"web/lobby/ws":         "lobby_handlers",

	// Web Layer — Middleware
	"web/middleware": "middleware",
	"web/mux":        "middleware",
	"web/nbio":       "middleware",
	"web/otel":       "middleware",

	// Web Layer — REST Utils
	"web/rest":        "rest_utils",
	"web/rest/health": "rest_utils",
	"web/rest/route":  "rest_utils",
	"web/rest/utils":  "rest_utils",

	// Web Layer — WebSocket
	"web/ws":         "websocket",
	"web/ws/message": "websocket",

	// Logic Layer — Move Pipeline
	"logic/game/move/orchestration": "move_pipeline",
	"logic/game/move/attack":        "move_pipeline",
	"logic/game/move/attack/dice":   "move_pipeline",
	"logic/game/move/deploy":        "move_pipeline",
	"logic/game/move/conquer":       "move_pipeline",
	"logic/game/move/reinforce":     "move_pipeline",
	"logic/game/move/cards":         "move_pipeline",
	"logic/game/move/validation":    "move_pipeline",

	// Logic Layer — Game Services
	"logic/game/board":             "game_services",
	"logic/game/phase":             "game_services",
	"logic/game/player":            "game_services",
	"logic/game/region":            "game_services",
	"logic/game/region/assignment": "game_services",
	"logic/game/card":              "game_services",
	"logic/game/mission":           "game_services",
	"logic/game/mission/checker":   "game_services",
	"logic/game/snapshot":          "game_services",
	"logic/game/state":             "game_services",
	"logic/game/creation":          "game_services",
	"logic/game/advancement":       "game_services",
	"logic/game/timing":            "game_services",
	"logic/game/headlines":         "game_services",

	// Logic Layer — Lobby
	"lobby/logic/creation":   "lobby_logic",
	"lobby/logic/management": "lobby_logic",
	"lobby/logic/start":      "lobby_logic",
	"lobby/logic/state":      "lobby_logic",

	// Shared — Domain Errors
	"logic/errors": "domain_errors",

	// Events — Event Bus
	"events":        "event_bus",
	"events/logger": "event_bus",

	// Events — Game Events
	"events/game": "game_events",

	// Events — Lobby Events
	"lobby/events": "lobby_events",

	// Data Layer — Game Data
	"data/game/db": "game_data",

	// Data Layer — Lobby Data
	"lobby/data/db": "lobby_data",

	// Data Layer — Database
	"data/db":        "database",
	"data/pool":      "database",
	"data/migration": "database",

	// Infrastructure — Observability
	"config":  "observability",
	"metrics": "observability",
	"tracing": "observability",
	"slog":    "observability",

	// Infrastructure — Context
	"ctx": "context",

	// Infrastructure — Utilities
	"rand":               "utilities",
	"upgradablerw_mutex": "utilities",

	// API DTOs
	"api/game":                "api_dtos",
	"api/game/messaging":      "api_dtos",
	"api/game/rest/request":   "api_dtos",
	"api/game/rest/response":  "api_dtos",
	"lobby/api/messaging":     "api_dtos",
	"lobby/api/rest/request":  "api_dtos",
	"lobby/api/rest/response": "api_dtos",

	// Testing
	"testing/invariant": "testing",
	"testonly":          "testing",
}

// subsystemDefs defines the display properties for each subsystem group.
//
//nolint:gochecknoglobals // package-level lookup table for subsystem definitions
var subsystemDefs = map[string]subsystem{
	// Web Layer
	"game_handlers":  {ID: "game_handlers", Label: "Game Handlers", Layer: "Web"},
	"lobby_handlers": {ID: "lobby_handlers", Label: "Lobby Handlers", Layer: "Web"},
	"middleware":     {ID: "middleware", Label: "Middleware", Layer: "Web"},
	"rest_utils":     {ID: "rest_utils", Label: "REST Utils", Layer: "Web"},
	"websocket":      {ID: "websocket", Label: "WebSocket", Layer: "Web"},

	// Logic Layer
	"move_pipeline": {ID: "move_pipeline", Label: "Move Pipeline", Layer: "Logic"},
	"game_services": {ID: "game_services", Label: "Game Services", Layer: "Logic"},
	"lobby_logic":   {ID: "lobby_logic", Label: "Lobby", Layer: "Logic"},

	// Shared
	"domain_errors": {ID: "domain_errors", Label: "Domain Errors", Layer: "Shared"},

	// Events
	"event_bus":    {ID: "event_bus", Label: "Event Bus", Layer: "Events"},
	"game_events":  {ID: "game_events", Label: "Game Events", Layer: "Events-domain"},
	"lobby_events": {ID: "lobby_events", Label: "Lobby Events", Layer: "Events-domain"},

	// Data Layer
	"game_data":  {ID: "game_data", Label: "Game Data", Layer: "Data"},
	"lobby_data": {ID: "lobby_data", Label: "Lobby Data", Layer: "Data"},
	"database":   {ID: "database", Label: "Database", Layer: "Data"},

	// Infrastructure
	"observability": {ID: "observability", Label: "Observability", Layer: "Infrastructure"},
	"context":       {ID: "context", Label: "Context", Layer: "Ctx"},
	"utilities":     {ID: "utilities", Label: "Utilities", Layer: "Infrastructure"},

	// API
	"api_dtos": {ID: "api_dtos", Label: "API DTOs", Layer: "API"},

	// Testing
	"testing": {ID: "testing", Label: "Testing", Layer: "Test"},
}

// explicitLayerMap maps known package suffixes to their layer.
// Matches arch_test.go's expectedLayer map exactly.
//
//nolint:gochecknoglobals // package-level lookup table for layer classification
var explicitLayerMap = map[string]string{
	"api/game":                "API",
	"api/game/messaging":      "API",
	"api/game/rest/request":   "API",
	"api/game/rest/response":  "API",
	"lobby/api/messaging":     "API",
	"lobby/api/rest/request":  "API",
	"lobby/api/rest/response": "API",

	"config":             "Infrastructure",
	"metrics":            "Infrastructure",
	"rand":               "Infrastructure",
	"slog":               "Infrastructure",
	"tracing":            "Infrastructure",
	"upgradablerw_mutex": "Infrastructure",

	"ctx": "Ctx",

	"data/db":        "Data",
	"data/game/db":   "Data",
	"lobby/data/db":  "Data",
	"data/migration": "Data",
	"data/pool":      "Data",

	"events":        "Events",
	"events/logger": "Events",

	"events/game":  "Events-domain",
	"lobby/events": "Events-domain",

	"game/publisher":           "Web",
	"game/publisher/converter": "Web",
	"lobby/publisher":          "Web",

	"logic/errors": "Shared",

	"testing/invariant": "Test",
	"testonly":          "Test",
}

// layerFromPrefix derives the layer for a package suffix.
// Mirrors arch_test.go's layerFromPrefix exactly.
func layerFromPrefix(suffix string) string {
	if layer, ok := explicitLayerMap[suffix]; ok {
		return layer
	}

	return layerFromPrefixFallback(suffix)
}

// layerFromPrefixFallback uses prefix matching for packages not in the explicit map.
//
//nolint:cyclop // switch mirrors arch_test.go's layerFromPrefix exactly
func layerFromPrefixFallback(suffix string) string {
	switch {
	case strings.HasPrefix(suffix, "api/"):
		return "API"
	case strings.HasPrefix(suffix, "lobby/api/"):
		return "API"
	case strings.HasPrefix(suffix, "lobby/data/"):
		return "Data"
	case strings.HasPrefix(suffix, "lobby/events"):
		return "Events-domain"
	case strings.HasPrefix(suffix, "lobby/logic/"):
		return "Logic"
	case strings.HasPrefix(suffix, "lobby/publisher"):
		return "Web"
	case strings.HasPrefix(suffix, "data/"):
		return "Data"
	case strings.HasPrefix(suffix, "events/game") || strings.HasPrefix(suffix, "events/lobby"):
		return "Events-domain"
	case strings.HasPrefix(suffix, "events/"):
		return "Events"
	case strings.HasPrefix(suffix, "logic/"):
		return "Logic"
	case strings.HasPrefix(suffix, "web/"):
		return "Web"
	case strings.HasPrefix(suffix, "testing/") || suffix == "testonly":
		return "Test"
	default:
		return ""
	}
}

// subsystemFromSuffix returns the subsystem ID for a package suffix.
// Falls back to prefix-based matching if no explicit mapping exists.
func subsystemFromSuffix(suffix string) string {
	if sub, ok := subsystemMap[suffix]; ok {
		return sub
	}

	return subsystemFromSuffixFallback(suffix)
}

// subsystemFromSuffixFallback uses prefix matching for packages not in the explicit map.
//
//nolint:cyclop,funlen // switch covers all subsystem prefix patterns
func subsystemFromSuffixFallback(suffix string) string {
	switch {
	case strings.HasPrefix(suffix, "api/"):
		return "api_dtos"
	case strings.HasPrefix(suffix, "lobby/api/"):
		return "api_dtos"
	case strings.HasPrefix(suffix, "lobby/publisher"):
		return "lobby_handlers"
	case strings.HasPrefix(suffix, "web/game/"):
		return "game_handlers"
	case strings.HasPrefix(suffix, "web/lobby/"):
		return "lobby_handlers"
	case strings.HasPrefix(suffix, "web/ws"):
		return "websocket"
	case strings.HasPrefix(suffix, "web/rest"):
		return "rest_utils"
	case strings.HasPrefix(suffix, "web/"):
		return "middleware"
	case strings.HasPrefix(suffix, "logic/game/move/"):
		return "move_pipeline"
	case strings.HasPrefix(suffix, "logic/game/"):
		return "game_services"
	case strings.HasPrefix(suffix, "logic/lobby/"):
		return "lobby_logic"
	case strings.HasPrefix(suffix, "lobby/logic/"):
		return "lobby_logic"
	case strings.HasPrefix(suffix, "logic/"):
		return "domain_errors"
	case strings.HasPrefix(suffix, "events/game"):
		return "game_events"
	case strings.HasPrefix(suffix, "events/lobby"):
		return "lobby_events"
	case strings.HasPrefix(suffix, "lobby/events"):
		return "lobby_events"
	case strings.HasPrefix(suffix, "events/"):
		return "event_bus"
	case strings.HasPrefix(suffix, "data/game/"):
		return "game_data"
	case strings.HasPrefix(suffix, "data/lobby/"):
		return "lobby_data"
	case strings.HasPrefix(suffix, "lobby/data/"):
		return "lobby_data"
	case strings.HasPrefix(suffix, "data/"):
		return "database"
	case strings.HasPrefix(suffix, "testing/") || suffix == "testonly":
		return "testing"
	default:
		return ""
	}
}

// packageSuffix returns the import path after the module internal prefix.
func packageSuffix(importPath string) string {
	return strings.TrimPrefix(importPath, modulePrefix)
}

// isExcluded returns true for packages that should not appear in the diagram.
func isExcluded(suffix string) bool {
	if wiringRoots[suffix] {
		return true
	}

	if strings.Contains(suffix, "/sqlc") || strings.Contains(suffix, "/mocks") {
		return true
	}

	return false
}

// nodeID converts a package suffix to a D2-safe node identifier.
// Slashes become dots, underscores stay.
func nodeID(suffix string) string {
	return strings.ReplaceAll(suffix, "/", ".")
}

// shortName returns the last path segment of a suffix for compact display.
func shortName(suffix string) string {
	parts := strings.Split(suffix, "/")

	return parts[len(parts)-1]
}

// internalImports returns only imports under our module's internal/.
func internalImports(pkg goPackage) []string {
	var result []string

	for _, imp := range pkg.Imports {
		if strings.HasPrefix(imp, modulePrefix) {
			result = append(result, imp)
		}
	}

	return result
}

// edge represents a dependency between two entities (layers or subsystems).
type edge struct {
	From string
	To   string
}

// loadPackages runs `go list -json` and returns parsed packages.
func loadPackages(ctx context.Context, pattern string) ([]goPackage, error) {
	cmd := exec.CommandContext(ctx, "go", "list", "-json", pattern)

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("go list failed: %w", err)
	}

	var pkgs []goPackage

	dec := json.NewDecoder(strings.NewReader(string(out)))
	for dec.More() {
		var pkg goPackage
		if err := dec.Decode(&pkg); err != nil {
			return nil, fmt.Errorf("failed to decode go list output: %w", err)
		}

		pkgs = append(pkgs, pkg)
	}

	return pkgs, nil
}

// generateD2 produces the D2 diagram source string with subsystem-grouped nodes.
func generateD2(activeSubsystems []subsystem, crossLayerEdges []edge) string {
	var buf strings.Builder

	buf.WriteString("# Architecture diagram — auto-generated by cmd/archdiagram.\n")
	buf.WriteString("# Do not edit manually; run 'make diagrams' to regenerate.\n\n")
	buf.WriteString("title: go-risk-it Architecture {\n")
	buf.WriteString("  near: top-center\n")
	buf.WriteString("  style.font-size: 28\n")
	buf.WriteString("  style.underline: true\n")
	buf.WriteString("}\n\n")
	buf.WriteString("direction: right\n\n")

	writeLayerContainers(&buf, activeSubsystems)
	writeCrossLayerEdges(&buf, crossLayerEdges)

	return buf.String()
}

// writeLayerContainers emits D2 container blocks with subsystem nodes.
func writeLayerContainers(buf *strings.Builder, activeSubsystems []subsystem) {
	// Group subsystems by layer
	byLayer := make(map[string][]subsystem)
	for _, sub := range activeSubsystems {
		byLayer[sub.Layer] = append(byLayer[sub.Layer], sub)
	}

	// Sort layers by order
	sortedLayers := make([]string, 0, len(byLayer))
	for l := range byLayer {
		sortedLayers = append(sortedLayers, l)
	}

	sort.Slice(sortedLayers, func(i, j int) bool {
		return layers[sortedLayers[i]].Order < layers[sortedLayers[j]].Order
	})

	for _, layerKey := range sortedLayers {
		info := layers[layerKey]
		layerSubs := byLayer[layerKey]

		sort.Slice(layerSubs, func(i, j int) bool {
			return layerSubs[i].ID < layerSubs[j].ID
		})

		containerID := strings.ReplaceAll(strings.ToLower(layerKey), "-", "_")

		fmt.Fprintf(buf, "%s: %s {\n", containerID, info.Name)
		fmt.Fprintf(buf, "  style.fill: %q\n", info.Color)

		for _, sub := range layerSubs {
			fmt.Fprintf(buf, "  %s: %q\n", sub.ID, sub.Label)
		}

		buf.WriteString("}\n\n")
	}
}

// writeCrossLayerEdges emits sorted D2 edge declarations at the layer level.
func writeCrossLayerEdges(buf *strings.Builder, edges []edge) {
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].From != edges[j].From {
			return edges[i].From < edges[j].From
		}

		return edges[i].To < edges[j].To
	})

	if len(edges) > 0 {
		buf.WriteString("# Cross-layer dependencies (layer-to-layer)\n")
	}

	for _, e := range edges {
		fromContainer := layerContainerID(e.From)
		toContainer := layerContainerID(e.To)

		fmt.Fprintf(buf, "%s -> %s\n", fromContainer, toContainer)
	}
}

// containerIDForSuffix returns the D2 container ID for a package suffix's layer.
func containerIDForSuffix(suffix string) string {
	layer := layerFromPrefix(suffix)

	return strings.ReplaceAll(strings.ToLower(layer), "-", "_")
}

// layerContainerID returns the D2 container ID for a layer name.
func layerContainerID(layerName string) string {
	return strings.ReplaceAll(strings.ToLower(layerName), "-", "_")
}

// classifySubsystems determines which subsystems are active based on discovered packages,
// and returns the suffix-to-layer lookup for edge classification.
func classifySubsystems(pkgs []goPackage) ([]subsystem, map[string]string) {
	suffixToLayer := make(map[string]string)
	activeSubIDs := make(map[string]bool)

	for _, pkg := range pkgs {
		suffix := packageSuffix(pkg.ImportPath)
		if isExcluded(suffix) {
			continue
		}

		layer := layerFromPrefix(suffix)
		if layer == "" {
			log.Printf("WARNING: no layer mapping for %s, skipping", suffix)

			continue
		}

		suffixToLayer[suffix] = layer

		subID := subsystemFromSuffix(suffix)
		if subID == "" {
			log.Printf("WARNING: no subsystem mapping for %s, skipping", suffix)

			continue
		}

		activeSubIDs[subID] = true
	}

	// Collect active subsystem definitions, preserving definition order
	var active []subsystem
	for id := range activeSubIDs {
		if def, ok := subsystemDefs[id]; ok {
			active = append(active, def)
		} else {
			log.Printf("WARNING: subsystem %q referenced but not defined", id)
		}
	}

	return active, suffixToLayer
}

// collectCrossLayerEdges finds all imports that cross layer boundaries,
// deduplicated at the layer level.
func collectCrossLayerEdges(pkgs []goPackage, suffixToLayer map[string]string) []edge {
	seen := make(map[edge]bool)
	var result []edge

	for _, pkg := range pkgs {
		fromSuffix := packageSuffix(pkg.ImportPath)
		fromLayer, ok := suffixToLayer[fromSuffix]

		if !ok {
			continue
		}

		for _, imp := range internalImports(pkg) {
			toSuffix := packageSuffix(imp)
			toLayer, ok := suffixToLayer[toSuffix]

			if !ok || fromLayer == toLayer {
				continue
			}

			layerEdge := edge{
				From: fromLayer,
				To:   toLayer,
			}
			if !seen[layerEdge] {
				seen[layerEdge] = true
				result = append(result, layerEdge)
			}
		}
	}

	return result
}

func run() error {
	outputDir := flag.String("output", "docs", "output directory for generated files")
	d2Bin := flag.String("d2", "d2", "path to d2 binary")
	flag.Parse()

	ctx := context.Background()

	log.Println("Loading packages...")

	pkgs, err := loadPackages(ctx, "./internal/...")
	if err != nil {
		return fmt.Errorf("loading packages: %w", err)
	}

	log.Printf("Loaded %d packages", len(pkgs))

	activeSubsystems, suffixToLayer := classifySubsystems(pkgs)
	crossLayerEdges := collectCrossLayerEdges(pkgs, suffixToLayer)

	log.Printf("Diagram: %d subsystem nodes, %d cross-layer edges",
		len(activeSubsystems), len(crossLayerEdges))

	d2Source := generateD2(activeSubsystems, crossLayerEdges)

	if err := os.MkdirAll(*outputDir, 0o755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	d2Path := filepath.Join(*outputDir, "architecture-diagram.d2")
	svgPath := filepath.Join(*outputDir, "architecture-diagram.svg")

	if err := os.WriteFile(d2Path, []byte(d2Source), 0o600); err != nil {
		return fmt.Errorf("writing D2 source: %w", err)
	}

	log.Printf("Wrote D2 source: %s", d2Path)

	return renderSVG(ctx, *d2Bin, d2Path, svgPath)
}

// renderSVG shells out to the d2 CLI to render the SVG.
func renderSVG(ctx context.Context, d2Bin, d2Path, svgPath string) error {
	start := time.Now()

	cmd := exec.CommandContext(
		ctx,
		d2Bin,
		"--layout",
		"dagre",
		"--theme",
		"0",
		d2Path,
		svgPath,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("d2 rendering failed: %w", err)
	}

	elapsed := time.Since(start)

	log.Printf("Rendered SVG in %s: %s", elapsed.Round(time.Millisecond), svgPath)

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("ERROR: %v", err)
	}
}
