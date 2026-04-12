package internal_test

import (
	"go/doc/comment"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// expectedLayer maps package suffix to its architectural layer name.
// The suffix is the import path after "github.com/go-risk-it/go-risk-it/internal/".
//
//nolint:gochecknoglobals // test-only mapping used by doc.go validation rules
var expectedLayer = map[string]string{
	// game/api
	"game/api":               "API",
	"game/api/messaging":     "API",
	"game/api/moves/attack":  "API",
	"game/api/moves/cards":   "API",
	"game/api/rest/request":  "API",
	"game/api/rest/response": "API",
	"game/api/snapshot":      "API",

	// lobby/api
	"lobby/api/messaging":     "API",
	"lobby/api/rest/request":  "API",
	"lobby/api/rest/response": "API",
	"lobby/api/snapshot":      "API",

	// kernel
	"kernel":                    "Kernel",
	"kernel/bus":                "Kernel",
	"kernel/config":             "Kernel",
	"kernel/ctx":                "Kernel",
	"kernel/data":               "Kernel",
	"kernel/data/migration":     "Kernel",
	"kernel/data/pool":          "Kernel",
	"kernel/errors":             "Kernel",
	"kernel/logger":             "Kernel",
	"kernel/metrics":            "Kernel",
	"kernel/observe":            "Kernel",
	"kernel/otelsetup":          "Kernel",
	"kernel/slog":               "Kernel",
	"kernel/upgradablerw_mutex": "Kernel",

	// game domain
	"game/ctx":    "Game-domain",
	"game/events": "Events-domain",
	"game/ws":     "Web",

	// game/internal
	"game/internal/config":   "Game-support",
	"game/internal/data/db":  "Data",
	"game/internal/handlers": "Game-support",
	"game/internal/rand":     "Logic",
	"game/internal/snapshot": "Game-support",

	// game/web
	"game/web/routes": "Web",

	// lobby domain
	"lobby/ctx": "Lobby-domain",
	"lobby/ws":  "Web",

	// lobby/internal
	"lobby/internal/data/db": "Data",

	// lobby/web
	"lobby/web":        "Web",
	"lobby/web/routes": "Web",

	// events
	"lobby/events": "Events-domain",

	// test
	"game/testing":           "Test",
	"game/testing/invariant": "Test",
	"testonly":               "Test",
}

// layerFromPrefix derives the expected layer for a package suffix using prefix matching.
// It first checks the explicit mapping, then falls back to prefix-based rules.
func layerFromPrefix(suffix string) string {
	if layer, ok := expectedLayer[suffix]; ok {
		return layer
	}

	switch {
	// game module
	case strings.HasPrefix(suffix, "game/api/"):
		return "API"
	case strings.HasPrefix(suffix, "game/ctx"):
		return "Game-domain"
	case strings.HasPrefix(suffix, "game/events"):
		return "Events-domain"
	case strings.HasPrefix(suffix, "game/internal/data/"):
		return "Data"
	case strings.HasPrefix(suffix, "game/internal/config") ||
		strings.HasPrefix(suffix, "game/internal/handlers") ||
		strings.HasPrefix(suffix, "game/internal/snapshot"):
		return "Game-support"
	case strings.HasPrefix(suffix, "game/internal/logic/") ||
		strings.HasPrefix(suffix, "game/internal/rand"):
		return "Logic"
	case strings.HasPrefix(suffix, "game/web/") ||
		suffix == "game/web":
		return "Web"
	case strings.HasPrefix(suffix, "game/ws"):
		return "Web"

	// lobby module
	case strings.HasPrefix(suffix, "lobby/api/"):
		return "API"
	case strings.HasPrefix(suffix, "lobby/ctx"):
		return "Lobby-domain"
	case strings.HasPrefix(suffix, "lobby/events"):
		return "Events-domain"
	case strings.HasPrefix(suffix, "lobby/internal/data/"):
		return "Data"
	case strings.HasPrefix(suffix, "lobby/internal/logic/"):
		return "Logic"
	case strings.HasPrefix(suffix, "lobby/web/") ||
		suffix == "lobby/web":
		return "Web"
	case strings.HasPrefix(suffix, "lobby/ws"):
		return "Web"

	// shared infrastructure
	case strings.HasPrefix(suffix, "kernel/"):
		return "Kernel"
	case strings.HasPrefix(suffix, "web/"):
		return "Web"
	case strings.HasPrefix(suffix, "events/lobby"):
		return "Events-domain"
	case strings.HasPrefix(suffix, "events/"):
		return "Events"

	// test and loadtest
	case strings.HasPrefix(suffix, "testing/") || suffix == "testonly":
		return "Test"
	case strings.HasPrefix(suffix, "loadtest/"):
		return "Loadtest"
	default:
		return ""
	}
}

// wiringRoots are single-file fx.Module aggregation packages that exist solely
// to compose sub-package modules. These are excluded from doc.go requirements.
//
//nolint:gochecknoglobals // test-only set used by doc.go validation rules
var wiringRoots = map[string]bool{
	"":                                 true, // internal root
	"kernel":                           true,
	"game":                             true,
	"game/internal/data":               true,
	"game/internal/logic":              true,
	"game/internal/logic/move":         true,
	"game/internal/logic/move/service": true,
	"game/web":                         true,
	"lobby":                            true,
	"lobby/internal/data":              true,
	"lobby/internal/logic":             true,
	"web":                              true,
}

// isWiringRoot returns true for known fx.Module aggregation packages.
func isWiringRoot(pkg goPackage) bool {
	suffix := packageSuffix(pkg.ImportPath)

	return wiringRoots[suffix]
}

// isFullTier returns true for packages that define a service boundary:
// has service.go, more than 2 non-test Go files, and is not wiring/generated/test.
func isFullTier(pkg goPackage) bool {
	if !containsFile(pkg, "service.go") {
		return false
	}

	if isWiringRoot(pkg) || isGeneratedPackage(pkg.ImportPath) {
		return false
	}

	return len(pkg.GoFiles) > 2
}

// parsePackageDoc reads the doc.go (or first .go file) in a package directory,
// extracts the package comment, and parses it with go/doc/comment.Parser.
// Returns nil if no package comment exists.
func parsePackageDoc(t *testing.T, pkg goPackage) *comment.Doc {
	t.Helper()

	// Prefer doc.go, fall back to first non-test .go file
	candidates := []string{"doc.go"}
	for _, f := range pkg.GoFiles {
		if f != "doc.go" && !strings.HasSuffix(f, "_test.go") {
			candidates = append(candidates, f)
		}
	}

	fset := token.NewFileSet()

	for _, candidate := range candidates {
		path := filepath.Join(pkg.Dir, candidate)

		if _, err := os.Stat(path); err != nil {
			continue
		}

		file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			t.Fatalf("failed to parse %s: %v", path, err)
		}

		if file.Doc == nil || file.Doc.Text() == "" {
			continue
		}

		var p comment.Parser
		doc := p.Parse(file.Doc.Text())

		return doc
	}

	return nil
}

