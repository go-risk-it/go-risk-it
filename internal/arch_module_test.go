package internal_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ─── Module Conformance ───

// requiredSubdirs lists the subdirectories every domain module must contain.
//
//nolint:gochecknoglobals // test-only constant
var requiredSubdirs = []string{"api", "ctx", "events", "internal/logic"}

// discoverModules scans internal/ for directories that have both api/ and ctx/
// subdirectories. Those are domain modules.
func discoverModules(t *testing.T) []string {
	t.Helper()

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("failed to read internal/ directory: %v", err)
	}

	var modules []string

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()

		hasAPI := dirExists(filepath.Join(name, "api"))
		hasCTX := dirExists(filepath.Join(name, "ctx"))

		if hasAPI && hasCTX {
			modules = append(modules, name)
		}
	}

	if len(modules) == 0 {
		t.Fatal("no domain modules discovered — expected at least game and lobby")
	}

	return modules
}

// dirExists reports whether the given path exists and is a directory.
func dirExists(path string) bool {
	info, err := os.Stat(path)

	return err == nil && info.IsDir()
}

// TestModuleConformance_DiscoverModules verifies that auto-discovery finds
// both the game and lobby modules.
func TestModuleConformance_DiscoverModules(t *testing.T) {
	t.Parallel()

	modules := discoverModules(t)

	expected := map[string]bool{"game": false, "lobby": false}

	for _, mod := range modules {
		if _, ok := expected[mod]; ok {
			expected[mod] = true
		}
	}

	for mod, found := range expected {
		if !found {
			t.Errorf("expected module %q not discovered", mod)
		}
	}

	t.Logf("discovered modules: %v", modules)
}

// TestModuleConformance_RequiredStructure validates that each discovered module
// contains the required subdirectories: api/, ctx/, events/, internal/logic/.
func TestModuleConformance_RequiredStructure(t *testing.T) {
	t.Parallel()

	modules := discoverModules(t)

	for _, mod := range modules {
		for _, sub := range requiredSubdirs {
			path := filepath.Join(mod, sub)
			if !dirExists(path) {
				t.Errorf("module %s missing required subdirectory %s", mod, sub)
			}
		}
	}
}

// TestModuleConformance_NoPublicImportsInternal validates that no public package
// (api/, ctx/, events/) within a module imports that module's internal/ packages.
// Public packages form the module's API contract and must not leak implementation details.
func TestModuleConformance_NoPublicImportsInternal(t *testing.T) {
	t.Parallel()

	modules := discoverModules(t)

	for _, mod := range modules {
		publicPatterns := []string{
			"./internal/" + mod + "/api/...",
			"./internal/" + mod + "/ctx/...",
			"./internal/" + mod + "/events/...",
		}

		internalPrefix := modulePrefix + mod + "/internal/"

		for _, pattern := range publicPatterns {
			pkgs := loadPackages(t, pattern)

			for _, pkg := range pkgs {
				for _, imp := range internalImports(pkg) {
					if strings.HasPrefix(imp, internalPrefix) {
						t.Errorf(
							"%s (public) imports internal package %s",
							pkg.ImportPath, imp,
						)
					}
				}
			}
		}
	}
}

// crossModuleAllowedPrefixes returns the import prefixes that a module is allowed
// to import from another module. Only the other module's events/ package is permitted.
func crossModuleAllowedPrefixes(otherModule string) []string {
	return []string{
		modulePrefix + otherModule + "/events",
		modulePrefix + otherModule + "/api",
	}
}

// TestModuleConformance_CrossModuleEventsOnly validates that when one module
// imports from another module, only the events/ and api/ packages are used.
// This enforces the cross-module contract boundary.
func TestModuleConformance_CrossModuleEventsOnly(t *testing.T) {
	t.Parallel()

	modules := discoverModules(t)

	for _, mod := range modules {
		pkgs := loadPackages(t, "./internal/"+mod+"/...")

		for _, other := range modules {
			if other == mod {
				continue
			}

			otherPrefix := modulePrefix + other + "/"
			allowed := crossModuleAllowedPrefixes(other)

			for _, pkg := range pkgs {
				for _, imp := range internalImports(pkg) {
					if !strings.HasPrefix(imp, otherPrefix) {
						continue
					}

					if hasPrefix(imp, allowed...) {
						continue
					}

					t.Errorf(
						"%s imports %s from module %s — cross-module imports "+
							"must be limited to events/ and api/ packages",
						pkg.ImportPath, imp, other,
					)
				}
			}
		}
	}
}

// isDomainModule reports whether the import path belongs to one of the
// discovered domain modules (not shared infrastructure like kernel/, web/).
func isDomainModule(importPath string, modules []string) (string, bool) {
	suffix := strings.TrimPrefix(importPath, modulePrefix)

	for _, mod := range modules {
		if strings.HasPrefix(suffix, mod+"/") || suffix == mod {
			return mod, true
		}
	}

	return "", false
}

// TestModuleConformance_IndependenceScore computes and validates the module
// independence score for each discovered module.
//
// The score measures domain-level independence: of all internal/ imports from
// within a module, what fraction stays within the module or goes to shared
// infrastructure (kernel/, web/, testing/, testonly/)? The remaining fraction
// is cross-domain-module coupling.
//
// Score = 1 - (cross-module domain imports / total internal imports).
// Threshold: 90%.
func TestModuleConformance_IndependenceScore(t *testing.T) {
	t.Parallel()

	const threshold = 0.90

	modules := discoverModules(t)

	for _, mod := range modules {
		pkgs := loadPackages(t, "./internal/"+mod+"/...")

		var totalInternal, crossModuleDomain int

		for _, pkg := range pkgs {
			for _, imp := range internalImports(pkg) {
				totalInternal++

				owner, isDomain := isDomainModule(imp, modules)
				if isDomain && owner != mod {
					crossModuleDomain++
				}
			}
		}

		if totalInternal == 0 {
			t.Logf("module %s: no internal imports (trivially independent)", mod)

			continue
		}

		score := 1.0 - float64(crossModuleDomain)/float64(totalInternal)

		t.Logf(
			"module %s independence: %.1f%% (%d total internal imports, %d cross-module)",
			mod, score*100, totalInternal, crossModuleDomain,
		)

		if score < threshold {
			t.Errorf(
				"module %s independence score %.1f%% is below threshold %.0f%% "+
					"(%d of %d imports cross into other domain modules)",
				mod, score*100, threshold*100,
				crossModuleDomain, totalInternal,
			)
		}
	}
}
