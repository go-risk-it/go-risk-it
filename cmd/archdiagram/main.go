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

// layerContainerID returns the D2 container ID for a layer name.
// Converts to lowercase and replaces hyphens with underscores.
func layerContainerID(layerName string) string {
	return strings.ReplaceAll(strings.ToLower(layerName), "-", "_")
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
	buf.WriteString("direction: right\n\n")

	writeLayerContainers(&buf, archModel)
	writeCrossLayerEdges(&buf, archModel)

	return buf.String()
}

// writeLayerContainers emits D2 container blocks with subsystem nodes.
func writeLayerContainers(buf *strings.Builder, archModel *model.ArchModel) {
	// Group subsystems by layer
	byLayer := make(map[string][]*model.SubsystemInfo)
	for _, sub := range archModel.Subsystems {
		byLayer[sub.Layer] = append(byLayer[sub.Layer], sub)
	}

	// Sort layers by order
	sortedLayers := make([]string, 0, len(byLayer))
	for l := range byLayer {
		sortedLayers = append(sortedLayers, l)
	}

	sort.Slice(sortedLayers, func(i, j int) bool {
		layerI, okI := archModel.Layers[sortedLayers[i]]
		layerJ, okJ := archModel.Layers[sortedLayers[j]]

		if !okI || !okJ {
			return sortedLayers[i] < sortedLayers[j]
		}

		return layerI.Order < layerJ.Order
	})

	for _, layerKey := range sortedLayers {
		info, ok := archModel.Layers[layerKey]
		if !ok {
			continue
		}

		layerSubs := byLayer[layerKey]

		sort.Slice(layerSubs, func(i, j int) bool {
			return layerSubs[i].ID < layerSubs[j].ID
		})

		containerID := layerContainerID(layerKey)

		fmt.Fprintf(buf, "%s: %s {\n", containerID, info.Name)
		fmt.Fprintf(buf, "  style.fill: %q\n", info.Color)

		for _, sub := range layerSubs {
			fmt.Fprintf(buf, "  %s: %q\n", sub.ID, sub.Label)
		}

		buf.WriteString("}\n\n")
	}
}

// writeCrossLayerEdges emits sorted D2 edge declarations at the layer level.
func writeCrossLayerEdges(buf *strings.Builder, archModel *model.ArchModel) {
	edges := make([]model.Edge, len(archModel.Edges))
	copy(edges, archModel.Edges)

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