// extractHeadings returns the text of all headings in a parsed doc comment.
func extractHeadings(doc *comment.Doc) []string {
	var headings []string

	for _, block := range doc.Content {
		if h, ok := block.(*comment.Heading); ok {
			headings = append(headings, textContent(h.Text))
		}
	}

	return headings
}

// textContent renders a slice of comment.Text elements to a plain string.
func textContent(texts []comment.Text) string {
	var builder strings.Builder

	for _, text := range texts {
		switch val := text.(type) {
		case comment.Plain:
			builder.WriteString(string(val))
		case comment.Italic:
			builder.WriteString(string(val))
		case *comment.Link:
			builder.WriteString(textContent(val.Text))
		case *comment.DocLink:
			builder.WriteString(textContent(val.Text))
		}
	}

	return builder.String()
}

// paragraphText returns the text content of a Paragraph block.
func paragraphText(p *comment.Paragraph) string {
	return textContent(p.Text)
}

// findParagraphAfterHeading returns the first Paragraph block following a heading
// with the given text. Returns nil if not found.
func findParagraphAfterHeading(doc *comment.Doc, heading string) *comment.Paragraph {
	for idx, block := range doc.Content {
		h, ok := block.(*comment.Heading)
		if !ok {
			continue
		}

		if textContent(h.Text) != heading {
			continue
		}

		// Look for the next Paragraph after this heading
		for j := idx + 1; j < len(doc.Content); j++ {
			if p, ok := doc.Content[j].(*comment.Paragraph); ok {
				return p
			}
			// Stop if we hit another heading
			if _, ok := doc.Content[j].(*comment.Heading); ok {
				break
			}
		}
	}

	return nil
}

