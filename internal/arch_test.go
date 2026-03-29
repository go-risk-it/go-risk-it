package internal_test

import (
	"encoding/json"
	"go/ast"
	"go/doc/comment"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

const modulePrefix = "github.com/go-risk-it/go-risk-it/internal/"

// goPackage is the subset of `go list -json` fields we need.
type goPackage struct {
	ImportPath string   `json:"ImportPath"` //nolint:tagliatelle // matches go list -json output
	Imports    []string `json:"Imports"`    //nolint:tagliatelle // matches go list -json output
	Dir        string   `json:"Dir"`        //nolint:tagliatelle // matches go list -json output
	GoFiles    []string `json:"GoFiles"`    //nolint:tagliatelle // matches go list -json output
}

// loadPackages runs `go list -json` for the given pattern and returns parsed packages.
func loadPackages(t *testing.T, pattern string) []goPackage {
	t.Helper()

	cmd := exec.CommandContext(t.Context(), "go", "list", "-json", pattern)
	cmd.Dir = ".."
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("go list failed: %v", err)
	}

	var pkgs []goPackage

	dec := json.NewDecoder(strings.NewReader(string(out)))
	for dec.More() {
		var pkg goPackage
		if err := dec.Decode(&pkg); err != nil {
			t.Fatalf("failed to decode go list output: %v", err)
		}

		pkgs = append(pkgs, pkg)
	}

	return pkgs
}

// internalImports filters a package's imports to only those under our module's internal/.
func internalImports(pkg goPackage) []string {
	var result []string

	for _, imp := range pkg.Imports {
		if strings.HasPrefix(imp, modulePrefix) {
			result = append(result, imp)
		}
	}

	return result
}

// hasPrefix checks if an import path starts with any of the given prefixes.
func hasPrefix(importPath string, prefixes ...string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(importPath, prefix) {
			return true
		}
	}

	return false
}

// assertNoImports checks that no package in `pkgs` imports packages matching any of the
// forbidden prefixes.
func assertNoImports(t *testing.T, pkgs []goPackage, forbiddenPrefixes ...string) {
	t.Helper()

	for _, pkg := range pkgs {
		for _, imp := range internalImports(pkg) {
			if hasPrefix(imp, forbiddenPrefixes...) {
				t.Errorf("%s imports forbidden package %s", pkg.ImportPath, imp)
			}
		}
	}
}

// assertNoRawImports checks that no package in `pkgs` imports any of the exact forbidden packages
// (not prefixes — exact match on stdlib or third-party packages).
func assertNoRawImports(t *testing.T, pkgs []goPackage, forbidden ...string) {
	t.Helper()

	forbiddenSet := make(map[string]bool, len(forbidden))
	for _, f := range forbidden {
		forbiddenSet[f] = true
	}

	for _, pkg := range pkgs {
		for _, imp := range pkg.Imports {
			if forbiddenSet[imp] {
				t.Errorf("%s imports forbidden package %s", pkg.ImportPath, imp)
			}
		}
	}
}

// Rule 1: logic/ must never import web/ or api/.
func TestArch_LogicNeverImportsWebOrAPI(t *testing.T) {
	t.Parallel()

	lobbyPkgs := loadPackages(t, "./internal/logic/...")
	gamePkgs := loadPackages(t, "./internal/game/logic/...")
	lobbyPkgs = append(lobbyPkgs, gamePkgs...)
	pkgs := lobbyPkgs

	assertNoImports(t, pkgs,
		modulePrefix+"web/",
		modulePrefix+"api/",
		modulePrefix+"game/api/",
	)
}

// Rule 2: logic/ must never import net/http.
func TestArch_LogicNeverImportsNetHTTP(t *testing.T) {
	t.Parallel()

	lobbyPkgs := loadPackages(t, "./internal/logic/...")
	gamePkgs := loadPackages(t, "./internal/game/logic/...")
	lobbyPkgs = append(lobbyPkgs, gamePkgs...)
	pkgs := lobbyPkgs

	assertNoRawImports(t, pkgs, "net/http")
}

// Rule 3: data/ must never import logic/ or web/.
func TestArch_DataNeverImportsLogicOrWeb(t *testing.T) {
	t.Parallel()

	lobbyPkgs := loadPackages(t, "./internal/data/...")
	gamePkgs := loadPackages(t, "./internal/game/data/...")
	lobbyPkgs = append(lobbyPkgs, gamePkgs...)
	pkgs := lobbyPkgs

	assertNoImports(t, pkgs,
		modulePrefix+"logic/",
		modulePrefix+"game/logic/",
		modulePrefix+"web/",
	)
}

