// Package main generates architecture documentation from the go-risk-it internal packages.
//
// It reads all internal packages via `go list -json`, classifies them by parsing
// their doc.go files using the docparser package, groups them into subsystem
// nodes within architectural layer containers via the model package, and produces:
//
//   - A D2 diagram with cross-layer dependency edges
//   - Mermaid component and package architecture diagrams injected into docs/
//   - A project structure tree injected into README.md
//   - Package classification tables injected into docs/doc-go-spec.md
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

	"github.com/go-risk-it/go-risk-it/cmd/archdiagram/docparser"
	"github.com/go-risk-it/go-risk-it/cmd/archdiagram/inject"
	"github.com/go-risk-it/go-risk-it/cmd/archdiagram/model"
	"github.com/go-risk-it/go-risk-it/cmd/archdiagram/render"
)

// layers defines the architectural layers and their visual properties.
// Must stay in sync with the doc.go taxonomy and completeness_test.go.
//
//nolint:gochecknoglobals // package-level layer definition for the D2 renderer
var layers = map[string]*model.LayerInfo{
	"API":           {Name: "API", Color: "#E8EAF6", Order: 0},
	"Kernel":        {Name: "Kernel", Color: "#F3E5F5", Order: 1},
	"Data":          {Name: "Data", Color: "#FFF3E0", Order: 2},
	"Events-domain": {Name: "Events-domain", Color: "#FCE4EC", Order: 3},
	"Game-domain":   {Name: "Game-domain", Color: "#E0F7FA", Order: 4},
	"Game-support":  {Name: "Game-support", Color: "#FFF9C4", Order: 5},
	"Lobby-domain":  {Name: "Lobby-domain", Color: "#E0F7FA", Order: 6},
	"Logic":         {Name: "Logic", Color: "#E8F5E9", Order: 7},
	"Web":           {Name: "Web", Color: "#E3F2FD", Order: 8},
	"Test":          {Name: "Test", Color: "#F5F5F5", Order: 9},
}

// visualContainer defines how a doc.go layer renders in the D2 diagram.
// The diagram uses fewer, larger containers than the 10-layer enforcement taxonomy.
type visualContainer struct {
	Name  string // display name (with emoji prefix)
	Color string // fill color (hex)
	Order int    // sort order for container placement
}

// layerToVisual maps doc.go layer names to visual container names.
// Multiple layers can consolidate into one visual container.
// Empty string means excluded from diagram.
//
//nolint:gochecknoglobals // package-level visual container mapping
var layerToVisual = map[string]string{
	"API":           "API",
	"Kernel":        "Kernel",
	"Data":          "Data",
	"Events-domain": "Events",
	"Game-domain":   "Logic",
	"Game-support":  "Logic",
	"Lobby-domain":  "Logic",
	"Logic":         "Logic",
	"Web":           "Web",
	"Test":          "",
}

// visualContainers defines the 6 visual containers for the D2 diagram.
//
//nolint:gochecknoglobals // package-level visual container definition
var visualContainers = map[string]*visualContainer{
	"API":    {Name: "📋 API", Color: "#E8EAF6", Order: 0},
	"Web":    {Name: "🌐 Web", Color: "#E3F2FD", Order: 1},
	"Events": {Name: "📡 Events", Color: "#FCE4EC", Order: 2},
	"Logic":  {Name: "⚙️ Logic", Color: "#E8F5E9", Order: 3},
	"Data":   {Name: "💾 Data", Color: "#FFF3E0", Order: 4},
	"Kernel": {Name: "🔧 Kernel", Color: "#F3E5F5", Order: 5},
}

// containerStrokes maps visual container names to their darker stroke colors.
//
//nolint:gochecknoglobals // package-level stroke color definition
var containerStrokes = map[string]string{
	"API":    "#283593",
	"Web":    "#1565C0",
	"Events": "#AD1457",
	"Logic":  "#2E7D32",
	"Data":   "#E65100",
	"Kernel": "#6A1B9A",
}

