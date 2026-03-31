package internal_test

import (
	"strings"
	"testing"
)

// ─── Rules 14–16: Package Quality Metrics ───

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

// ─── Headroom Report ───

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

// ─── Rules 21–22: Kernel Ratchets ───

// Rule 21: kernel production file count must not exceed the ceiling.
func TestArch_KernelProductionFileCeiling(t *testing.T) {
	t.Parallel()

	baseline := loadBaseline(t)
	pkgs := loadPackages(t, "./internal/kernel/...")

	fileCount := 0

	for _, pkg := range pkgs {
		for _, goFile := range pkg.GoFiles {
			if strings.HasSuffix(goFile, "_test.go") || goFile == "doc.go" {
				continue
			}

			fileCount++
		}
	}

	if fileCount > baseline.KernelMaxProductionFiles {
		t.Errorf("kernel has %d production files (ceiling: %d) — ratchet exceeded",
			fileCount, baseline.KernelMaxProductionFiles)
	}

	t.Logf("kernel production files: %d (ceiling: %d, headroom: %d)",
		fileCount, baseline.KernelMaxProductionFiles, baseline.KernelMaxProductionFiles-fileCount)
}

// kernelMultiConsumerAllowlist lists kernel sub-packages that are allowed to have
// fewer than 2 consumer groups. These are infrastructure packages consumed only at
// the composition root or by a single module that has no peer yet.
//
//nolint:gochecknoglobals // test-only set
var kernelMultiConsumerAllowlist = map[string]bool{
	"kernel/slog":               true, // consumed only by app composition root
	"kernel/logger":             true, // consumed at composition root — no production importers outside kernel
	"kernel/observe":            true, // new foundational package — consumers will be added as call sites migrate
	"kernel/otelsetup":          true, // consumed only by web.Module composition root
	"kernel/upgradablerw_mutex": true, // consolidated into web/ws.ScopeMap + web/ws.PlayerConnections
}

// consumerGroup classifies an import path into its top-level consumer group.
func consumerGroup(importPath string) string {
	suffix := strings.TrimPrefix(importPath, modulePrefix)

	switch {
	case strings.HasPrefix(suffix, "game/"):
		return "game"
	case strings.HasPrefix(suffix, "lobby/"):
		return "lobby"
	case strings.HasPrefix(suffix, "web/"):
		return "web"
	default:
		return "other"
	}
}

func buildKernelSet(kernelPkgs []goPackage) map[string]bool {
	kernelSet := make(map[string]bool)

	for _, kpkg := range kernelPkgs {
		suffix := packageSuffix(kpkg.ImportPath)
		if suffix == "kernel" {
			continue // wiring root
		}

		kernelSet[kpkg.ImportPath] = true
	}

	return kernelSet
}

func countConsumerGroups(
	allPkgs []goPackage,
	kernelSet map[string]bool,
) map[string]map[string]bool {
	consumers := make(map[string]map[string]bool)
	for k := range kernelSet {
		consumers[k] = make(map[string]bool)
	}

	for _, pkg := range allPkgs {
		group := consumerGroup(pkg.ImportPath)
		if group == "other" {
			continue
		}

		for _, imp := range pkg.Imports {
			if kernelSet[imp] {
				consumers[imp][group] = true
			}
		}
	}

	return consumers
}

// Rule 22: every kernel sub-package must be consumed by >=2 distinct groups
// (game/, lobby/, web/) or be in the allowlist.
func TestArch_KernelPackagesHaveMultipleConsumers(
	t *testing.T,
) {
	t.Parallel()

	allPkgs := loadPackages(t, "./internal/...")
	kernelPkgs := loadPackages(t, "./internal/kernel/...")

	kernelSet := buildKernelSet(kernelPkgs)
	consumers := countConsumerGroups(allPkgs, kernelSet)

	for kpkg, groups := range consumers {
		suffix := packageSuffix(kpkg)
		if kernelMultiConsumerAllowlist[suffix] {
			continue
		}

		if len(groups) < 2 {
			groupNames := make([]string, 0, len(groups))
			for g := range groups {
				groupNames = append(groupNames, g)
			}

			t.Errorf("%s consumed by only %d group(s) %v — needs >=2 or add to allowlist",
				kpkg, len(groups), groupNames)
		}
	}
}
