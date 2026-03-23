# Server Hardening Roadmap

Spec-first hardening: write the spec (test, interface, arch rule) first, watch it fail, then implement.

PR convention: `hardening(X.Y): description`

## Phase 0: Fix Architectural Violations

- [x] **0.1** Extract domain types from `api/` into logic layer
- [x] **0.2** Break cross-domain import (`web/lobby/` → `web/game/`)

## Phase 1: Error Chain Architecture

- [x] **1.1** Error category contract + enhanced domain errors
- [x] **1.2** Error response envelope with trace ID
- [x] **1.3** Error middleware (core)
- [x] **1.4** Convert move handlers
- [x] **1.5** Convert lobby + game management handlers
- [x] **1.6** Convert middleware + websocket upgrader callsites
- [ ] **1.7** OTel error metrics

## Phase 2: Architecture Tests

- [x] **2.1** Scaffold `arch_test.go` + Rule 1 (logic/ never imports web/ or api/)
- [x] **2.2** Rules 2-5 (layer + cross-domain separation)
- [x] **2.3** Rule 6 (every logic service defines exported interface)
- ~~**2.4a-c**~~ Per-service interfaces — dropped (see ADR-006)
- [ ] **2.5** Circular dependency check + CI

## Phase 3: Domain Invariant Testing

- [ ] **3.1** Test infrastructure + DB bootstrapper
- [ ] **3.2** Invariants 1-5 (structural + phase)
- [ ] **3.3** Move generator: deploy + attack
- [ ] **3.4** Move generator: conquer + reinforce + cards
- [ ] **3.5** Invariants 6-10 + CI job

---

## ADR Log

### ADR-001: Domain Player type in player package

**Date:** 2026-03-23
**Decision:** Define `player.Player` struct with `UserID` and `Name` fields in `internal/logic/game/player/types.go`. The logic layer owns its domain types; API-to-domain mapping happens in controllers. Initially considered placing it in `creation/`, but that would create an import cycle since `creation` imports `player` and `player` would need to import `creation` for the type.
**Rationale:** `internal/logic/` should never import `internal/api/`. The `request.Player` type carries JSON tags and validation that are API concerns, not domain concerns.

### ADR-002: GameCreator interface for cross-domain decoupling

**Date:** 2026-03-23
**Decision:** Define `GameCreator` interface in `internal/web/lobby/controller/` with `CreateGame(ctx, regions, players) (int64, error)` method. `StartControllerImpl` depends on this interface, not the concrete `web/game/controller.GameController`.
**Rationale:** `web/lobby/` and `web/game/` are separate domains. The lobby controller should depend on an abstraction, not reach across domain boundaries.

### ADR-003: DecodeRequest returns domain errors, no HTTP side effects

**Date:** 2026-03-23
**Decision:** `DecodeRequest` no longer writes HTTP error responses. It returns `ValidationError` for malformed/invalid requests. The error middleware handles all response writing. The `malformedRequestError` internal type was removed — all cases map to `ValidationError`.
**Rationale:** The previous design wrote HTTP responses AND returned errors, forcing callers to discard the error (`return nil //nolint:nilerr`). This violated the error middleware pattern and hid decode errors from tracing. The refactor gives decode errors trace IDs and span recording for free. The 413/415 status codes collapse to 400 — acceptable for a game server where descriptive messages matter more than HTTP status granularity.

### ADR-004: UnauthorizedError for authentication failures

**Date:** 2026-03-23
**Decision:** Added `UnauthorizedError` type with `CategoryUnauthorized` mapping to HTTP 401. Auth middleware uses `WrapUnauthorizedError` for JWT failures. Internal context errors use plain `errors.New()` → 500.
**Rationale:** 401 (unauthenticated) vs 403 (unauthorized) is a meaningful distinction. Auth failures are a domain concern — the client needs to know their credentials are invalid, not that they lack permission.

### ADR-005: Middleware http.Error → JSON via WriteError

**Date:** 2026-03-23
**Decision:** Middleware (auth, otel, util) and the WS upgrader now use `restutils.WriteError()` instead of `http.Error()`. The single remaining `http.Error` call is in `writeJSONErrorWithTrace` as a last-resort fallback when JSON marshaling itself fails.
**Rationale:** Consistent JSON error responses across all request paths. Also fixed a double-write bug in `util.go` where `buildContext` AND `buildDomainContext` both called `http.Error` on the same error path. Removed `writer` parameter from `buildContext` since it no longer writes responses.

### ADR-006: Drop per-service querier interfaces (2.4)

**Date:** 2026-03-23
**Decision:** Removed the skipped "no logic service references db.Querier directly" arch test (old Rule 6) and dropped increments 2.4a-c from the roadmap. Services continue to accept `db.Querier` directly.
**Rationale:** The codebase already enforces narrow abstractions at the correct layer: service interfaces (Rule 6, née Rule 7) ensure every logic service defines an exported interface with 2-5 methods. Controllers and orchestrators depend on these narrow service interfaces — they never see `db.Querier`. Narrowing `db.Querier` itself would fight the Q-suffixed transaction composition pattern (`InTransactionWithIsolation[Q Transactable[Q], T any]`), which requires a single concrete querier type flowing through the entire transaction. Go's "small interface" guidance applies to API boundaries, not implementation-internal data access. The sqlc-generated interface is the natural data contract and wrapping it in narrow per-service interfaces would create a maintenance surface with no meaningful benefit.