// subContainerFills maps sub-container module types to their fill colors.
//
//nolint:gochecknoglobals // package-level sub-container color definition
var subContainerFills = map[string]string{
	"game":   "#BBDEFB",
	"lobby":  "#C8E6C9",
	"shared": "#F5F5F5",
}

// loadPackages runs `go list -json` and returns parsed packages.
func loadPackages(ctx context.Context, pattern string) ([]model.GoPackage, error) {
	cmd := exec.CommandContext(ctx, "go", "list", "-json", pattern)

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("go list failed: %w", err)
	}

	var pkgs []model.GoPackage

	dec := json.NewDecoder(strings.NewReader(string(out)))
	for dec.More() {
		var pkg model.GoPackage
		if err := dec.Decode(&pkg); err != nil {
			return nil, fmt.Errorf("failed to decode go list output: %w", err)
		}

		pkgs = append(pkgs, pkg)
	}

	return pkgs, nil
}

// buildSuffixToDirMap builds a map from package suffix to its filesystem Dir path.
// Only includes packages under the module's internal/ prefix.
func buildSuffixToDirMap(pkgs []model.GoPackage) map[string]string {
	result := make(map[string]string, len(pkgs))

	for _, pkg := range pkgs {
		if !strings.HasPrefix(pkg.ImportPath, model.ModulePrefix) {
			continue
		}

		suffix := model.PackageSuffix(pkg.ImportPath)
		result[suffix] = pkg.Dir
	}

	return result
}

// makeClassifier returns a model.Classifier that reads doc.go files from disk
// using the docparser package. It resolves package suffixes to filesystem
// directories using the provided suffix-to-dir map.
func makeClassifier(suffixToDir map[string]string) model.Classifier {
	return func(suffix string) (string, string) {
		dir, ok := suffixToDir[suffix]
		if !ok {
			return "", ""
		}

		layer, summary, err := docparser.ParseLayerAndSummary(dir)
		if err != nil {
			log.Printf("WARNING: failed to parse doc.go for %s: %v", suffix, err)

			return "", ""
		}

		return layer, summary
	}
}

// setSubsystemLayers assigns each subsystem's Layer field by majority vote
// of its member packages' layers. If a subsystem has packages from multiple
// layers, the most common layer wins. Ties are broken alphabetically for
// deterministic output.
func setSubsystemLayers(m *model.ArchModel) {
	for _, sub := range m.Subsystems {
		sub.Layer = majorityLayer(m, sub.Packages)
	}
}

// majorityLayer returns the most common layer among the given package suffixes.
// Ties are broken alphabetically for deterministic output.
func majorityLayer(m *model.ArchModel, suffixes []string) string {
	counts := make(map[string]int)

	for _, suffix := range suffixes {
		pkg, ok := m.Packages[suffix]
		if !ok {
			continue
		}

		counts[pkg.Layer]++
	}

	var bestLayer string

	bestCount := 0

	for layer, count := range counts {
		if count > bestCount || (count == bestCount && layer < bestLayer) {
			bestLayer = layer
			bestCount = count
		}
	}

	return bestLayer
}

// detectModule determines whether a subsystem belongs to the game module,
// lobby module, or is shared infrastructure.
func detectModule(sub *model.SubsystemInfo) string {
	allGame, allLobby := true, true

	for _, suffix := range sub.Packages {
		if !strings.HasPrefix(suffix, "game/") {
			allGame = false
		}

		if !strings.HasPrefix(suffix, "lobby/") {
			allLobby = false
		}
	}

	if allGame {
		return "game"
	}

	if allLobby {
		return "lobby"
	}

	return ""
}

// containerID returns the D2 container ID for a container name.
// Converts to lowercase and replaces hyphens with underscores.
func containerID(name string) string {
	return strings.ReplaceAll(strings.ToLower(name), "-", "_")
}

