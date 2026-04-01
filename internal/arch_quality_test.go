package internal_test

import (
	"bufio"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// ─── Rules 14–16: Package Quality Metrics ───

const docGoFile = "doc.go"

// Rule 14: no package exceeds the export ceiling (excluding generated packages).
func TestArch_MaxExportsPerPackage(t *testing.T) {
	t.Parallel()

	baseline := loadBaseline(t)
	pkgs := loadPackages(t, "./internal/...")

	for _, pkg := range pkgs {
		if isGeneratedPackage(pkg.ImportPath) || isLoadtestPackage(pkg.ImportPath) {
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
		if isLoadtestPackage(pkg.ImportPath) {
			continue
		}

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
		if isGeneratedPackage(pkg.ImportPath) || isLoadtestPackage(pkg.ImportPath) {
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
			if strings.HasSuffix(goFile, "_test.go") || goFile == docGoFile {
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

// ─── Rule 25: Observe done(nil) Safety ───

// doneNilPattern matches literal `defer done(nil)` statements.
//
//nolint:gochecknoglobals // test-only regex
var doneNilPattern = regexp.MustCompile(`\bdefer\s+done\(nil\)`)

// enclosingFuncReturnsError scans backward from lineIdx to find the nearest
// func declaration and checks if its return type list includes "error".
func enclosingFuncReturnsError(lines []string, lineIdx int) bool {
	for scanIdx := lineIdx - 1; scanIdx >= 0; scanIdx-- {
		line := strings.TrimSpace(lines[scanIdx])

		// Skip blank lines and comments.
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		// Found the enclosing func declaration (may span multiple lines;
		// look for the closing `)` of the return list on this or subsequent lines).
		if strings.HasPrefix(line, "func ") || strings.HasPrefix(line, "func(") {
			// Reconstruct the full signature (may be split across lines).
			sig := line
			var sigSb273 strings.Builder
			for j := scanIdx + 1; j < lineIdx; j++ {
				sigSb273.WriteString(" " + strings.TrimSpace(lines[j]))
				if strings.Contains(lines[j], "{") {
					break
				}
			}
			sig += sigSb273.String()

			return strings.Contains(sig, "error")
		}

		// If we hit a closing brace of a previous function, stop — we're
		// not inside a func.
		if line == "}" {
			return false
		}
	}

	return false
}

// Rule 25: defer done(nil) must only appear in void functions (no error return).
// In error-returning functions, use `defer func() { done(err) }()` or
// SpanFunc/SpanErr instead. This prevents silently dropping error status on spans.
func TestArch_DoneNilOnlyInVoidFunctions(t *testing.T) {
	t.Parallel()

	pkgs := loadPackages(t, "./internal/...")

	for _, pkg := range pkgs {
		for _, goFile := range pkg.GoFiles {
			if strings.HasSuffix(goFile, "_test.go") || goFile == docGoFile {
				continue
			}

			path := filepath.Join(pkg.Dir, goFile)

			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to read %s: %v", path, err)
			}

			// Quick check: skip files without defer done(nil).
			if !doneNilPattern.Match(content) {
				continue
			}

			// Scan line by line to find each occurrence.
			lines := readLines(t, path)
			for idx, line := range lines {
				if !doneNilPattern.MatchString(line) {
					continue
				}

				// Skip comment lines.
				trimmed := strings.TrimSpace(line)
				if strings.HasPrefix(trimmed, "//") {
					continue
				}

				if enclosingFuncReturnsError(lines, idx) {
					t.Errorf(
						"%s:%d has defer done(nil) in an error-returning function "+
							"— use defer func() { done(err) }() or SpanFunc/SpanErr",
						path, idx+1,
					)
				}
			}
		}
	}
}

// readLines reads a file into a slice of lines.
func readLines(t *testing.T, path string) []string {
	t.Helper()

	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("failed to open %s: %v", path, err)
	}
	defer file.Close()

	var lines []string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("failed to scan %s: %v", path, err)
	}

	return lines
}

// ─── Rule 26: Rebaseable Contract on Typed Contexts ───

// ctxPackagesWithRebaseable maps context packages to the concrete types that
// must have compile-time Rebaseable assertions. Each entry is the struct name
// (unexported) that implements the context interface.
//
//nolint:gochecknoglobals // test-only mapping
var ctxPackagesWithRebaseable = map[string][]string{
	"./internal/kernel/ctx": {"userContext"},
	"./internal/game/ctx":   {"gameContext"},
	"./internal/lobby/ctx":  {"lobbyContext"},
}

// Rule 26: every concrete typed-context struct must have a compile-time
// `var _ Rebaseable = (*T)(nil)` assertion. This ensures auto-rebase in
// observe.Span and observe.RawSpan works for all context types.
//

func TestArch_TypedContextsImplementRebaseable(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()

	for pkgPattern, requiredTypes := range ctxPackagesWithRebaseable {
		pkgs := loadPackages(t, pkgPattern)

		for _, pkg := range pkgs {
			found := findRebaseableAssertions(t, fset, pkg)

			for _, requiredType := range requiredTypes {
				if !found[requiredType] {
					t.Errorf(
						"%s: missing compile-time assertion "+
							"var _ Rebaseable = (*%s)(nil)",
						pkg.ImportPath, requiredType,
					)
				}
			}
		}
	}
}

// findRebaseableAssertions scans a package's Go files for
// `var _ ... Rebaseable = (*T)(nil)` assertions and returns a set of the
// type names T found.
func findRebaseableAssertions(
	t *testing.T,
	fset *token.FileSet,
	pkg goPackage,
) map[string]bool {
	t.Helper()

	found := make(map[string]bool)

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
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}

			for _, spec := range genDecl.Specs {
				valueSpec, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}

				// Must be `var _ SomeType = (*T)(nil)`
				if len(valueSpec.Names) != 1 || valueSpec.Names[0].Name != "_" {
					continue
				}

				// Check if the type is Rebaseable (possibly qualified).
				if !isRebaseableType(valueSpec.Type) {
					continue
				}

				// Extract the concrete type from the value `(*T)(nil)`.
				typeName := extractNilPtrType(valueSpec.Values)
				if typeName != "" {
					found[typeName] = true
				}
			}
		}
	}

	return found
}