// Rule 4: game/logic/ and logic/lobby/ are mutually isolated.
func TestArch_LogicGameAndLobbyIsolated(t *testing.T) {
	t.Parallel()

	gamePkgs := loadPackages(t, "./internal/game/logic/...")
	assertNoImports(t, gamePkgs, modulePrefix+"logic/lobby/")

	lobbyPkgs := loadPackages(t, "./internal/logic/lobby/...")
	assertNoImports(t, lobbyPkgs,
		modulePrefix+"game/logic/",
		modulePrefix+"logic/game/",
	)
}

// Rule 4b: events/logger is infrastructure — it must never import logic/ or web/.
// Sub-packages events/game/ and events/lobby/ carry domain payloads and may import logic/ types.
func TestArch_EventsRootIsolation(t *testing.T) {
	t.Parallel()

	loggerPkgs := loadPackages(t, "./internal/events/logger")

	assertNoImports(t, loggerPkgs,
		modulePrefix+"logic/",
		modulePrefix+"web/",
	)
}

// Rule 4c: kernel/ must never import game or lobby domain packages.
func TestArch_KernelNeverImportsDomain(t *testing.T) {
	t.Parallel()

	pkgs := loadPackages(t, "./internal/kernel/...")

	assertNoImports(t, pkgs,
		modulePrefix+"logic/",
		modulePrefix+"web/",
		modulePrefix+"game/",
		modulePrefix+"data/game/",
		modulePrefix+"data/lobby/",
		modulePrefix+"events/game/",
		modulePrefix+"events/lobby/",
	)
}

// Rule 5: web/game/ and web/lobby/ are mutually isolated.
func TestArch_WebGameAndLobbyIsolated(t *testing.T) {
	t.Parallel()

	gamePkgs := loadPackages(t, "./internal/web/game/...")
	assertNoImports(t, gamePkgs, modulePrefix+"web/lobby/")

	lobbyPkgs := loadPackages(t, "./internal/web/lobby/...")
	assertNoImports(t, lobbyPkgs, modulePrefix+"web/game/")
}

// containsFile checks if a package has a specific file in its GoFiles list.
func containsFile(pkg goPackage, name string) bool {
	return slices.Contains(pkg.GoFiles, name)
}

// exportedInterfacePattern matches "type <Exported> interface" declarations,
// including generic interfaces like "type Service[T any] interface".
var exportedInterfacePattern = regexp.MustCompile(`type\s+[A-Z]\w*(?:\[.*?\])?\s+interface\b`)

// assertHasExportedInterface verifies that at least one Go source file in the package
// contains an exported interface declaration.
func assertHasExportedInterface(t *testing.T, pkg goPackage) {
	t.Helper()

	for _, file := range pkg.GoFiles {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}

		content, err := os.ReadFile(filepath.Join(pkg.Dir, file))
		if err != nil {
			t.Fatalf("failed to read %s: %v", filepath.Join(pkg.Dir, file), err)
		}

		if exportedInterfacePattern.Match(content) {
			return
		}
	}

	t.Errorf("%s has service.go but no exported interface declaration", pkg.ImportPath)
}

// Rule 6: every logic service package defines at least one exported interface.
func TestArch_LogicServicesDefineExportedInterface(t *testing.T) {
	t.Parallel()

	lobbyPkgs := loadPackages(t, "./internal/logic/...")
	gamePkgs := loadPackages(t, "./internal/game/logic/...")
	lobbyPkgs = append(lobbyPkgs, gamePkgs...)
	pkgs := lobbyPkgs

	for _, pkg := range pkgs {
		if !containsFile(pkg, "service.go") {
			continue
		}

		assertHasExportedInterface(t, pkg)
	}
}

// ─── Phase 1: Layer Boundary Completion ───

// Rule 7: api/ is DTOs-only — it may only import other api/ packages.
func TestArch_APIOnlyImportsAPI(t *testing.T) {
	t.Parallel()

	lobbyPkgs := loadPackages(t, "./internal/api/...")
	gamePkgs := loadPackages(t, "./internal/game/api/...")
	lobbyPkgs = append(lobbyPkgs, gamePkgs...)
	pkgs := lobbyPkgs

	for _, pkg := range pkgs {
		for _, imp := range internalImports(pkg) {
			short := strings.TrimPrefix(imp, modulePrefix)
			if !strings.HasPrefix(short, "api/") && !strings.HasPrefix(short, "game/api") {
				t.Errorf("%s imports non-api package %s", pkg.ImportPath, imp)
			}
		}
	}
}