// hasDocGoFile checks if a package has a dedicated doc.go file.
func hasDocGoFile(pkg goPackage) bool {
	return slices.Contains(pkg.GoFiles, "doc.go")
}

// Rule 18: doc.go existence ratchet — count of packages with doc.go/package comment
// must meet or exceed the baseline minimum.
func TestArch_DocGoExists(t *testing.T) {
	t.Parallel()

	baseline := loadBaseline(t)
	pkgs := loadPackages(t, "./internal/...")

	docCount := 0
	var missing []string

	for _, pkg := range pkgs {
		if isGeneratedPackage(pkg.ImportPath) {
			continue
		}

		suffix := packageSuffix(pkg.ImportPath)
		if suffix == "" {
			continue // internal root
		}

		if isWiringRoot(pkg) {
			continue
		}

		doc := parsePackageDoc(t, pkg)
		if doc != nil {
			docCount++
		} else {
			missing = append(missing, pkg.ImportPath)
		}
	}

	if docCount < baseline.MinDocGoCount {
		t.Errorf("doc.go count %d is below baseline minimum %d", docCount, baseline.MinDocGoCount)
	}

	t.Logf("doc.go coverage: %d packages documented, %d missing", docCount, len(missing))

	for _, m := range missing {
		t.Logf("  missing doc.go: %s", m)
	}
}

// Rule 19: doc.go sections — full-tier packages must have Layer, Key Types, Dependencies headings.
// Lightweight packages must have Layer. Only checks packages that HAVE a doc.go file with headings,
// indicating they were written to the living architecture spec.
func TestArch_DocGoSections(t *testing.T) {
	t.Parallel()

	pkgs := loadPackages(t, "./internal/...")

	for _, pkg := range pkgs {
		if isGeneratedPackage(pkg.ImportPath) {
			continue
		}

		suffix := packageSuffix(pkg.ImportPath)
		if suffix == "" {
			continue
		}

		if isWiringRoot(pkg) {
			continue
		}

		doc := parsePackageDoc(t, pkg)
		if doc == nil {
			continue // no package comment at all
		}

		if !hasDocGoFile(pkg) {
			continue // only enforce sections for dedicated doc.go files
		}

		headings := extractHeadings(doc)
		if !slices.Contains(headings, "Layer") {
			continue // pre-spec doc.go without Layer heading — skip
		}

		headingSet := make(map[string]bool, len(headings))

		for _, h := range headings {
			headingSet[h] = true
		}

		if isFullTier(pkg) {
			for _, required := range []string{"Layer", "Key Types", "Dependencies"} {
				if !headingSet[required] {
					t.Errorf(
						"%s (full tier) missing required heading # %s",
						pkg.ImportPath,
						required,
					)
				}
			}
		} else if !headingSet["Layer"] {
			t.Errorf("%s (lightweight tier) missing required heading # Layer", pkg.ImportPath)
		}
	}
}

// Rule 20: doc.go layer accuracy — the paragraph after # Layer must start with
// the expected layer name for that package.
func TestArch_DocGoLayer(t *testing.T) {
	t.Parallel()

	pkgs := loadPackages(t, "./internal/...")

	for _, pkg := range pkgs {
		if isGeneratedPackage(pkg.ImportPath) {
			continue
		}

		suffix := packageSuffix(pkg.ImportPath)
		if suffix == "" {
			continue
		}

		if isWiringRoot(pkg) {
			continue
		}

		doc := parsePackageDoc(t, pkg)
		if doc == nil {
			continue
		}

		headings := extractHeadings(doc)
		if !slices.Contains(headings, "Layer") {
			continue // no Layer heading to check
		}

		expected := layerFromPrefix(suffix)
		if expected == "" {
			t.Logf("no expected layer mapping for %s, skipping", pkg.ImportPath)

			continue
		}

		para := findParagraphAfterHeading(doc, "Layer")
		if para == nil {
			t.Errorf("%s has # Layer heading but no paragraph after it", pkg.ImportPath)

			continue
		}

		text := paragraphText(para)
		if !strings.HasPrefix(text, expected) {
			t.Errorf("%s # Layer paragraph starts with %q, expected prefix %q",
				pkg.ImportPath, text, expected)
		}
	}
}
