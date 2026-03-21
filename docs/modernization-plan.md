# go-risk-it Backend Modernization Plan

## Context

The go-risk-it backend was started on Go ~1.20 and is now on 1.26.1. A thorough code review identified opportunities across 4 areas: Go modernization, duplication reduction, testing gaps, and architecture refinements. Performance improvements are **excluded** — a separate benchmarking effort will measure before/after.

---

## Phase 1: Go Modernization (Code Quality Only)

### 1.1 Run `go fix ./...`
Go 1.26 ships 20+ auto-modernizers. Run once, review diff, verify tests pass.

### 1.2 Convert for loops to range-over-int
Two instances in `internal/upgradablerw_mutex/`:
- `upgradablerw_mutex.go:150`: `for i := 0; i < int(r); i++` → `for range int(r)`
- `upgradablerw_mutex_test.go:48`: `for i := 0; i < 10; i++` → `for range 10`

### 1.3 Adopt `slices`/`maps` where beneficial
Already imported in some files (`graph.go`, `transitions.go`). Key candidate:
- `graph.go:26-34`: Replace manual key collection + sort with `slices.Sorted(maps.Keys(g.Edges))`
- Scan for other opportunities after `go fix` is applied

### 1.4 Evaluate iterator adoption (`iter.Seq[T]`)
Pilot on one function, adopt only where it genuinely simplifies:

| Candidate | File | Benefit |
|-----------|------|---------|
| `GetRegions()` | `board/graph.go:26` | Replace build+sort with `maps.Keys` iterator |
| `getPlayerRegionsWithID` | `cards/performer.go` | Lazy filter iterator |
| `CanReach` usable regions | `board/service.go` | Lazy set from iterator |
| 18-territory counting | `mission/service.go` | Iterator + count |

Decision: start with `GetRegions()` pilot, expand if it reads well.

### 1.5 Adopt `testing/synctest` for concurrent tests
Files with concurrent test patterns:
- `internal/web/game/ws/manager_test.go`
- `internal/web/lobby/ws/manager_test.go`
- `internal/upgradablerw_mutex/upgradablerw_mutex_test.go`

Wrap concurrent test bodies in `synctest.Run()`, replace `sync.WaitGroup` with `synctest.Wait()`.

---

## Phase 2: Reduce Duplication

### 2.1 Generic querier wrapper
**Problem:** `internal/data/game/db/querier.go` and `lobby/db/querier.go` are 95% identical (45 lines each), differing only by import paths.

**Solution:** Extract shared generic wrapper to `internal/data/db/querier.go`. Each domain instantiates it with their sqlc types.

**Files:** Create `internal/data/db/querier.go`, modify `game/db/querier.go` + `lobby/db/querier.go`

### 2.2 Consolidate transaction tests
**Problem:** `game/db/transaction_test.go` and `lobby/db/transaction_test.go` are 99% identical (94 lines each), testing the same `InTransaction` function from `internal/data/db/`.

**Solution:** Move to `internal/data/db/transaction_test.go` with a minimal mock `Transactable`. Delete domain-specific copies.

### 2.3 Parameterized pool factory
**Problem:** `game/pool/pool.go` and `lobby/pool/pool.go` are 75% identical. Lobby adds `SET search_path TO lobby;`.

**Solution:** Create `internal/data/pool/factory.go` with:
```go
func NewConnectionPool(lc fx.Lifecycle, log, cfg, schema string, initSQL ...string) (*pgxpool.Pool, error)
```
Domain pools become thin wrappers.

### 2.4 Shared FetchState utility
**Problem:** `game/fetcher/fetcher.go` and `lobby/fetcher/fetcher.go` have identical `FetchState[T]` functions (22 lines).

**Solution:** Extract to `internal/web/fetcher/fetch.go` using `ctx.LogContext` as the common constraint. Both domains import the shared function.

### 2.5 Extract shared move validation
**Problem:** `attack/performer.go` and `reinforce/performer.go` have exact duplicate `checkDeclaredValues` functions.

