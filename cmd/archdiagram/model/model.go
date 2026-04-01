// Package model defines the shared architecture model types and the BuildModel
// constructor. It is the contract between the parser (doc.go reader), grouper
// (subsystem classifier), and renderers (D2, Mermaid, tree, tables).
//
// # Layer: Shared
//
// Shared types for the archdiagram tool.
package model

import (
	"strings"
)

// ModulePrefix is the import path prefix for internal packages.
const ModulePrefix = "github.com/go-risk-it/go-risk-it/internal/"

// GoPackage is the subset of `go list -json` fields needed by the model builder.
type GoPackage struct {
	ImportPath string   `json:"ImportPath"` //nolint:tagliatelle // matches go list -json output
	Dir        string   `json:"Dir"`        //nolint:tagliatelle // matches go list -json output
	Imports    []string `json:"Imports"`    //nolint:tagliatelle // matches go list -json output
	GoFiles    []string `json:"GoFiles"`    //nolint:tagliatelle // matches go list -json output
}

// Classifier maps a package suffix to its architectural layer and summary.
// Returns empty layer for packages that cannot be classified.
type Classifier func(suffix string) (layer, summary string)

// PackageInfo represents a parsed Go package with its architectural metadata.
type PackageInfo struct {
	ImportPath   string   // full import path
	Suffix       string   // path after module prefix (e.g., "game/logic/board")
	Layer        string   // architectural layer from doc.go (e.g., "Logic", "Web")
	Summary      string   // first paragraph of doc.go
	GoFiles      []string // non-test Go source files from go list
	InternalDeps []string // suffixes of internal package dependencies
}

// SubsystemInfo groups related packages into a diagram node.
type SubsystemInfo struct {
	ID       string   // D2-safe identifier
	Label    string   // display label
	Layer    string   // architectural layer
	Packages []string // package suffixes in this subsystem
}

// LayerInfo describes a layer's display properties.
type LayerInfo struct {
	Name  string // display name
	Color string // fill color (hex)
	Order int    // sort order
}

// Edge represents a dependency between two architectural layers.
type Edge struct {
	From string // source layer
	To   string // target layer
}

// ArchModel is the shared model consumed by all renderers.
type ArchModel struct {
	Packages   map[string]*PackageInfo   // suffix -> info
	Subsystems map[string]*SubsystemInfo // subsystem ID -> info
	Layers     map[string]*LayerInfo     // layer name -> info
	Edges      []Edge                    // cross-layer dependency edges
}

// wiringRoots are single-file fx.Module aggregation packages excluded from the diagram.
// Matches the set in arch_test.go and main.go.
//
//nolint:gochecknoglobals // package-level lookup table for exclusion
var wiringRoots = map[string]bool{
	"":                        true, // internal root
	"kernel":                  true,
	"game":                    true,
	"game/logic":              true,
	"game/logic/move":         true,
	"game/logic/move/service": true,
	"game/data":               true,
	"lobby":                   true,
	"lobby/logic":             true,
	"lobby/data":              true,
	"web":                     true,
}

// PackageSuffix returns the import path after the module internal prefix.
func PackageSuffix(importPath string) string {
	return strings.TrimPrefix(importPath, ModulePrefix)
}

// IsExcluded returns true for packages that should not appear in the diagram.
// This includes wiring roots, sqlc-generated packages, and mock packages.
func IsExcluded(suffix string) bool {
	if wiringRoots[suffix] {
		return true
	}

	if strings.Contains(suffix, "/sqlc") || strings.Contains(suffix, "/mocks") {
		return true
	}

	// Loadtest packages are a separate domain — excluded from the server architecture diagram.
	if strings.HasPrefix(suffix, "loadtest/") {
		return true
	}

	return false
}

// internalImportSuffixes returns the suffixes of imports under the module's internal/.
func internalImportSuffixes(pkg GoPackage) []string {
	var result []string

	for _, imp := range pkg.Imports {
		if strings.HasPrefix(imp, ModulePrefix) {
			result = append(result, PackageSuffix(imp))
		}
	}

	return result
}

// BuildModel constructs an ArchModel from raw go list output, a classifier function,
// and layer definitions. It excludes wiring roots, sqlc, and mock packages. Packages
// for which the classifier returns an empty layer are skipped.
//
// The subsystems map is initialized empty — subsystem grouping is handled by T3.
func BuildModel(
	packages []GoPackage,
	classify Classifier,
	layers map[string]*LayerInfo,
) *ArchModel {
	archModel := &ArchModel{
		Packages:   make(map[string]*PackageInfo),
		Subsystems: make(map[string]*SubsystemInfo),
		Layers:     layers,
	}

	// First pass: classify all packages and populate the package map.
	for _, pkg := range packages {
		if !strings.HasPrefix(pkg.ImportPath, ModulePrefix) {
			continue
		}

		suffix := PackageSuffix(pkg.ImportPath)
		if IsExcluded(suffix) {
			continue
		}

		layer, summary := classify(suffix)
		if layer == "" {
			continue
		}

		archModel.Packages[suffix] = &PackageInfo{
			ImportPath:   pkg.ImportPath,
			Suffix:       suffix,
			Layer:        layer,
			Summary:      summary,
			GoFiles:      pkg.GoFiles,
			InternalDeps: internalImportSuffixes(pkg),
		}
	}

	// Second pass: compute cross-layer edges from the import graph.
	archModel.Edges = collectCrossLayerEdges(archModel.Packages)

	return archModel
}

// collectCrossLayerEdges finds all imports that cross layer boundaries,
// deduplicated at the layer level. Only considers packages present in the model.
func collectCrossLayerEdges(packages map[string]*PackageInfo) []Edge {
	seen := make(map[Edge]bool)
	var result []Edge

	for _, pkg := range packages {
		for _, depSuffix := range pkg.InternalDeps {
			dep, ok := packages[depSuffix]
			if !ok || pkg.Layer == dep.Layer {
				continue
			}

			e := Edge{From: pkg.Layer, To: dep.Layer}
			if !seen[e] {
				seen[e] = true
				result = append(result, e)
			}
		}
	}

	return result
}