// generateD2 produces the D2 diagram source string from the architecture model.
func generateD2(archModel *model.ArchModel) string {
	var buf strings.Builder

	buf.WriteString("# Architecture diagram — auto-generated by cmd/archdiagram.\n")
	buf.WriteString("# Do not edit manually; run 'make diagrams' to regenerate.\n\n")
	buf.WriteString("title: go-risk-it Architecture {\n")
	buf.WriteString("  near: top-center\n")
	buf.WriteString("  style.font-size: 28\n")
	buf.WriteString("  style.underline: true\n")
	buf.WriteString("}\n\n")
	buf.WriteString("direction: down\n\n")

	writeLayerContainers(&buf, archModel)
	writeCrossLayerEdges(&buf, archModel)

	return buf.String()
}

// subsystemPlacement holds a subsystem and its detected module for grouping.
type subsystemPlacement struct {
	Subsystem *model.SubsystemInfo
	Module    string // "game", "lobby", or "" (shared)
}

// writeLayerContainers emits D2 container blocks with subsystem nodes,
// consolidating the 10 enforcement layers into 6 visual containers.
// Containers with both game and lobby subsystems get nested sub-containers.
func writeLayerContainers(buf *strings.Builder, archModel *model.ArchModel) {
	// Group subsystems by visual container.
	byVisual := make(map[string][]subsystemPlacement)

	for _, sub := range archModel.Subsystems {
		visualName := layerToVisual[sub.Layer]
		if visualName == "" {
			continue
		}

		byVisual[visualName] = append(byVisual[visualName], subsystemPlacement{
			Subsystem: sub,
			Module:    detectModule(sub),
		})
	}

	// Sort visual containers by order.
	sortedContainers := make([]string, 0, len(byVisual))
	for name := range byVisual {
		sortedContainers = append(sortedContainers, name)
	}

	sort.Slice(sortedContainers, func(i, j int) bool {
		containerI, okI := visualContainers[sortedContainers[i]]
		containerJ, okJ := visualContainers[sortedContainers[j]]

		if !okI || !okJ {
			return sortedContainers[i] < sortedContainers[j]
		}

		return containerI.Order < containerJ.Order
	})

	for _, visualName := range sortedContainers {
		info, ok := visualContainers[visualName]
		if !ok {
			continue
		}

		placements := byVisual[visualName]
		writeVisualContainer(buf, info, visualName, placements)
	}
}

// writeVisualContainer emits a single D2 container block. If the container has
// subsystems from multiple modules (game + lobby), it creates nested
// sub-containers. Otherwise it emits flat subsystem nodes.
func writeVisualContainer(
	buf *strings.Builder,
	info *visualContainer,
	visualName string,
	placements []subsystemPlacement,
) {
	visID := containerID(visualName)

	fmt.Fprintf(buf, "%s: %q {\n", visID, info.Name)
	fmt.Fprintf(buf, "  style.fill: %q\n", info.Color)

	if stroke, ok := containerStrokes[visualName]; ok {
		fmt.Fprintf(buf, "  style.stroke: %q\n", stroke)
	}

	buf.WriteString("  style.border-radius: 12\n")

	hasGame, hasLobby := detectModulePresence(placements)
	needsNesting := hasGame && hasLobby

	if needsNesting {
		writeNestedSubContainers(buf, placements)
	} else {
		writeFlatSubsystems(buf, placements)
	}

	buf.WriteString("}\n\n")
}

// detectModulePresence checks whether placements contain game and/or lobby modules.
func detectModulePresence(placements []subsystemPlacement) (bool, bool) {
	var gamePresent, lobbyPresent bool

	for _, placement := range placements {
		switch placement.Module {
		case "game":
			gamePresent = true
		case "lobby":
			lobbyPresent = true
		}
	}

	return gamePresent, lobbyPresent
}