// isRebaseableType checks if a type expression is "Rebaseable" or "pkg.Rebaseable".
func isRebaseableType(expr ast.Expr) bool {
	switch typed := expr.(type) {
	case *ast.Ident:
		return typed.Name == "Rebaseable"
	case *ast.SelectorExpr:
		return typed.Sel.Name == "Rebaseable"
	default:
		return false
	}
}

// extractNilPtrType extracts T from `(*T)(nil)` in a value spec.
//
//nolint:varnamelen // ok is idiomatic for type assertion checks
func extractNilPtrType(values []ast.Expr) string {
	if len(values) != 1 {
		return ""
	}

	// The pattern is a type conversion: (*T)(nil)
	call, ok := values[0].(*ast.CallExpr)
	if !ok || len(call.Args) != 1 {
		return ""
	}

	// Check arg is nil.
	argIdent, ok := call.Args[0].(*ast.Ident)
	if !ok || argIdent.Name != "nil" {
		return ""
	}

	// The function position is a ParenExpr wrapping *T.
	paren, ok := call.Fun.(*ast.ParenExpr)
	if !ok {
		return ""
	}

	star, ok := paren.X.(*ast.StarExpr)
	if !ok {
		return ""
	}

	ident, ok := star.X.(*ast.Ident)
	if !ok {
		return ""
	}

	return ident.Name
}

// ─── Rules R1, E5, M1: Route Package Enforcement ───

// isRoutesPackage returns true if the package import path ends with "/routes".
func isRoutesPackage(importPath string) bool {
	suffix := packageSuffix(importPath)

	return strings.HasSuffix(suffix, "/routes")
}

// Rule R1: routes/ packages must not import the "errors" standard library package.
// After consolidation, all error creation in routes/ uses domainerrors.New*.
// Importing bare "errors" would allow regression to untyped error handling.
func TestArch_RoutesNeverImportBareErrors(t *testing.T) {
	t.Parallel()

	pkgs := loadPackages(t, "./internal/...")

	for _, pkg := range pkgs {
		if !isRoutesPackage(pkg.ImportPath) {
			continue
		}

		for _, imp := range pkg.Imports {
			if imp == "errors" {
				t.Errorf(
					"%s imports \"errors\" — routes/ packages must use "+
						"domainerrors (kernel/errors) instead of bare errors.New",
					pkg.ImportPath,
				)
			}
		}
	}
}

