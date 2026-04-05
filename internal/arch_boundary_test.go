package internal_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// ─── Rules 1–5: Layer Boundaries ───

// Rule 1: logic/ must never import web/ or api/.
func TestArch_LogicNeverImportsWebOrAPI(t *testing.T) {
	t.Parallel()

	lobbyPkgs := loadPackages(t, "./internal/lobby/internal/logic/...")
	gamePkgs := loadPackages(t, "./internal/game/internal/logic/...")
	lobbyPkgs = append(lobbyPkgs, gamePkgs...)
	pkgs := lobbyPkgs

	// Logic may import api/ (domain types like MoveResult, PhaseType).
	// Logic must NOT import web/, routes/, consumers/.
	assertNoImports(t, pkgs,
		modulePrefix+"web/",
	)
}

// Rule 2: logic/ must never import net/http.
func TestArch_LogicNeverImportsNetHTTP(t *testing.T) {
	t.Parallel()

	lobbyPkgs := loadPackages(t, "./internal/lobby/internal/logic/...")
	gamePkgs := loadPackages(t, "./internal/game/internal/logic/...")
	lobbyPkgs = append(lobbyPkgs, gamePkgs...)
	pkgs := lobbyPkgs

	assertNoRawImports(t, pkgs, "net/http")
}

// Rule 3: data/ must never import logic/ or web/.
func TestArch_DataNeverImportsLogicOrWeb(t *testing.T) {
	t.Parallel()

	lobbyPkgs := loadPackages(t, "./internal/lobby/internal/data/...")
	gamePkgs := loadPackages(t, "./internal/game/internal/data/...")
	lobbyPkgs = append(lobbyPkgs, gamePkgs...)
	pkgs := lobbyPkgs

	assertNoImports(t, pkgs,
		modulePrefix+"logic/",
		modulePrefix+"game/internal/logic/",
		modulePrefix+"lobby/internal/logic/",
		modulePrefix+"web/",
	)
}

// Rule 4: game/logic/ and lobby/logic/ are mutually isolated.
func TestArch_LogicGameAndLobbyIsolated(t *testing.T) {
	t.Parallel()

	gamePkgs := loadPackages(t, "./internal/game/internal/logic/...")
	assertNoImports(t, gamePkgs,
		modulePrefix+"lobby/internal/logic/",
		modulePrefix+"logic/lobby/",
	)

	lobbyPkgs := loadPackages(t, "./internal/lobby/internal/logic/...")
	assertNoImports(t, lobbyPkgs,
		modulePrefix+"game/internal/logic/",
		modulePrefix+"logic/game/",
	)
}

// Rule 4b: kernel/ must never import game or lobby domain packages.
func TestArch_KernelNeverImportsDomain(t *testing.T) {
	t.Parallel()

	pkgs := loadPackages(t, "./internal/kernel/...")

	for _, pkg := range pkgs {
		for _, imp := range internalImports(pkg) {
			if hasPrefix(imp,
				modulePrefix+"logic/",
				modulePrefix+"web/",
				modulePrefix+"game/",
				modulePrefix+"lobby/",
			) {
				t.Errorf("%s imports forbidden package %s", pkg.ImportPath, imp)
			}
		}
	}
}

// ─── Rules 5–5f: Module Isolation ───

// Rule 5: game/** and lobby/** are isolated except for cross-module event packages.
// Each module may import the other's events/ package (the event contract DTOs).
func TestArch_GameAndLobbyModulesIsolated(t *testing.T) {
	t.Parallel()

	gamePkgs := loadPackages(t, "./internal/game/...")
	for _, pkg := range gamePkgs {
		for _, imp := range internalImports(pkg) {
			// game/ may import lobby/events (cross-module event contract)
			if strings.HasPrefix(imp, modulePrefix+"lobby/events") {
				continue
			}

			if hasPrefix(imp, modulePrefix+"lobby/") {
				t.Errorf("%s imports forbidden package %s", pkg.ImportPath, imp)
			}
		}
	}

	lobbyPkgs := loadPackages(t, "./internal/lobby/...")
	for _, pkg := range lobbyPkgs {
		for _, imp := range internalImports(pkg) {
			// lobby/ may import game/events (cross-module event contract)
			if strings.HasPrefix(imp, modulePrefix+"game/events") {
				continue
			}

			if hasPrefix(imp, modulePrefix+"game/") {
				t.Errorf("%s imports forbidden package %s", pkg.ImportPath, imp)
			}
		}
	}
}

