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
- [ ] **1.6** Convert middleware + websocket upgrader callsites
- [ ] **1.7** OTel error metrics

## Phase 2: Architecture Tests

- [x] **2.1** Scaffold `arch_test.go` + Rule 1 (logic/ never imports web/ or api/)
- [x] **2.2** Rules 2-5 (layer + cross-domain separation)
- [ ] **2.3** Rules 6-8 (interface boundary + structural)
- [ ] **2.4a** Per-service interfaces: move services
- [ ] **2.4b** Per-service interfaces: core game services
- [ ] **2.4c** Per-service interfaces: support services
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
