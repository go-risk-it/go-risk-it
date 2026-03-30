package render

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/go-risk-it/go-risk-it/cmd/archdiagram/model"
)

// taxonomyRowPattern matches table rows in the Layer Taxonomy table of doc-go-spec.md.
// It captures the layer name (backtick-wrapped) from lines like:
// | `api` | Data transfer objects ... | ... |.
var taxonomyRowPattern = regexp.MustCompile(
	"^\\| `([^`]+)` \\|",
)

// CheckTaxonomyConsistency verifies doc-go-spec.md taxonomy matches the generator's
// layer set. Returns an error if the sets differ.
func CheckTaxonomyConsistency(
	repoRoot string,
	modelLayers map[string]*model.LayerInfo,
) error {
	specPath := filepath.Join(repoRoot, "docs", "doc-go-spec.md")

	data, err := os.ReadFile(specPath)
	if err != nil {
		return fmt.Errorf("reading doc-go-spec.md: %w", err)
	}

	docLayers := extractTaxonomyLayers(string(data))
	generatorLayers := extractGeneratorLayers(modelLayers)

	return compareLayers(docLayers, generatorLayers)
}

// extractTaxonomyLayers parses the Layer Taxonomy table from doc-go-spec.md
// and returns the set of layer names (lowercased for comparison).
func extractTaxonomyLayers(content string) map[string]bool {
	layers := make(map[string]bool)

	// Find the Layer Taxonomy section.
	sectionStart := strings.Index(content, "## Layer Taxonomy")
	if sectionStart < 0 {
		return layers
	}

	// Find the table within the section (between the heading and the next ## or end).
	section := content[sectionStart:]

	nextSection := strings.Index(section[1:], "\n## ")
	if nextSection > 0 {
		section = section[:nextSection+1]
	}

	lines := strings.SplitSeq(section, "\n")
	for line := range lines {
		matches := taxonomyRowPattern.FindStringSubmatch(line)
		if len(matches) >= 2 {
			layerName := matches[1]
			layers[strings.ToLower(layerName)] = true
		}
	}

	return layers
}

// extractGeneratorLayers converts the generator's layer map to a lowercased set.
// The "wiring" layer is excluded because it is a doc-go-spec-only concept
// (not represented in the generator's layers map).
func extractGeneratorLayers(modelLayers map[string]*model.LayerInfo) map[string]bool {
	layers := make(map[string]bool, len(modelLayers))

	for name := range modelLayers {
		layers[strings.ToLower(name)] = true
	}

	return layers
}

// compareLayers checks if two layer sets are equivalent.
// The "wiring" layer is allowed in doc-go-spec.md but not in the generator
// (it's a documentation-only concept for fx.Module aggregation roots).
func compareLayers(docLayers, generatorLayers map[string]bool) error {
	docLayersFiltered := filterWiringLayer(docLayers)

	missing := findMissing(generatorLayers, docLayersFiltered)
	extra := findMissing(docLayersFiltered, generatorLayers)

	sort.Strings(missing)
	sort.Strings(extra)

	if len(missing) == 0 && len(extra) == 0 {
		return nil
	}

	var msg strings.Builder

	msg.WriteString(
		"taxonomy mismatch between doc-go-spec.md and generator layers",
	)

	if len(missing) > 0 {
		fmt.Fprintf(&msg,
			"\n  in generator but not in doc-go-spec.md: %v", missing)
	}

	if len(extra) > 0 {
		fmt.Fprintf(&msg,
			"\n  in doc-go-spec.md but not in generator: %v", extra)
	}

	return fmt.Errorf("%s", msg.String())
}

// filterWiringLayer returns a copy of the layer set without the "wiring" entry.
func filterWiringLayer(layers map[string]bool) map[string]bool {
	filtered := make(map[string]bool, len(layers))
	for k, v := range layers {
		if k != "wiring" {
			filtered[k] = v
		}
	}

	return filtered
}

// findMissing returns keys in source that are not present in target.
func findMissing(source, target map[string]bool) []string {
	var result []string

	for key := range source {
		if !target[key] {
			result = append(result, key)
		}
	}

	return result
}