**Solution:** Create `internal/logic/game/move/validation/declared.go` with:
```go
type DeclaredValues interface {
    GetTroopsInSource() int64
    GetTroopsInTarget() int64
}
func CheckDeclaredValues(ctx, sourceRegion, targetRegion, move DeclaredValues) error
```
Note: `checkTroops` has domain-specific logic (attack checks defender, reinforce doesn't) — keep separate. Only extract `checkDeclaredValues`.

---

## Phase 3: Testing Gaps

Current: 28% of packages tested (23/81).

### 3.1 Advancement logic tests (CRITICAL)
**File:** `internal/logic/game/advancement/service.go` — zero tests, controls phase transitions.

Test cases: happy path advance, phase mismatch → conflict error, validation/walk/advance failures propagated, signal emission verified, transaction isolation.

**Create:** `internal/logic/game/advancement/service_test.go`

### 3.2 Mission checking tests (2 missing types)
**File:** `internal/logic/game/mission/service_test.go` — exists but missing `EighteenTerritoriesTwoTroops` and `TwentyFourTerritories`.

Test cases: boundary conditions (17 vs 18 regions, 23 vs 24 regions), troop count filtering.

### 3.3 Conquer and reinforce move tests
Attack and deploy have tests; conquer and reinforce don't.

**Create:**
- `internal/logic/game/move/conquer/service_test.go` — troop transfer, validation, region ownership
- `internal/logic/game/move/reinforce/service_test.go` — owned regions, reachability, troops, declared values

### 3.4 Lobby system tests
No tests for any lobby service.

**Create:**
- `internal/logic/lobby/creation/service_test.go`
- `internal/logic/lobby/management/service_test.go`
- `internal/logic/lobby/start/service_test.go`

### 3.5 CI coverage threshold
CI already generates coverage and publishes a badge. Add a threshold gate:
- Start at 30%, raise to 45% after Phase 3 tests are in place.
- Modify `.github/workflows/go.yml`

---

## Phase 4: Architecture Refinements

### 4.1 Remove phantom `R` type parameter from Service interface
**Problem:** `service.Service[T, R]` has `R` result type. 3 of 5 moves return empty `MoveResult{}`. Only attack uses `R` meaningfully.

**Solution:** Remove `R` from the composite interface. The orchestrator handles perform→advance result passing internally. Attack advancer reads conquering state from DB instead of receiving it as param.

**Files:** `service/service.go`, `orchestration/service.go`, `advancement/service.go`, all 5 performer/advancer files.

**Risk:** High scope. Do AFTER Phase 3 tests are in place.

### 4.2 Declarative phase transitions
**Assessment:** Phase transitions are ALREADY declarative in `internal/logic/game/phase/transitions.go` (`ValidTransitions` map + `ValidateTransition` function). Walkers contain game-state queries that cannot be pure data tables.

**Decision: No changes needed.** The current design is already clean.

### 4.3 Mission type registry
**Problem:** `mission/service.go` uses manual `switch` on `GameMissionType` — adding a mission requires modifying the switch.

**Solution:** Registry pattern with `MissionChecker` interface. Each mission type becomes a self-contained checker file.

**Create:** `internal/logic/game/mission/registry.go`, `checker/` directory with one file per type.
**Modify:** `mission/service.go` — delegate to registry.

---

## Implementation Order

| Step | Phase | Task | Effort | Depends On |
|------|-------|------|--------|------------|
| 1 | 1.1 | `go fix ./...` | 15 min | — |
| 2 | 1.2 | Range-over-int loops | 5 min | — |
| 3 | 1.3 | `slices`/`maps` adoption | 30 min | — |
| 4 | 2.1 | Generic querier wrapper | 1 hr | — |
| 5 | 2.2 | Consolidate transaction tests | 30 min | 4 |
| 6 | 2.3 | Parameterized pool factory | 45 min | — |
| 7 | 2.4 | Shared FetchState | 30 min | — |
| 8 | 2.5 | Extract shared move validation | 45 min | — |
| 9 | 3.2 | Mission test gaps | 30 min | — |
| 10 | 3.3 | Conquer/reinforce tests | 2 hr | — |
| 11 | 3.4 | Lobby system tests | 1.5 hr | — |
| 12 | 3.1 | Advancement logic tests | 1.5 hr | — |
| 13 | 3.5 | CI coverage threshold | 15 min | 9-12 |
| 14 | 1.4 | Iterator evaluation (pilot) | 1 hr | 3 |
| 15 | 1.5 | `testing/synctest` adoption | 45 min | — |
| 16 | 4.1 | Remove phantom R type param | 3 hr | 10, 12 |
| 17 | 4.3 | Mission type registry | 2 hr | 9 |

**Total: ~16 hours**

---

## Verification

After each step:
1. `go build ./...` succeeds
2. `go test ./...` passes (all existing + new tests)
3. `golangci-lint run` clean

After Phase 2 (duplication):
4. `make run` — full docker compose stack starts, component tests pass

After Phase 3 (testing):
5. Coverage % reported in CI, above threshold

After Phase 4 (architecture):
6. All mocks regenerated (`make mock`), all tests pass