// Rule 8: infrastructure packages (kernel/config/, kernel/metrics/, kernel/rand/) have no internal imports.
// kernel/slog/ is excepted — it legitimately imports config + ctx.
func TestArch_InfrastructureIsolation(t *testing.T) {
	t.Parallel()

	infraPatterns := []string{
		"./internal/kernel/config/...",
		"./internal/kernel/metrics/...",
		"./internal/game/rand/...",
	}

	for _, pattern := range infraPatterns {
		pkgs := loadPackages(t, pattern)

		for _, pkg := range pkgs {
			intImps := internalImports(pkg)
			if len(intImps) > 0 {
				t.Errorf("%s has internal imports %v (infrastructure must be leaf packages)",
					pkg.ImportPath, intImps)
			}
		}
	}
}

// Rule 9: testonly/ must never be imported by production code.
func TestArch_TestOnlyNeverImportedByProduction(t *testing.T) {
	t.Parallel()

	pkgs := loadPackages(t, "./internal/...")

	for _, pkg := range pkgs {
		short := strings.TrimPrefix(pkg.ImportPath, modulePrefix)
		if strings.HasPrefix(short, "testonly") {
			continue
		}

		assertNoImports(t, []goPackage{pkg}, modulePrefix+"testonly")
	}
}

// Rule 10: data/ must never import net/http.
func TestArch_DataNeverImportsNetHTTP(t *testing.T) {
	t.Parallel()

	lobbyPkgs := loadPackages(t, "./internal/data/...")
	gamePkgs := loadPackages(t, "./internal/game/data/...")
	lobbyPkgs = append(lobbyPkgs, gamePkgs...)
	pkgs := lobbyPkgs

	assertNoRawImports(t, pkgs, "net/http")
}

// Rule 11: web/ must never import data/ querier packages (must go through logic/).
// Importing data/*/sqlc for model types is allowed.
func TestArch_WebNeverImportsDataQuerier(t *testing.T) {
	t.Parallel()

	pkgs := loadPackages(t, "./internal/web/...")

	for _, pkg := range pkgs {
		for _, imp := range internalImports(pkg) {
			dataPart := strings.TrimPrefix(imp, modulePrefix+"data/")
			if dataPart != imp && strings.Contains(dataPart, "/db") {
				t.Errorf("%s imports data querier %s (web must go through logic)",
					pkg.ImportPath, imp)
			}

			gameDataPart := strings.TrimPrefix(imp, modulePrefix+"game/data/")
			if gameDataPart != imp && strings.Contains(gameDataPart, "/db") {
				t.Errorf("%s imports game data querier %s (web must go through logic)",
					pkg.ImportPath, imp)
			}
		}
	}
}

// ─── Phase 4: Import Guards (stdlib anti-patterns) ───

// Rule 12: no package may import stdlib "log" (use log/slog or internal/slog).
func TestArch_NoStdlibLog(t *testing.T) {
	t.Parallel()

	pkgs := loadPackages(t, "./internal/...")

	assertNoRawImports(t, pkgs, "log")
}

// Rule 13: no package may import "math/rand" (use math/rand/v2 or internal/rand).
func TestArch_NoOldMathRand(t *testing.T) {
	t.Parallel()

	pkgs := loadPackages(t, "./internal/...")

	assertNoRawImports(t, pkgs, "math/rand")
}

// ─── Phase 2: Package Quality Metrics ───

// archBaseline defines the metric ceilings for package quality ratcheting.
type archBaseline struct {
	MaxExportsPerPackage int `json:"maxExportsPerPackage"`
	MaxFanOut            int `json:"maxFanOut"`
	MaxFilesPerPackage   int `json:"maxFilesPerPackage"`
	MinDocGoCount        int `json:"minDocGoCount"`
}

func loadBaseline(t *testing.T) archBaseline {
	t.Helper()

	data, err := os.ReadFile("arch_baseline.json")
	if err != nil {
		t.Fatalf("failed to read arch_baseline.json: %v", err)
	}

	var baseline archBaseline
	if err := json.Unmarshal(data, &baseline); err != nil {
		t.Fatalf("failed to parse arch_baseline.json: %v", err)
	}

	return baseline
}