// Rule 5b: web/ must not import game/ or lobby/ module packages.
func TestArch_WebNeverImportsModules(t *testing.T) {
	t.Parallel()

	pkgs := loadPackages(t, "./internal/web/...")

	assertNoImports(t, pkgs,
		modulePrefix+"game/",
		modulePrefix+"lobby/",
	)
}

// Rule 5d: game/data/ and lobby/data/ are mutually isolated.
func TestArch_DataModulesIsolated(t *testing.T) {
	t.Parallel()

	gameDataPkgs := loadPackages(t, "./internal/game/internal/data/...")
	assertNoImports(t, gameDataPkgs,
		modulePrefix+"lobby/internal/data/",
	)

	lobbyDataPkgs := loadPackages(t, "./internal/lobby/internal/data/...")
	assertNoImports(t, lobbyDataPkgs,
		modulePrefix+"game/internal/data/",
	)
}

// Rule 5e: game/ctx/ may only be imported by game/** and testing/** packages.
func TestArch_GameCtxImportIsolation(t *testing.T) {
	t.Parallel()

	pkgs := loadPackages(t, "./internal/...")

	for _, pkg := range pkgs {
		short := strings.TrimPrefix(pkg.ImportPath, modulePrefix)

		// Allow: game, game/**, testing/**
		if short == "game" || strings.HasPrefix(short, "game/") ||
			strings.HasPrefix(short, "testing/") {
			continue
		}

		// Kernel test files (external test packages) are allowed to import game/ctx
		// for testing the Rebaseable/LogEnricher contracts.
		if strings.HasPrefix(short, "kernel/") {
			continue
		}

		assertNoImports(t, []goPackage{pkg}, modulePrefix+"game/ctx")
	}
}

// Rule 5f: lobby/ctx/ may only be imported by lobby/** and testing/** packages.
func TestArch_LobbyCtxImportIsolation(t *testing.T) {
	t.Parallel()

	pkgs := loadPackages(t, "./internal/...")

	for _, pkg := range pkgs {
		short := strings.TrimPrefix(pkg.ImportPath, modulePrefix)

		// Allow: lobby/**, testing/**
		if strings.HasPrefix(short, "lobby/") || strings.HasPrefix(short, "testing/") {
			continue
		}

		// Kernel test files are allowed for testing contracts.
		if strings.HasPrefix(short, "kernel/") {
			continue
		}

		assertNoImports(t, []goPackage{pkg}, modulePrefix+"lobby/ctx")
	}
}

// Rule 5g: game/internal/ may only be imported within the game/ module.
// lobby/internal/ may only be imported within the lobby/ module.
// This enforces Go's internal package convention at the architecture test level.
func TestArch_InternalNeverImportedFromOutsideModule(t *testing.T) {
	t.Parallel()

	pkgs := loadPackages(t, "./internal/...")

	for _, pkg := range pkgs {
		short := strings.TrimPrefix(pkg.ImportPath, modulePrefix)

		for _, imp := range internalImports(pkg) {
			impShort := strings.TrimPrefix(imp, modulePrefix)

			// game/internal/ should only be imported by game/ or game/**
			if strings.HasPrefix(impShort, "game/internal/") &&
				short != "game" && !strings.HasPrefix(short, "game/") {
				t.Errorf("%s imports %s (game/internal/ is private to the game module)",
					pkg.ImportPath, imp)
			}

			// lobby/internal/ should only be imported by lobby/ or lobby/**
			if strings.HasPrefix(impShort, "lobby/internal/") &&
				short != "lobby" && !strings.HasPrefix(short, "lobby/") {
				t.Errorf("%s imports %s (lobby/internal/ is private to the lobby module)",
					pkg.ImportPath, imp)
			}
		}
	}
}

// ─── Rules 6, 24: Interface Contracts ───

// Rule 6: every logic service package defines at least one exported interface.
func TestArch_LogicServicesDefineExportedInterface(t *testing.T) {
	t.Parallel()

	lobbyPkgs := loadPackages(t, "./internal/lobby/internal/logic/...")
	gamePkgs := loadPackages(t, "./internal/game/internal/logic/...")
	lobbyPkgs = append(lobbyPkgs, gamePkgs...)
	pkgs := lobbyPkgs

	for _, pkg := range pkgs {
		if !containsFile(pkg, "service.go") {
			continue
		}

		assertHasExportedInterface(t, pkg)
	}
}

