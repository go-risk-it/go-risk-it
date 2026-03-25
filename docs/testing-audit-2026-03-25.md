# Testing Debt Audit — 2026-03-25

## Summary

Systematic quality audit of the go-risk-it test suite (391+ tests across 142 packages). The codebase has a healthy testing baseline: setup() helpers, table-driven tests, t.Parallel() everywhere, proper domain error assertions. This audit identified 2 critical findings (resolved), 4 moderate findings (documented for incremental improvement), and 2 minor findings.

**Methodology:** Static analysis of all 41 test files across three dimensions: over-mocking patterns, error path coverage gaps in critical services, and brittle implementation-asserting tests.

## Findings

### Critical (Resolved)

#### C-1: Redundant AssertExpectations — mockery v3 auto-cleanup ignored

**Severity:** Critical | **Status:** Resolved

Mockery v3 mocks created with `NewXxx(t)` register `t.Cleanup` automatically, making explicit `AssertExpectations(t)` calls redundant. 16 redundant calls existed across 6 files:

| File | Calls removed |
|------|---------------|
| `logic/game/player/service_test.go` | 3 |
| `logic/game/creation/service_test.go` | 3 |
| `logic/game/region/service_test.go` | 8 |
| `web/game/controller/player_test.go` | 1 |
| `web/game/controller/board_test.go` | 1 |
| `logic/game/move/deploy/service_test.go` | 1 (including region assignment) |

**Exception:** `data/db/transaction_test.go` correctly retains 6 `AssertExpectations` calls — it uses hand-rolled mocks (`mockTx`, `mockQuerier`) embedding `mock.Mock` directly, which don't get auto-cleanup.

**Resolution:** All redundant calls removed. `.On()` API also migrated to `.EXPECT()` for consistency (9 calls across player, region, board test files).

#### C-2: CI coverage threshold effectively meaningless

**Severity:** Critical | **Status:** Resolved

CI coverage threshold was 7% while actual coverage is 28.3%. The threshold would only trigger on catastrophic test deletion (losing >75% of all tests).

**Resolution:** Threshold raised to 25% in `.github/workflows/go.yml`. This is safely below actual coverage (3.3% headroom) while making the threshold meaningful — it now catches significant regressions.

### Moderate (Documented for Incremental Improvement)

#### M-1: Move orchestration has zero unit tests

**Severity:** Moderate

`internal/logic/game/move/orchestration/service.go` is the central coordination layer: validate -> perform -> log -> check-mission -> walk -> advance -> signal. This 7-step pipeline has **no unit test file**. It is exercised by:
- Integration tests via the invariant test framework (200+ random games)
- Indirectly via move-type-specific service tests (deploy, attack, conquer, reinforce, cards)

**Risk:** Pipeline coordination bugs (wrong step ordering, missing error propagation between steps, incorrect transaction boundaries) are only caught at integration level. Unit tests for the orchestration service would catch these faster and with better failure localization.

**Recommendation:** Add `orchestration/service_test.go` with table-driven tests covering:
- Happy path for each move type (verify all 7 steps execute)
- Error propagation (error in step N prevents steps N+1..7)
- Transaction boundary behavior
- Signal emission on success vs failure

#### M-2: Mission checkers have zero unit tests

**Severity:** Moderate

5 mission checker implementations have no unit tests:
- `eliminate_player.go`
- `eighteen_territories.go`
- `twenty_four_territories.go`
- `two_continents.go`
- `two_continents_plus_one.go`

These are exercised by the invariant test framework (PhaseTransitionLegality invariant), but edge cases in mission completion logic are not directly tested.

**Recommendation:** Add `mission/checker/*_test.go` with table-driven tests for boundary conditions (e.g., exactly 18 territories, exactly 24, continent ownership edge cases).

#### M-3: EqualError exact-string assertions (15 instances)

**Severity:** Moderate

15 `require.EqualError(t, err, "exact string")` assertions across 9 files match error messages verbatim. These break when error message wording changes (even cosmetic rewording).