// isGeneratedPackage returns true for packages that contain auto-generated code
// (e.g. sqlc, mocks) which should be excluded from quality metrics.
func isGeneratedPackage(importPath string) bool {
	short := strings.TrimPrefix(importPath, modulePrefix)

	return strings.Contains(short, "/sqlc") || strings.Contains(short, "/mocks")
}

// countExports counts exported symbols in a package using go/ast.
//
//nolint:gocognit,cyclop // AST traversal is inherently nested
func countExports(
	t *testing.T,
	pkg goPackage,
) int {
	t.Helper()

	total := 0
	fset := token.NewFileSet()

	for _, goFile := range pkg.GoFiles {
		if strings.HasSuffix(goFile, "_test.go") {
			continue
		}

		path := filepath.Join(pkg.Dir, goFile)

		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			t.Fatalf("failed to parse %s: %v", path, err)
		}

		for _, decl := range file.Decls {
			switch typedDecl := decl.(type) {
			case *ast.GenDecl:
				for _, spec := range typedDecl.Specs {
					switch typedSpec := spec.(type) {
					case *ast.TypeSpec:
						if ast.IsExported(typedSpec.Name.Name) {
							total++
						}
					case *ast.ValueSpec:
						for _, name := range typedSpec.Names {
							if ast.IsExported(name.Name) {
								total++
							}
						}
					}
				}
			case *ast.FuncDecl:
				if ast.IsExported(typedDecl.Name.Name) {
					total++
				}
			}
		}
	}

	return total
}

// Rule 14: no package exceeds the export ceiling (excluding generated packages).
func TestArch_MaxExportsPerPackage(t *testing.T) {
	t.Parallel()

	baseline := loadBaseline(t)
	pkgs := loadPackages(t, "./internal/...")

	for _, pkg := range pkgs {
		if isGeneratedPackage(pkg.ImportPath) {
			continue
		}

		exports := countExports(t, pkg)
		if exports > baseline.MaxExportsPerPackage {
			t.Errorf("%s has %d exports (ceiling: %d)",
				pkg.ImportPath, exports, baseline.MaxExportsPerPackage)
		}
	}
}

// Rule 15: no package exceeds the internal fan-out ceiling.
func TestArch_MaxFanOut(t *testing.T) {
	t.Parallel()

	baseline := loadBaseline(t)
	pkgs := loadPackages(t, "./internal/...")

	for _, pkg := range pkgs {
		fanOut := len(internalImports(pkg))
		if fanOut > baseline.MaxFanOut {
			t.Errorf("%s has fan-out %d (ceiling: %d)",
				pkg.ImportPath, fanOut, baseline.MaxFanOut)
		}
	}
}

// Rule 16: no package exceeds the file count ceiling (excluding generated packages).
func TestArch_MaxFilesPerPackage(t *testing.T) {
	t.Parallel()

	baseline := loadBaseline(t)
	pkgs := loadPackages(t, "./internal/...")

	for _, pkg := range pkgs {
		if isGeneratedPackage(pkg.ImportPath) {
			continue
		}

		files := len(pkg.GoFiles)
		if files > baseline.MaxFilesPerPackage {
			t.Errorf("%s has %d files (ceiling: %d)",
				pkg.ImportPath, files, baseline.MaxFilesPerPackage)
		}
	}
}

// ─── Phase 5: Headroom Reporting ───

// TestArch_ReportHeadroom logs current headroom for each metric (informational, never fails).
func TestArch_ReportHeadroom(t *testing.T) {
	t.Parallel()

	baseline := loadBaseline(t)
	pkgs := loadPackages(t, "./internal/...")

	maxExports, maxFanOut, maxFiles := 0, 0, 0
	maxExportsPkg, maxFanOutPkg, maxFilesPkg := "", "", ""

	for _, pkg := range pkgs {
		if !isGeneratedPackage(pkg.ImportPath) {
			exports := countExports(t, pkg)
			if exports > maxExports {
				maxExports = exports
				maxExportsPkg = pkg.ImportPath
			}

			files := len(pkg.GoFiles)
			if files > maxFiles {
				maxFiles = files
				maxFilesPkg = pkg.ImportPath
			}
		}

		fanOut := len(internalImports(pkg))
		if fanOut > maxFanOut {
			maxFanOut = fanOut
			maxFanOutPkg = pkg.ImportPath
		}
	}

	t.Logf(
		"exports:  actual %d (%s), ceiling %d — headroom %d",
		maxExports,
		maxExportsPkg,
		baseline.MaxExportsPerPackage,
		baseline.MaxExportsPerPackage-maxExports,
	)
	t.Logf("fan-out:  actual %d (%s), ceiling %d — headroom %d",
		maxFanOut, maxFanOutPkg, baseline.MaxFanOut, baseline.MaxFanOut-maxFanOut)
	t.Logf("files:    actual %d (%s), ceiling %d — headroom %d",
		maxFiles, maxFilesPkg, baseline.MaxFilesPerPackage, baseline.MaxFilesPerPackage-maxFiles)
}