// Rule 24: every game-support service package defines at least one exported interface.
func TestArch_GameSupportServicesDefineExportedInterface(t *testing.T) {
	t.Parallel()

	pkgs := loadPackages(t, "./internal/game/internal/snapshot/...")

	for _, pkg := range pkgs {
		if !containsFile(pkg, "service.go") {
			continue
		}

		assertHasExportedInterface(t, pkg)
	}
}

// ─── Rules 7–13: API, Infra, TestOnly, Stdlib Guards ───

// Rule 7: api/ is DTOs-only — it may only import other api/ packages.
func TestArch_APIOnlyImportsAPI(t *testing.T) {
	t.Parallel()

	lobbyPkgs := loadPackages(t, "./internal/lobby/api/...")
	gamePkgs := loadPackages(t, "./internal/game/api/...")
	lobbyPkgs = append(lobbyPkgs, gamePkgs...)
	pkgs := lobbyPkgs

	// api/ packages may also import typed context packages — port interfaces
	// need GameContext/LobbyContext in their signatures.
	apiAllowedNonAPI := map[string]bool{
		modulePrefix + "game/ctx":  true, // GameContext in port signatures
		modulePrefix + "lobby/ctx": true, // LobbyContext in port signatures
	}

	for _, pkg := range pkgs {
		for _, imp := range internalImports(pkg) {
			if apiAllowedNonAPI[imp] {
				continue
			}

			short := strings.TrimPrefix(imp, modulePrefix)
			if !strings.HasPrefix(short, "api/") &&
				!strings.HasPrefix(short, "game/api") &&
				!strings.HasPrefix(short, "lobby/api") {
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
		"./internal/game/internal/rand/...",
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

	lobbyPkgs := loadPackages(t, "./internal/lobby/internal/data/...")
	gamePkgs := loadPackages(t, "./internal/game/internal/data/...")
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

			gameDataPart := strings.TrimPrefix(imp, modulePrefix+"game/internal/data/")
			if gameDataPart != imp && strings.Contains(gameDataPart, "/db") {
				t.Errorf("%s imports game data querier %s (web must go through logic)",
					pkg.ImportPath, imp)
			}

			lobbyDataPart := strings.TrimPrefix(imp, modulePrefix+"lobby/internal/data/")
			if lobbyDataPart != imp && strings.Contains(lobbyDataPart, "/db") {
				t.Errorf("%s imports lobby data querier %s (web must go through logic)",
					pkg.ImportPath, imp)
			}
		}
	}
}

// Rule 12: no package may import stdlib "log" (use log/slog or internal/slog).
func TestArch_NoStdlibLog(t *testing.T) {
	t.Parallel()

	pkgs := excludeLoadtest(loadPackages(t, "./internal/..."))

	assertNoRawImports(t, pkgs, "log")
}

// Rule 13: no package may import "math/rand" (use math/rand/v2 or internal/rand).
func TestArch_NoOldMathRand(t *testing.T) {
	t.Parallel()

	pkgs := loadPackages(t, "./internal/...")

	assertNoRawImports(t, pkgs, "math/rand")
}

// ─── Rule 23: Game-Support Layer ───

// Rule 23: game/config and game/snapshot must never import game/logic/.
// (game/headlines is excluded — it legitimately imports game/logic/board.)
func TestArch_GameSupportNeverImportsGameLogic(t *testing.T) {
	t.Parallel()

	configPkgs := loadPackages(t, "./internal/game/internal/config/...")
	snapshotPkgs := loadPackages(t, "./internal/game/internal/snapshot/...")
	configPkgs = append(configPkgs, snapshotPkgs...)

	assertNoImports(t, configPkgs, modulePrefix+"game/internal/logic/")
}

// ─── Rules N1–N6: Event System Boundaries ───

// Rule N1: event type packages must never import logic.
// Move result types live in game/api/moves/ (not logic/), so no allowlist is needed.
func TestArch_EventsNeverImportLogic(t *testing.T) {
	t.Parallel()

	pkgs := loadPackages(t, "./internal/game/events/...")
	pkgs = append(pkgs, loadPackages(t, "./internal/lobby/events/...")...)

	assertNoImports(t, pkgs,
		modulePrefix+"logic/",
		modulePrefix+"game/internal/logic/",
		modulePrefix+"lobby/internal/logic/",
	)
}

// Rule N2: event type packages must never import web or consumers.
func TestArch_EventsNeverImportWebOrConsumers(t *testing.T) {
	t.Parallel()

	pkgs := loadPackages(t, "./internal/game/events/...")
	pkgs = append(pkgs, loadPackages(t, "./internal/lobby/events/...")...)

	assertNoImports(t, pkgs,
		modulePrefix+"web/",
		modulePrefix+"game/web/",
		modulePrefix+"lobby/web/",
	)
}

// Rule N3: data packages must never import the event bus.
func TestArch_DataNeverImportsBus(t *testing.T) {
	t.Parallel()

	pkgs := loadPackages(t, "./internal/game/internal/data/...")
	pkgs = append(pkgs, loadPackages(t, "./internal/lobby/internal/data/...")...)

	assertNoImports(t, pkgs, modulePrefix+"kernel/bus")
}

// Rule N4: API DTO packages must never import the event bus.
func TestArch_APINeverImportsBus(t *testing.T) {
	t.Parallel()

	pkgs := loadPackages(t, "./internal/game/api/...")
	pkgs = append(pkgs, loadPackages(t, "./internal/lobby/api/...")...)

	assertNoImports(t, pkgs, modulePrefix+"kernel/bus")
}

// approvedBusImporters is the allowlist of packages permitted to import kernel/bus.
// Every new bus importer must be explicitly approved here.
//
//nolint:gochecknoglobals // test-only allowlist
var approvedBusImporters = map[string]bool{
	// Event type definitions (Subscriber)
	"game/events":  true,
	"lobby/events": true,
	// State transition publishers (Publisher)
	"game/internal/logic":                    true,
	"game/internal/logic/creation":           true,
	"game/internal/logic/move/orchestration": true,
	"lobby/internal/logic/management":        true,
	// Connection lifecycle publishers (Publisher)
	"game/ws":  true,
	"lobby/ws": true,
	// Derived event publisher (Publisher)
	"game/internal/handlers": true,
	// Event consumers/broadcasters (Subscriber)
	"lobby/web": true,
	// Cross-module event publisher (lobby start → game creation)
	"lobby/web/routes": true,
	// Event observer (Subscriber)
	"kernel/logger": true,
	// Game summary recorder (Subscriber)
	"game/internal/logic/metrics": true,
}

// Rule N5: only approved packages may import kernel/bus.
func TestArch_BusImportRatchet(t *testing.T) {
	t.Parallel()

	allPkgs := loadPackages(t, "./internal/...")

	for _, pkg := range allPkgs {
		short := packageSuffix(pkg.ImportPath)
		if strings.HasPrefix(short, "kernel/bus") {
			continue // bus package itself
		}

		for _, imp := range pkg.Imports {
			if strings.HasPrefix(imp, modulePrefix+"kernel/bus") {
				if !approvedBusImporters[short] {
					t.Errorf("%s imports kernel/bus but is not in the approved importers list",
						pkg.ImportPath)
				}
			}
		}
	}
}

// Rule N6: bus.Bus composite type must only be referenced in bus package itself
// and FX wiring roots. Consumers should use bus.Publisher or bus.Subscriber.
// Test files are excluded — test helpers legitimately reference bus.Bus.
func TestArch_BusTypeRestrictedToWiring(t *testing.T) {
	t.Parallel()

	allPkgs := loadPackages(t, "./internal/...")

	// Packages allowed to reference the composite Bus type.
	busTypeAllowed := map[string]bool{
		"kernel/bus": true, // the bus package itself
		// FX wiring roots that pass Bus to module registration
		"game/internal/logic": true, // game.go wiring root passes bus
		"lobby":               true, // lobby.go wiring root
		"game":                true, // game.go wiring root
	}

	fset := token.NewFileSet()

	for _, pkg := range allPkgs {
		short := packageSuffix(pkg.ImportPath)
		if busTypeAllowed[short] {
			continue
		}

		// Check if this package even imports bus.
		importsBus := false

		for _, imp := range pkg.Imports {
			if strings.HasPrefix(imp, modulePrefix+"kernel/bus") {
				importsBus = true

				break
			}
		}

		if !importsBus {
			continue
		}

		// Scan source files for bus.Bus type references.
		for _, goFile := range pkg.GoFiles {
			if strings.HasSuffix(goFile, "_test.go") {
				continue
			}

			path := filepath.Join(pkg.Dir, goFile)

			file, err := parser.ParseFile(fset, path, nil, 0)
			if err != nil {
				t.Fatalf("failed to parse %s: %v", path, err)
			}

			ast.Inspect(file, func(node ast.Node) bool {
				selector, isSelector := node.(*ast.SelectorExpr)
				if !isSelector || selector.Sel.Name != "Bus" {
					return true
				}

				ident, isIdent := selector.X.(*ast.Ident)
				if !isIdent {
					return true
				}

				// Check if the identifier refers to the bus package (any alias).
				if ident.Name == "bus" || ident.Name == "eventbus" {
					t.Errorf(
						"%s:%d references bus.Bus type — use bus.Publisher or bus.Subscriber instead",
						path,
						fset.Position(selector.Pos()).Line,
					)
				}

				return true
			})
		}
	}
}

// ─── Rules O1–O2: Observability Enforcement ───

// Rule O1: business logic packages must not import log/slog.
// Structured logging should flow through kernel/observe spans and slog handlers,
// not through direct slog calls in business logic. Known slog users in web/
// (engine.go, upgrader.go) are outside this scope.
func TestArch_BusinessLogicNoSlog(t *testing.T) {
	t.Parallel()

	pkgs := loadPackages(t, "./internal/game/internal/logic/...")
	pkgs = append(pkgs, loadPackages(t, "./internal/lobby/internal/logic/...")...)
	pkgs = append(pkgs, loadPackages(t, "./internal/game/web/...")...)
	pkgs = append(pkgs, loadPackages(t, "./internal/lobby/web/...")...)

	assertNoRawImports(t, pkgs, "log/slog")
}

// directOTelAllowlist lists business logic packages that are permitted to import
// forbidden OTel packages (otel root or otel/trace) as migration debt or
// infrastructure necessity. Each entry maps to the specific forbidden import
// it permits.
//
//nolint:gochecknoglobals // test-only allowlist
var directOTelAllowlist = map[string]string{
	// Needs otel.Meter() for custom game metrics registration.
	"game/internal/logic/metrics": "go.opentelemetry.io/otel",
}

// Rule O2: business logic packages must not import go.opentelemetry.io/otel (root)
// or go.opentelemetry.io/otel/trace directly — use kernel/observe instead.
// Allowed: go.opentelemetry.io/otel/attribute (needed by observe API callers)
// and go.opentelemetry.io/otel/metric (needed by game/logic/metrics/).
// Note: go list -json Imports only includes production imports, so test files
// (which often import otel/trace/noop) are naturally excluded.
func TestArch_BusinessLogicNoDirectOTel(t *testing.T) {
	t.Parallel()

	pkgs := loadPackages(t, "./internal/game/internal/logic/...")
	pkgs = append(pkgs, loadPackages(t, "./internal/lobby/internal/logic/...")...)
	pkgs = append(pkgs, loadPackages(t, "./internal/game/web/...")...)
	pkgs = append(pkgs, loadPackages(t, "./internal/lobby/web/...")...)

	forbidden := map[string]bool{
		"go.opentelemetry.io/otel":       true,
		"go.opentelemetry.io/otel/trace": true,
	}

	for _, pkg := range pkgs {
		short := packageSuffix(pkg.ImportPath)
		allowedImport := directOTelAllowlist[short]

		for _, imp := range pkg.Imports {
			if !forbidden[imp] {
				continue
			}

			if imp == allowedImport {
				continue
			}

			t.Errorf("%s imports forbidden OTel package %s — use kernel/observe instead",
				pkg.ImportPath, imp)
		}
	}
}

// ─── Rule O3: Observe RawSpan Restriction ───

// rawSpanPattern matches observe.RawSpan( calls in source files.
//
//nolint:gochecknoglobals // test-only regex
var rawSpanPattern = regexp.MustCompile(`observe\.RawSpan\(`)

// Rule O3: business logic packages must use observe.Span or observe.SpanErr
// for span creation. observe.RawSpan is restricted to infrastructure packages
// (kernel/, web/) because it requires manual lifecycle management (done function)
// and risks discarded-context bugs.
func TestArch_BusinessLogicNoRawSpan(t *testing.T) {
	t.Parallel()

	pkgs := loadPackages(t, "./internal/game/internal/logic/...")
	pkgs = append(pkgs, loadPackages(t, "./internal/lobby/internal/logic/...")...)
	pkgs = append(pkgs, loadPackages(t, "./internal/game/web/...")...)
	pkgs = append(pkgs, loadPackages(t, "./internal/lobby/web/...")...)
	pkgs = append(pkgs, loadPackages(t, "./internal/game/internal/snapshot/...")...)
	pkgs = append(pkgs, loadPackages(t, "./internal/game/internal/handlers/...")...)

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

			if rawSpanPattern.Match(content) {
				t.Errorf(
					"%s uses observe.RawSpan — business logic must use "+
						"observe.Span or observe.SpanErr instead",
					path,
				)
			}
		}
	}
}

