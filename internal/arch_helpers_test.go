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

// containsFile checks if a package has a specific file in its GoFiles list.
func containsFile(pkg goPackage, name string) bool {
	return slices.Contains(pkg.GoFiles, name)
}

// exportedInterfacePattern matches "type <Exported> interface" declarations,
// including generic interfaces like "type Service[T any] interface".
//
//nolint:gochecknoglobals // test-only regex used by interface assertion rules
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

// isGeneratedPackage returns true for packages that contain auto-generated code
// (e.g. sqlc, mocks) which should be excluded from quality metrics.
func isGeneratedPackage(importPath string) bool {
	short := strings.TrimPrefix(importPath, modulePrefix)

	return strings.Contains(short, "/sqlc") || strings.Contains(short, "/mocks")
}

// archBaseline defines the metric ceilings for package quality ratcheting.
type archBaseline struct {
	MaxExportsPerPackage     int `json:"maxExportsPerPackage"`
	MaxFanOut                int `json:"maxFanOut"`
	MaxFilesPerPackage       int `json:"maxFilesPerPackage"`
	MinDocGoCount            int `json:"minDocGoCount"`
	KernelMaxProductionFiles int `json:"kernelMaxProductionFiles"`
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

// packageSuffix returns the import path after the module prefix.
func packageSuffix(importPath string) string {
	return strings.TrimPrefix(importPath, modulePrefix)
}