// writeNestedSubContainers emits game, lobby, and optionally shared
// sub-containers within a visual container.
func writeNestedSubContainers(buf *strings.Builder, placements []subsystemPlacement) {
	grouped := map[string][]subsystemPlacement{
		"game":   {},
		"lobby":  {},
		"shared": {},
	}

	for _, placement := range placements {
		key := placement.Module
		if key == "" {
			key = "shared"
		}

		grouped[key] = append(grouped[key], placement)
	}

	// Emit sub-containers in fixed order: Game, Lobby, Shared.
	subContainerOrder := []struct {
		key   string
		label string
	}{
		{"game", "Game"},
		{"lobby", "Lobby"},
		{"shared", "Shared"},
	}

	for _, subContainer := range subContainerOrder {
		subs := grouped[subContainer.key]
		if len(subs) == 0 {
			continue
		}

		sort.Slice(subs, func(i, j int) bool {
			return subs[i].Subsystem.ID < subs[j].Subsystem.ID
		})

		fmt.Fprintf(buf, "  %s: %q {\n", subContainer.key, subContainer.label)

		if fill, ok := subContainerFills[subContainer.key]; ok {
			fmt.Fprintf(buf, "    style.fill: %q\n", fill)
		}

		buf.WriteString("    style.border-radius: 8\n")

		for _, placement := range subs {
			fmt.Fprintf(buf, "    %s: %q {\n", placement.Subsystem.ID, placement.Subsystem.Label)
			buf.WriteString("      style.border-radius: 6\n")
			buf.WriteString("    }\n")
		}

		buf.WriteString("  }\n")
	}
}

// writeFlatSubsystems emits subsystem nodes directly within a container (no nesting).
func writeFlatSubsystems(buf *strings.Builder, placements []subsystemPlacement) {
	sort.Slice(placements, func(i, j int) bool {
		return placements[i].Subsystem.ID < placements[j].Subsystem.ID
	})

	for _, placement := range placements {
		fmt.Fprintf(buf, "  %s: %q {\n", placement.Subsystem.ID, placement.Subsystem.Label)
		buf.WriteString("    style.border-radius: 6\n")
		buf.WriteString("  }\n")
	}
}

// writeCrossLayerEdges emits sorted, deduplicated D2 edge declarations
// at the visual container level. Edges where either end maps to an excluded
// container (Test) are dropped.
func writeCrossLayerEdges(buf *strings.Builder, archModel *model.ArchModel) {
	// Map edges through layerToVisual and deduplicate.
	seen := make(map[model.Edge]bool)

	var visualEdges []model.Edge

	for _, edge := range archModel.Edges {
		fromVisual := layerToVisual[edge.From]
		toVisual := layerToVisual[edge.To]

		if fromVisual == "" || toVisual == "" {
			continue
		}

		if fromVisual == toVisual {
			continue
		}

		mapped := model.Edge{From: fromVisual, To: toVisual}
		if !seen[mapped] {
			seen[mapped] = true
			visualEdges = append(visualEdges, mapped)
		}
	}

	sort.Slice(visualEdges, func(i, j int) bool {
		if visualEdges[i].From != visualEdges[j].From {
			return visualEdges[i].From < visualEdges[j].From
		}

		return visualEdges[i].To < visualEdges[j].To
	})

	if len(visualEdges) > 0 {
		buf.WriteString("# Cross-layer dependencies (container-to-container)\n")
	}

	for _, edge := range visualEdges {
		fromContainer := containerID(edge.From)
		toContainer := containerID(edge.To)

		fmt.Fprintf(buf, "%s -> %s\n", fromContainer, toContainer)
	}
}