// ─── Rules S1–S2: Snapshot DTO Purity ───

// Rule S1: game/api/snapshot/ must be pure domain types — no logic, data, or web imports.
func TestArch_GameAPISnapshotIsPureDTO(t *testing.T) {
	t.Parallel()

	pkgs := loadPackages(t, "./internal/game/api/snapshot/...")

	for _, pkg := range pkgs {
		for _, imp := range internalImports(pkg) {
			if hasPrefix(imp,
				modulePrefix+"game/internal/logic/",
				modulePrefix+"lobby/internal/logic/",
				modulePrefix+"game/internal/data/",
				modulePrefix+"lobby/internal/data/",
				modulePrefix+"game/web/",
				modulePrefix+"lobby/web/",
				modulePrefix+"web/",
			) || strings.Contains(imp, "/sqlc") || strings.Contains(imp, "/route") {
				t.Errorf(
					"%s imports forbidden package %s (snapshot DTOs must be pure domain types)",
					pkg.ImportPath,
					imp,
				)
			}
		}
	}
}

// Rule S2: lobby/api/snapshot/ must be pure domain types — no logic, data, or web imports.
func TestArch_LobbyAPISnapshotIsPureDTO(t *testing.T) {
	t.Parallel()

	pkgs := loadPackages(t, "./internal/lobby/api/snapshot/...")

	for _, pkg := range pkgs {
		for _, imp := range internalImports(pkg) {
			if hasPrefix(imp,
				modulePrefix+"game/internal/logic/",
				modulePrefix+"lobby/internal/logic/",
				modulePrefix+"game/internal/data/",
				modulePrefix+"lobby/internal/data/",
				modulePrefix+"game/web/",
				modulePrefix+"lobby/web/",
				modulePrefix+"web/",
			) || strings.Contains(imp, "/sqlc") || strings.Contains(imp, "/route") {
				t.Errorf(
					"%s imports forbidden package %s (snapshot DTOs must be pure domain types)",
					pkg.ImportPath,
					imp,
				)
			}
		}
	}
}