// Rule O1 extension: routes/ packages must not import "log/slog".
// After consolidation, all logging in routes/ uses observe.Info/observe.Error.
func TestArch_RoutesNeverImportSlog(t *testing.T) {
	t.Parallel()

	pkgs := loadPackages(t, "./internal/...")

	for _, pkg := range pkgs {
		if !isRoutesPackage(pkg.ImportPath) {
			continue
		}

		for _, imp := range pkg.Imports {
			if imp == "log/slog" {
				t.Errorf(
					"%s imports \"log/slog\" — routes/ packages must use "+
						"kernel/observe instead of log/slog",
					pkg.ImportPath,
				)
			}
		}
	}
}

// countProductionFiles counts production .go files in a package,
// excluding _test.go and doc.go files.
func countProductionFiles(pkg goPackage) int {
	count := 0

	for _, goFile := range pkg.GoFiles {
		if strings.HasSuffix(goFile, "_test.go") || goFile == docGoFile {
			continue
		}

		count++
	}

	return count
}

// Rule E5: route file count ratchet — the number of production .go files
// in each routes/ directory must not exceed the baseline ceiling.
// This prevents route file proliferation; new endpoints should be consolidated
// into existing controller files.
func TestArch_RouteFileCountCeiling(t *testing.T) {
	t.Parallel()

	baseline := loadBaseline(t)
	pkgs := loadPackages(t, "./internal/...")

	for _, pkg := range pkgs {
		if !isRoutesPackage(pkg.ImportPath) {
			continue
		}

		fileCount := countProductionFiles(pkg)

		if fileCount > baseline.RouteMaxProductionFilesPerDir {
			t.Errorf(
				"%s has %d production files (ceiling: %d) — "+
					"route file count exceeded ceiling — routes should not proliferate; "+
					"consider consolidating into existing files",
				pkg.ImportPath, fileCount, baseline.RouteMaxProductionFilesPerDir,
			)
		}

		t.Logf("routes %s: %d production files (ceiling: %d, headroom: %d)",
			packageSuffix(pkg.ImportPath), fileCount,
			baseline.RouteMaxProductionFilesPerDir,
			baseline.RouteMaxProductionFilesPerDir-fileCount,
		)
	}
}

// routeModuleDirs lists the module routes/ directories that must conform
// to the standard module shape.
//
//nolint:gochecknoglobals // test-only list
var routeModuleDirs = []string{
	"./internal/game/routes",
	"./internal/lobby/routes",
}

// Rule M1: module shape ratchet — both game/routes/ and lobby/routes/ must
// contain the canonical file set: routes.go, module.go, doc.go, and at least
// one file matching *_controller.go. This enforces the post-consolidation
// module structure and prevents drift to ad-hoc file layouts.
func TestArch_RouteModuleShape(t *testing.T) {
	t.Parallel()

	for _, dir := range routeModuleDirs {
		pkgs := loadPackages(t, dir)
		if len(pkgs) == 0 {
			t.Errorf("%s: no package found", dir)

			continue
		}

		pkg := pkgs[0]
		allFiles := make(map[string]bool)

		for _, f := range pkg.GoFiles {
			allFiles[f] = true
		}

		// Required files: routes.go, module.go, doc.go
		for _, required := range []string{"routes.go", "module.go", docGoFile} {
			if !allFiles[required] {
				t.Errorf(
					"%s missing required file %s — "+
						"module routes/ directory missing required files — "+
						"each module needs routes.go, module.go, doc.go, "+
						"and at least one *_controller.go",
					pkg.ImportPath, required,
				)
			}
		}

		// At least one *_controller.go file
		hasController := false

		for f := range allFiles {
			if strings.HasSuffix(f, "_controller.go") &&
				!strings.HasSuffix(f, "_test.go") {
				hasController = true

				break
			}
		}

		if !hasController {
			t.Errorf(
				"%s has no *_controller.go file — "+
					"module routes/ directory missing required files — "+
					"each module needs routes.go, module.go, doc.go, "+
					"and at least one *_controller.go",
				pkg.ImportPath,
			)
		}
	}
}
