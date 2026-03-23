package internal_test

import (
	"encoding/json"
	"os/exec"
	"strings"
	"testing"
)

const modulePrefix = "github.com/go-risk-it/go-risk-it/internal/"

// goPackage is the subset of `go list -json` fields we need.
type goPackage struct {
	ImportPath string   `json:"ImportPath"` //nolint:tagliatelle // matches go list -json output
	Imports    []string `json:"Imports"`    //nolint:tagliatelle // matches go list -json output
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