// ─── Rules L1–L2: Loadtest Isolation ───

// loadtestAllowedGameAPIPkgs lists game/api sub-packages that the loadtest
// may import for shared DTO types (snapshot types, wire format constants).
// The loadtest must NOT import game/internal/ or game/web/ — only the api layer.
//
//nolint:gochecknoglobals // test-only allowlist
var loadtestAllowedGameAPIPkgs = map[string]bool{
	modulePrefix + "game/api/snapshot":  true, // shared PlayerView, Phase, Mission DTOs
	modulePrefix + "game/api/messaging": true, // wire format constants (PlayerViewType)
}

// Rule L1: loadtest packages must never import game or lobby domain packages,
// except for game/api/ DTO packages explicitly allowlisted above.
// The loadtest harness operates against the server's HTTP/WS API boundary only.
func TestArch_LoadtestNeverImportsGameOrLobby(t *testing.T) {
	t.Parallel()

	pkgs := loadPackages(t, "./internal/loadtest/...")

	for _, pkg := range pkgs {
		for _, imp := range internalImports(pkg) {
			if loadtestAllowedGameAPIPkgs[imp] {
				continue
			}

			if hasPrefix(imp, modulePrefix+"game/", modulePrefix+"lobby/") {
				t.Errorf("%s imports forbidden package %s", pkg.ImportPath, imp)
			}
		}
	}
}

// Rule L2: game and lobby domain packages must never import loadtest.
func TestArch_GameLobbyNeverImportsLoadtest(t *testing.T) {
	t.Parallel()

	gamePkgs := loadPackages(t, "./internal/game/...")
	lobbyPkgs := loadPackages(t, "./internal/lobby/...")
	gamePkgs = append(gamePkgs, lobbyPkgs...)
	assertNoImports(t, gamePkgs, modulePrefix+"loadtest/")
}
