# Quality Guardrails Roadmap

## Status: All phases complete

## Rules Catalog

| # | Rule | Test | Added |
|---|------|------|-------|
| 1 | logic/ never imports web/ or api/ | TestArch_LogicNeverImportsWebOrAPI | existing |
| 2 | logic/ never imports net/http | TestArch_LogicNeverImportsNetHTTP | existing |
| 3 | data/ never imports logic/ or web/ | TestArch_DataNeverImportsLogicOrWeb | existing |
| 4 | logic/game/ and logic/lobby/ are mutually isolated | TestArch_LogicGameAndLobbyIsolated | existing |
| 5 | web/game/ and web/lobby/ are mutually isolated | TestArch_WebGameAndLobbyIsolated | existing |
| 6 | every logic service package defines exported interface | TestArch_LogicServicesDefineExportedInterface | existing |
| 7 | api/ only imports other api/ packages | TestArch_APIOnlyImportsAPI | Phase 1 |
| 8 | infrastructure packages are leaf packages | TestArch_InfrastructureIsolation | Phase 1 |
| 9 | testonly/ never imported by production code | TestArch_TestOnlyNeverImportedByProduction | Phase 1 |
| 10 | data/ never imports net/http | TestArch_DataNeverImportsNetHTTP | Phase 1 |
| 11 | web/ never imports data/ querier packages | TestArch_WebNeverImportsDataQuerier | Phase 1 |
| 12 | no stdlib "log" imports (use log/slog) | TestArch_NoStdlibLog | Phase 4 |
| 13 | no "math/rand" imports (use math/rand/v2) | TestArch_NoOldMathRand | Phase 4 |
| 14 | max exports per package ceiling | TestArch_MaxExportsPerPackage | Phase 2 |
| 15 | max internal fan-out ceiling | TestArch_MaxFanOut | Phase 2 |
| 16 | max files per package ceiling | TestArch_MaxFilesPerPackage | Phase 2 |

## Metrics Baseline (`internal/arch_baseline.json`)

| Metric | Current max | Ceiling | Headroom |
|--------|-------------|---------|----------|
| Exports/package | 42 (web/game/controller) | 45 | 3 |
| Internal fan-out | 22 (web/game/controller) | 25 | 3 |
| Files/package | 10 (web/game/controller) | 12 | 2 |

Generated code (sqlc, mocks) is excluded from export and file counts.

## CI Checks

| Check | Workflow | Status |
|-------|----------|--------|
| Architecture rules | go.yml `architecture` job | Blocking |
| Dead code detection | deadcode.yml | Advisory (continue-on-error) |
| Coverage threshold | go.yml `build` job | Blocking (7% floor) |
| Security vulns | go.yml `security` job | Advisory |
| Linting | golangci-lint.yml | Blocking |

## Phase Completion

- [x] Phase 1: Layer Boundary Completion (Rules 7-11)
- [x] Phase 2: Package Quality Metrics (Rules 14-16, baseline file)
- [x] Phase 3: Dead Code Cleanup (removed 10 dead functions, advisory CI)
- [x] Phase 4: Import Guards & Coverage (Rules 12-13, threshold 20% → 7%)
- [x] Phase 5: Architecture CI Job & Ratcheting (dedicated check, headroom report)

## Case Law

Rules added from real violations will be documented here:

<!-- When adding a rule from a real violation, add a line like:
| 17 | <description> | <test name> | Case law: agent bypassed Rule N by... |
-->