// renderSVG shells out to the d2 CLI to render the SVG.
func renderSVG(ctx context.Context, d2Bin, d2Path, svgPath string) error {
	start := time.Now()

	cmd := exec.CommandContext(
		ctx,
		d2Bin,
		"--layout",
		"elk",
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

// generateD2Pipeline writes the D2 source and renders the SVG.
func generateD2Pipeline(
	ctx context.Context,
	archModel *model.ArchModel,
	outputDir, d2Bin string,
) error {
	d2Source := generateD2(archModel)

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	d2Path := filepath.Join(outputDir, "architecture-diagram.d2")
	svgPath := filepath.Join(outputDir, "architecture-diagram.svg")

	if err := os.WriteFile(d2Path, []byte(d2Source), 0o600); err != nil {
		return fmt.Errorf("writing D2 source: %w", err)
	}

	log.Printf("Wrote D2 source: %s", d2Path)

	return renderSVG(ctx, d2Bin, d2Path, svgPath)
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

	suffixToDir := buildSuffixToDirMap(pkgs)
	classify := makeClassifier(suffixToDir)

	archModel := model.BuildModel(pkgs, classify, layers)
	model.GroupPackages(archModel)
	setSubsystemLayers(archModel)

	log.Printf("Diagram: %d subsystem nodes, %d cross-layer edges",
		len(archModel.Subsystems), len(archModel.Edges))

	// --- D2 diagram pipeline (existing) ---
	if err := generateD2Pipeline(ctx, archModel, *outputDir, *d2Bin); err != nil {
		return err
	}

	// --- Count extraction ---
	ruleCount, err := render.CountArchRules(".")
	if err != nil {
		return fmt.Errorf("counting arch rules: %w", err)
	}

	log.Printf("Counted %d architecture rules", ruleCount)

	invariantCount, err := render.CountInvariants(".")
	if err != nil {
		return fmt.Errorf("counting invariants: %w", err)
	}

	log.Printf("Counted %d game invariants", invariantCount)

	// --- Taxonomy consistency check ---
	if err := render.CheckTaxonomyConsistency(".", archModel.Layers); err != nil {
		return fmt.Errorf("taxonomy consistency check failed: %w", err)
	}

	log.Println("Taxonomy consistency check passed")

	// --- Markdown documentation injection pipeline ---
	stats := render.RenderStats(ruleCount, invariantCount)

	if err := injectDocs(*outputDir, archModel, pkgs, stats); err != nil {
		return fmt.Errorf("injecting docs: %w", err)
	}

	return nil
}

// injectDocs renders all markdown outputs and injects them into documentation files.
// The outputDir is used to resolve doc file paths (docs/ relative to project root).
// The project root is the parent of outputDir (which defaults to "docs").
func injectDocs(
	outputDir string,
	archModel *model.ArchModel,
	pkgs []model.GoPackage,
	stats string,
) error {
	// Resolve project root from output dir. outputDir is typically "docs",
	// so project root is its parent. For absolute paths, use parent directly.
	projectRoot := filepath.Dir(outputDir)
	if !filepath.IsAbs(outputDir) {
		// outputDir is relative (e.g., "docs"), so project root is "."
		projectRoot = "."
	}

	// Render all outputs.
	componentArch := render.RenderComponentArch(archModel)
	packageArch := render.RenderPackageArch(archModel)
	projectTree := render.RenderProjectTree(archModel, pkgs)
	packageTables := render.RenderPackageTables(archModel, pkgs)

	// Inject into documentation files.
	injections := []struct {
		file    string
		marker  string
		content string
	}{
		{
			file:    filepath.Join(projectRoot, "docs", "architecture-components.md"),
			marker:  "MERMAID_COMPONENT_ARCH",
			content: componentArch,
		},
		{
			file:    filepath.Join(projectRoot, "docs", "architecture.md"),
			marker:  "MERMAID_PACKAGE_ARCH",
			content: packageArch,
		},
		{
			file:    filepath.Join(projectRoot, "README.md"),
			marker:  "PROJECT_TREE",
			content: projectTree,
		},
		{
			file:    filepath.Join(projectRoot, "README.md"),
			marker:  "STATS",
			content: stats,
		},
		{
			file:    filepath.Join(projectRoot, "docs", "doc-go-spec.md"),
			marker:  "PACKAGE_TABLES",
			content: packageTables,
		},
	}

	for _, inj := range injections {
		if err := inject.InjectFile(inj.file, inj.marker, inj.content); err != nil {
			return fmt.Errorf("injecting %s into %s: %w", inj.marker, inj.file, err)
		}

		log.Printf("Injected %s into %s", inj.marker, inj.file)
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("ERROR: %v", err)
	}
}