// ─── Phase 6: Doc.go Enforcement ───

// expectedLayer maps package suffix to its architectural layer name.
// The suffix is the import path after "github.com/go-risk-it/go-risk-it/internal/".
//
//nolint:gochecknoglobals // test-only mapping used by doc.go validation rules
var expectedLayer = map[string]string{
	// game/api
	"game/api":               "API",
	"game/api/messaging":     "API",
	"game/api/rest/request":  "API",
	"game/api/rest/response": "API",

	// api (lobby)
	"api/lobby/messaging":     "API",
	"api/lobby/rest/request":  "API",
	"api/lobby/rest/response": "API",

	// kernel
	"kernel":                    "Kernel",
	"kernel/bus":                "Kernel",
	"kernel/config":             "Kernel",
	"kernel/ctx":                "Kernel",
	"kernel/data":               "Kernel",
	"kernel/data/migration":     "Kernel",
	"kernel/data/pool":          "Kernel",
	"kernel/errors":             "Kernel",
	"kernel/metrics":            "Kernel",
	"kernel/otelsetup":          "Kernel",
	"kernel/slog":               "Kernel",
	"kernel/upgradablerw_mutex": "Kernel",

	// game domain
	"game/data/db":             "Data",
	"game/events":              "Events-domain",
	"game/tracing":             "Logic",
	"game/rand":                "Logic",
	"game/logic/config":        "Logic",
	"game/logic/metrics":       "Logic",
	"game/consumers":           "Web",
	"game/consumers/converter": "Web",

	// data (lobby)
	"data/lobby/db": "Data",

	// events
	"events/logger": "Events",
	"events/lobby":  "Events-domain",

	// test
	"testing/invariant": "Test",
	"testonly":          "Test",
}

// layerFromPrefix derives the expected layer for a package suffix using prefix matching.
// It first checks the explicit mapping, then falls back to prefix-based rules.
//
//nolint:cyclop // prefix matching requires branching per layer
func layerFromPrefix(suffix string) string {
	if layer, ok := expectedLayer[suffix]; ok {
		return layer
	}

	switch {
	case strings.HasPrefix(suffix, "game/api/"):
		return "API"
	case strings.HasPrefix(suffix, "game/data/"):
		return "Data"
	case strings.HasPrefix(suffix, "game/events"):
		return "Events-domain"
	case strings.HasPrefix(suffix, "game/logic/") || strings.HasPrefix(suffix, "game/tracing") ||
		strings.HasPrefix(suffix, "game/rand"):
		return "Logic"
	case strings.HasPrefix(suffix, "api/"):
		return "API"
	case strings.HasPrefix(suffix, "kernel/"):
		return "Kernel"
	case strings.HasPrefix(suffix, "data/"):
		return "Data"
	case strings.HasPrefix(suffix, "events/lobby"):
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

// packageSuffix returns the import path after the module prefix.
func packageSuffix(importPath string) string {
	return strings.TrimPrefix(importPath, modulePrefix)
}

// wiringRoots are single-file fx.Module aggregation packages that exist solely
// to compose sub-package modules. These are excluded from doc.go requirements.
//
//nolint:gochecknoglobals // test-only set used by doc.go validation rules
var wiringRoots = map[string]bool{
	"":                        true, // internal root
	"kernel":                  true,
	"logic":                   true,
	"game/logic":              true,
	"game/logic/move":         true,
	"game/logic/move/service": true,
	"game/data":               true,
	"logic/lobby":             true,
	"data":                    true,
	"data/lobby":              true,
	"web":                     true,
	"web/game":                true,
	"web/lobby":               true,
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

// hasDocGoFile checks if a package has a dedicated doc.go file.
func hasDocGoFile(pkg goPackage) bool {
	return slices.Contains(pkg.GoFiles, "doc.go")
}

// Rule 19: doc.go sections — full-tier packages must have Layer, Key Types, Dependencies headings.
// Lightweight packages must have Layer. Only checks packages that HAVE a doc.go file with headings,
// indicating they were written to the living architecture spec.
//
//nolint:cyclop // tier-aware section validation has inherent branching
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