| File | Count | Example |
|------|-------|---------|
| `advancement/service_test.go` | 1 | `"failed to get game state"` |
| `creation/service_test.go` | 2 | `"failed to create game"` |
| `player/service_test.go` | 2 | `"failed to insert players"` |
| `region/service_test.go` | 3 | `"failed to insert regions"` |
| `deploy/service_test.go` | 2 | `"failed to perform deploy"` |
| `attack/service_test.go` | 1 | `"failed to perform attack"` |
| `conquer/service_test.go` | 1 | `"failed to perform conquer"` |
| `reinforce/service_test.go` | 1 | `"failed to perform reinforce"` |
| `cards/service_test.go` | 2 | `"failed to perform cards"` |

The codebase already uses the better alternatives: `require.ErrorContains` (14 instances, substring match) and `require.ErrorAs` (5 instances, type check).

**Recommendation:** Migrate `EqualError` to `ErrorContains` where the exact string isn't semantically important (most cases). Use `ErrorAs` when the error type matters. Keep `EqualError` only when the exact message is part of the contract.

#### M-4: Mixed mock API styles (.On vs .EXPECT)

**Severity:** Moderate | **Status:** Partially resolved

After the C-1 fix, only `data/db/transaction_test.go` still uses `.On()` — but this is justified because it uses hand-rolled mocks (not mockery-generated). All mockery-generated mock usage now consistently uses `.EXPECT()`.

**Remaining:** `transaction_test.go` uses hand-rolled mocks with `mock.Mock` embedding. Consider migrating to mockery-generated mocks for the `Transactable` interface to enable auto-cleanup and `.EXPECT()` consistency. Low priority — the hand-rolled approach works and the tests are clear.

### Minor (Noted)

#### m-1: Coverage holes in non-critical packages

**Severity:** Minor

Several packages have 0% unit test coverage. Most are integration-tested via the invariant framework or E2E tests:

- `internal/logic/game/signals/` — Signal handlers (GameStateChanged, MovePerformed, PlayerConnected). Exercised via integration tests.
- `internal/config/` — Configuration loading. Exercised via startup.
- `internal/ctx/` — Context types. Exercised via every test that uses GameContext.

**Recommendation:** No action needed. These are appropriately covered by integration-level testing. Adding unit tests would over-mock the surrounding infrastructure.

#### m-2: data/db/transaction_test.go hand-rolled mocks

**Severity:** Minor

`transaction_test.go` defines `mockTx` and `mockQuerier` structs with `mock.Mock` embedding rather than using mockery-generated mocks. This is the only file in the codebase with hand-rolled mocks.

**Recommendation:** Low priority. The hand-rolled approach is appropriate here — the transaction boundary testing has specific needs (testing Begin/Commit/Rollback orchestration) that align well with manual mock control.

## Metrics

| Metric | Value |
|--------|-------|
| Total test files audited | 41 |
| Total tests | 391+ |
| t.Parallel() usage | 277+ calls (universal) |
| Max mock density | 9% (advancement/service_test.go) |
| Over-mocking threshold | 50% |
| Over-mocking instances found | 0 |
| .Times()/.Once()/.Twice() brittleness | 0 instances |
| AssertCalled/AssertNotCalled brittleness | 0 instances |
| CI coverage (actual) | 28.3% |
| CI coverage threshold (before) | 7% |
| CI coverage threshold (after) | 25% |

## Conclusion

The go-risk-it test suite is in good shape. The testing baseline (setup helpers, table-driven tests, t.Parallel, proper domain error handling) is solid and consistent. No over-mocking or call-count brittleness was found.

The two critical findings (redundant AssertExpectations, meaningless CI threshold) have been resolved. The moderate findings (untested orchestration, untested mission checkers, EqualError brittleness) represent genuine improvement opportunities but don't indicate systemic quality problems — they're natural candidates for incremental improvement as those areas are next touched.
