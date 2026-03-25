package internal_test

import (
	"encoding/json"
	"go/ast"
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

	pkgs := loadPackages(t, "./internal/logic/...")

	assertNoImports(t, pkgs,
		modulePrefix+"web/",
		modulePrefix+"api/",
	)
}

// Rule 2: logic/ must never import net/http.
func TestArch_LogicNeverImportsNetHTTP(t *testing.T) {
	t.Parallel()

	pkgs := loadPackages(t, "./internal/logic/...")

	assertNoRawImports(t, pkgs, "net/http")
}

// Rule 3: data/ must never import logic/ or web/.
func TestArch_DataNeverImportsLogicOrWeb(t *testing.T) {
	t.Parallel()

	pkgs := loadPackages(t, "./internal/data/...")

	assertNoImports(t, pkgs,
		modulePrefix+"logic/",
		modulePrefix+"web/",
	)
}

// Rule 4: logic/game/ and logic/lobby/ are mutually isolated.
func TestArch_LogicGameAndLobbyIsolated(t *testing.T) {
	t.Parallel()

	gamePkgs := loadPackages(t, "./internal/logic/game/...")
	assertNoImports(t, gamePkgs, modulePrefix+"logic/lobby/")

	lobbyPkgs := loadPackages(t, "./internal/logic/lobby/...")
	assertNoImports(t, lobbyPkgs, modulePrefix+"logic/game/")
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

	pkgs := loadPackages(t, "./internal/logic/...")

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

	pkgs := loadPackages(t, "./internal/api/...")

	for _, pkg := range pkgs {
		for _, imp := range internalImports(pkg) {
			if !hasPrefix(imp, modulePrefix+"api/") {
				t.Errorf("%s imports non-api package %s", pkg.ImportPath, imp)
			}
		}
	}
}

// Rule 8: infrastructure packages (config/, metrics/, rand/) have no internal imports.
// slog/ is excepted — it legitimately imports config + ctx.
func TestArch_InfrastructureIsolation(t *testing.T) {
	t.Parallel()

	infraPatterns := []string{
		"./internal/config/...",
		"./internal/metrics/...",
		"./internal/rand/...",
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

	pkgs := loadPackages(t, "./internal/data/...")

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
			if dataPart == imp {
				continue // not a data/ import
			}

			if strings.Contains(dataPart, "/db") {
				t.Errorf("%s imports data querier %s (web must go through logic)",
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
