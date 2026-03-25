# Route Modernization Completion Plan

> **For Claude:** REQUIRED SUB-SKILL: Use swe:executing-plans to implement this plan task-by-task.

**Goal:** Complete the route modernization vision from the brainstorm — eliminate all legacy patterns, make `route.New()` internal, fix WebSocket auth bug, delete the old `/ws` handler, and decide on controller consolidation.

**Architecture:** The typed route constructors (`Public`, `Authed`, `Game`, `Lobby`, `GameWS`, `LobbyWS`) become the only way to create routes. Middleware uses `Route.Wrap()` instead of `route.New()`. The legacy `/ws?gameID=X` endpoint is replaced by RESTful `/api/v1/games/{id}/ws` paths in both backend and clients.

**Tech Stack:** Go 1.24, uber/fx, nbio, gorilla/websocket, Go stdlib http.ServeMux

---

## Audit Summary (current state)

What's done:
- Typed constructors: `Public`, `Authed`, `Game`, `Lobby`, `GameWS`, `LobbyWS`
- Typed handler signatures: `PlainHandler`, `GameHandler`, `LobbyHandler`
- Per-module routing tables: `game/rest/routes.go`, `lobby/rest/routes.go`
- Old middleware deleted: GameMiddleware, LobbyMiddleware, HandleErrors, WSHeaderConverter
- Old scattered handler packages deleted

What's NOT done (this plan):
1. **BUG**: WebSocket auth broken — `ExtractWSToken` runs after auth middleware
2. `route.New()` still exported, used by middleware/health/testonly
3. No `isWebSocket` flag — OTel hardcodes `Pattern() == "/ws"` for WS detection
4. Health handler bypasses constructors
5. testonly handlers use old-style `http.Handler` + manual error handling
6. Legacy `/ws` handler still registered alongside new RESTful WS routes
7. Frontend + perf-test still use old `/ws?gameID=X` URL pattern
8. Controller consolidation from brainstorm not evaluated

---

## Task 1: Add `isWebSocket` flag to Route

**Why:** OTel middleware currently hardcodes `routeToWrap.Pattern() == "/ws"` to detect WebSocket routes. This only matches the legacy route, not `GET /api/v1/games/{id}/ws`. WebSocket routes need special handling (skip status recording — nbio needs the raw ResponseWriter for upgrade).

**Files:**
- Modify: `internal/web/rest/route/route.go`
- Modify: `internal/web/rest/route/constructors.go`
- Modify: `internal/web/middleware/otel.go`
- Test: `internal/web/rest/route/constructors_test.go`
- Test: `internal/web/middleware/otel_test.go`

### Step 1: Write failing test for `IsWebSocket()`

Add to `internal/web/rest/route/constructors_test.go`:

```go
func TestPublic_IsNotWebSocket(t *testing.T) {
    t.Parallel()
    r := route.Public("GET /status", func(w http.ResponseWriter, req *http.Request) error {
        return nil
    })
    assert.False(t, r.IsWebSocket())
}

func TestGameWS_IsWebSocket(t *testing.T) {
    t.Parallel()
    r := route.GameWS("GET /api/v1/games/{id}/ws", func(w http.ResponseWriter, req *http.Request, _ ctx.GameContext) error {
        return nil
    })
    assert.True(t, r.IsWebSocket())
}

func TestLobbyWS_IsWebSocket(t *testing.T) {
    t.Parallel()
    r := route.LobbyWS("GET /api/v1/lobbies/{id}/ws", func(w http.ResponseWriter, req *http.Request, _ ctx.LobbyContext) error {
        return nil
    })
    assert.True(t, r.IsWebSocket())
}
```

### Step 2: Run tests, confirm failure

```bash
go test ./internal/web/rest/route/ -run "TestPublic_IsNotWebSocket|TestGameWS_IsWebSocket|TestLobbyWS_IsWebSocket" -v
```

Expected: compile error — `IsWebSocket()` not defined.

### Step 3: Add `isWebSocket` field + accessor

In `internal/web/rest/route/route.go`, add field and method:

```go
type Route struct {
    handler      http.Handler
    pattern      string
    requiresAuth bool
    isWebSocket  bool
}

func (r *Route) IsWebSocket() bool {
    return r.isWebSocket
}
```

Update `New()` to accept the field (or set it via constructors — see step 4).

### Step 4: Set `isWebSocket: true` in `GameWS` and `LobbyWS` constructors

In `internal/web/rest/route/constructors.go`, update the two WS constructors to set the flag:

```go
func GameWS(pattern string, handler GameHandler) *Route {
    return &Route{
        pattern:      pattern,
        requiresAuth: true,
        isWebSocket:  true,
        handler:      // ... (existing handler logic)
    }
}

func LobbyWS(pattern string, handler LobbyHandler) *Route {
    return &Route{
        pattern:      pattern,
        requiresAuth: true,
        isWebSocket:  true,
        handler:      // ... (existing handler logic)
    }
}
```

### Step 5: Run tests, confirm pass

```bash
go test ./internal/web/rest/route/ -run "TestPublic_IsNotWebSocket|TestGameWS_IsWebSocket|TestLobbyWS_IsWebSocket" -v
```

### Step 6: Update OTel middleware to use `IsWebSocket()`

In `internal/web/middleware/otel.go`, replace:

```go
// Before:
isWebSocket := routeToWrap.Pattern() == "/ws"

// After:
isWebSocket := routeToWrap.IsWebSocket()
```

Also update the OTel test `internal/web/middleware/otel_test.go` — the WS detection test should use a route created with `GameWS` or mark isWebSocket, not pattern matching.

### Step 7: Run full middleware tests

```bash
go test ./internal/web/middleware/ -v
go test ./internal/web/rest/route/ -v
```

### Step 8: Commit

```bash
git add internal/web/rest/route/route.go internal/web/rest/route/constructors.go \
  internal/web/rest/route/constructors_test.go internal/web/middleware/otel.go \
  internal/web/middleware/otel_test.go
git commit -m "feat(route): add isWebSocket flag, fix OTel WS detection"
```

---

## Task 2: Fix WebSocket auth — move `ExtractWSToken` before auth

**Why:** Critical bug. `GameWS`/`LobbyWS` constructors call `ExtractWSToken` inside the route handler, but auth middleware runs first. Browser clients send JWT via `Sec-WebSocket-Protocol` header (not `Authorization`), so auth middleware sees no token and returns 401. The fix: call `ExtractWSToken` in auth middleware before JWT verification. It's a no-op for non-WS requests (checks for `Sec-WebSocket-Protocol` header).

**Files:**
- Modify: `internal/web/middleware/auth.go`
- Modify: `internal/web/rest/route/constructors.go` (remove ExtractWSToken calls from GameWS/LobbyWS)
- Test: `internal/web/middleware/auth_test.go`
- Test: `internal/web/rest/route/constructors_test.go`

### Step 1: Write failing test — WS subprotocol auth

Add to `internal/web/middleware/auth_test.go`:

```go
func TestAuthMiddleware_WebSocketSubprotocol(t *testing.T) {
    t.Parallel()

    // Create a GameWS route (requires auth, isWebSocket=true)
    inner := route.GameWS("GET /api/v1/games/{id}/ws",
        func(w http.ResponseWriter, r *http.Request, gc ctx.GameContext) error {
            w.WriteHeader(http.StatusOK)
            return nil
        },
    )

    mw := middleware.NewAuthMiddleware(testJwtConfig())
    wrapped := mw.Wrap(inner)

    // Simulate browser: token via Sec-WebSocket-Protocol, NOT Authorization header
    token := createValidJWT(t, testJwtConfig())
    req := httptest.NewRequest("GET", "/api/v1/games/1/ws", nil)
    req.Header.Set("Sec-WebSocket-Protocol", "risk-it.websocket.auth.token, "+token)
    // No Authorization header — this is the browser WS pattern

    rec := httptest.NewRecorder()
    wrapped.ServeHTTP(rec, req)

    // Should NOT be 401 — ExtractWSToken should convert subprotocol → Authorization
    assert.NotEqual(t, http.StatusUnauthorized, rec.Code)
}
```

### Step 2: Run test, confirm failure (401)

```bash
go test ./internal/web/middleware/ -run "TestAuthMiddleware_WebSocketSubprotocol" -v
```

Expected: 401 Unauthorized (subprotocol token not extracted before auth).

### Step 3: Move `ExtractWSToken` into auth middleware

In `internal/web/middleware/auth.go`, call `ExtractWSToken` before `verifyJWT`:

```go
func (m *AuthMiddleware) Wrap(routeToWrap *route.Route) *route.Route {
    if !routeToWrap.RequiresAuth() {
        return routeToWrap
    }

    return route.New(
        routeToWrap.Pattern(),
        routeToWrap.RequiresAuth(),
        http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
            // Extract WS subprotocol token to Authorization header (no-op for non-WS)
            route.ExtractWSToken(request)

            traceContext, ok := request.Context().(ctx.TraceContext)
            // ... rest unchanged
        }))
}
```

### Step 4: Remove `ExtractWSToken` from GameWS/LobbyWS constructors

In `internal/web/rest/route/constructors.go`, remove the `ExtractWSToken(request)` call from both `GameWS` and `LobbyWS`. They can now use `WrapErrors` like the non-WS constructors:

```go
func GameWS(pattern string, handler GameHandler) *Route {
    return &Route{
        pattern:      pattern,
        requiresAuth: true,
        isWebSocket:  true,
        handler: WrapErrors(func(writer http.ResponseWriter, request *http.Request) error {
            gameCtx, err := BuildGameContext(request)
            if err != nil {
                return err
            }
            return handler(writer, request, gameCtx)
        }),
    }
}

func LobbyWS(pattern string, handler LobbyHandler) *Route {
    return &Route{
        pattern:      pattern,
        requiresAuth: true,
        isWebSocket:  true,
        handler: WrapErrors(func(writer http.ResponseWriter, request *http.Request) error {
            lobbyCtx, err := BuildLobbyContext(request)
            if err != nil {
                return err
            }
            return handler(writer, request, lobbyCtx)
        }),
    }
}
```

**Note:** `WrapErrors` is now safe for WS routes because errors (context building, upgrade) happen before the WS connection is established. After upgrade, the handler returns `nil`.

### Step 5: Run all tests

```bash
go test ./internal/web/middleware/ -v
go test ./internal/web/rest/route/ -v
```

### Step 6: Commit

```bash
git add internal/web/middleware/auth.go internal/web/middleware/auth_test.go \
  internal/web/rest/route/constructors.go internal/web/rest/route/constructors_test.go
git commit -m "fix(auth): extract WS subprotocol token before JWT verification"
```

---

## Task 3: Add `Route.Wrap()` + make `route.New()` unexported

**Why:** `route.New()` is the old escape hatch that bypasses typed constructors. Making it unexported ensures all routes go through `Public`/`Authed`/`Game`/`Lobby`/`GameWS`/`LobbyWS`. Middleware needs a replacement: `Route.Wrap()` creates a new route with the same metadata but a different handler.

**Files:**
- Modify: `internal/web/rest/route/route.go`
- Test: `internal/web/rest/route/constructors_test.go` (or new `route_test.go`)

### Step 1: Write failing test for `Route.Wrap()`

Add to `internal/web/rest/route/constructors_test.go` (or create `route_test.go`):

```go
func TestRoute_Wrap_PreservesMetadata(t *testing.T) {
    t.Parallel()

    original := route.Game("POST /api/v1/games/{id}/deploy",
        func(w http.ResponseWriter, r *http.Request, gc ctx.GameContext) error {
            return nil
        })

    called := false
    wrapped := original.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        called = true
        original.ServeHTTP(w, r)
    }))

    assert.Equal(t, original.Pattern(), wrapped.Pattern())
    assert.Equal(t, original.RequiresAuth(), wrapped.RequiresAuth())
    assert.Equal(t, original.IsWebSocket(), wrapped.IsWebSocket())

    // Verify the wrapper runs
    req := httptest.NewRequest("POST", "/api/v1/games/1/deploy", nil)
    rec := httptest.NewRecorder()
    wrapped.ServeHTTP(rec, req)
    assert.True(t, called)
}
```

### Step 2: Implement `Route.Wrap()`

In `internal/web/rest/route/route.go`:

```go
// Wrap creates a new Route with the same metadata but a different handler.
// Used by middleware to decorate routes while preserving routing metadata.
func (r *Route) Wrap(handler http.Handler) *Route {
    return &Route{
        pattern:      r.pattern,
        requiresAuth: r.requiresAuth,
        isWebSocket:  r.isWebSocket,
        handler:      handler,
    }
}
```

### Step 3: Rename `New` → `newRoute` (unexported)

In `internal/web/rest/route/route.go`:

```go
// Before:
func New(pattern string, requiresAuth bool, handler http.Handler) *Route {

// After:
func newRoute(pattern string, requiresAuth bool, handler http.Handler) *Route {
```

This will cause compile errors in all callers — which is exactly what we want. We'll fix them in the next tasks.

### Step 4: Fix internal callers of `New()` within the route package

The constructors in `constructors.go` directly create `&Route{}` structs (not using `New()`), so they're unaffected.

### Step 5: Run route package tests

```bash
go test ./internal/web/rest/route/ -v
```

### Step 6: Commit (compile errors expected in middleware/health/testonly — fixed in next tasks)

```bash
git add internal/web/rest/route/route.go internal/web/rest/route/constructors_test.go
git commit -m "refactor(route): add Wrap method, make New unexported"
```

**Note:** The codebase will not compile at this point. Tasks 4-6 fix all remaining callers.

---

## Task 4: Migrate middleware to `Route.Wrap()`

**Why:** All four middleware currently call `route.New()`. Replace with `Route.Wrap()`.

**Files:**
- Modify: `internal/web/middleware/auth.go`
- Modify: `internal/web/middleware/cors.go`
- Modify: `internal/web/middleware/log.go`
- Modify: `internal/web/middleware/otel.go`
- Test: `internal/web/middleware/auth_test.go`
- Test: `internal/web/middleware/otel_test.go`

### Step 1: Update AuthMiddleware

In `internal/web/middleware/auth.go`, replace `route.New(...)` with `routeToWrap.Wrap(...)`:

```go
// Before:
return route.New(
    routeToWrap.Pattern(),
    routeToWrap.RequiresAuth(),
    http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
        // ...
    }))

// After:
return routeToWrap.Wrap(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
    // ... (body unchanged)
}))
```

### Step 2: Update CorsMiddleware

Same pattern in `internal/web/middleware/cors.go`.

### Step 3: Update LogMiddleware

Same pattern in `internal/web/middleware/log.go`.

### Step 4: Update OTelMiddleware

Same pattern in `internal/web/middleware/otel.go`. Note: OTel still reads `routeToWrap.Pattern()` and `routeToWrap.IsWebSocket()` from the closure — this is correct and unchanged.

### Step 5: Run all middleware tests

```bash
go test ./internal/web/middleware/ -v
```

### Step 6: Commit

```bash
git add internal/web/middleware/auth.go internal/web/middleware/cors.go \
  internal/web/middleware/log.go internal/web/middleware/otel.go
git commit -m "refactor(middleware): use Route.Wrap() instead of route.New()"
```

---

## Task 5: Convert health handler to `route.Public()`

**Why:** Health check currently uses `route.New("/status", false, healthCheck.Handler())`, bypassing typed constructors. Convert to `route.Public()`.

**Files:**
- Modify: `internal/web/rest/health/handler.go`

### Step 1: Convert to `route.Public()`

`healthCheck.Handler()` returns an `http.Handler`. `route.Public()` takes a `PlainHandler` (returns error). Wrap it:

```go
// Before:
return route.New("/status", false, healthCheck.Handler()), nil

// After:
return route.Public("GET /status", func(w http.ResponseWriter, r *http.Request) error {
    healthCheck.Handler().ServeHTTP(w, r)
    return nil
}), nil
```

**Important:** Add `GET` method prefix — Go 1.22+ ServeMux requires explicit method for method-based routing. The old pattern `/status` matched all methods. `GET /status` is more correct for health checks.

### Step 2: Update health module wiring

Check `internal/web/rest/health/health.go` module setup — the `fx.ResultTags` annotation should stay the same since the return type (`*route.Route`) is unchanged.

### Step 3: Run tests

```bash
go test ./internal/web/rest/... -v
```

### Step 4: Commit

```bash
git add internal/web/rest/health/handler.go
git commit -m "refactor(health): use route.Public() constructor"
```

---

## Task 6: Convert testonly handlers to typed constructors

**Why:** `testonly/reset_handler.go` and `testonly/setup_near_win_handler.go` use `route.New()` with old-style `http.Handler` interface and manual error handling.

**Files:**
- Modify: `internal/testonly/reset_handler.go`
- Modify: `internal/testonly/setup_near_win_handler.go`

### Step 1: Convert `NewResetHandler`

```go
// Before:
func NewResetHandler(testOnlyController Controller) *route.Route {
    h := &resetHandler{testOnlyController: testOnlyController}
    return route.New("/api/v1/reset", true, h)
}
type resetHandler struct { testOnlyController Controller }
func (h *resetHandler) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
    err := h.testOnlyController.ResetState(req.Context())
    if err != nil {
        http.Error(writer, err.Error(), http.StatusInternalServerError)
        return
    }
    restutils.WriteResponse(writer, []byte{}, http.StatusNoContent)
}

// After:
func NewResetHandler(testOnlyController Controller) *route.Route {
    return route.Authed("POST /api/v1/reset", func(w http.ResponseWriter, r *http.Request) error {
        if err := testOnlyController.ResetState(r.Context()); err != nil {
            return err
        }
        restutils.WriteResponse(w, []byte{}, http.StatusNoContent)
        return nil
    })
}
```

Delete the `resetHandler` struct — it's no longer needed.

### Step 2: Convert `NewSetupNearWinHandler`

```go
// Before: struct-based handler with manual error handling

// After:
func NewSetupNearWinHandler(testOnlyController Controller) *route.Route {
    return route.Authed("POST /api/v1/setup-near-win", func(w http.ResponseWriter, r *http.Request) error {
        body, err := restutils.DecodeRequest[SetupNearWinRequest](w, r)
        if err != nil {
            return err
        }
        if err := testOnlyController.SetupNearWin(r.Context(), body.GameID); err != nil {
            return err
        }
        restutils.WriteResponse(w, []byte{}, http.StatusNoContent)
        return nil
    })
}
```

Delete the `setupNearWinHandler` struct.

### Step 3: Verify compilation + run component tests

```bash
go build ./...
go test ./internal/testonly/ -v
```

### Step 4: Commit

```bash
git add internal/testonly/reset_handler.go internal/testonly/setup_near_win_handler.go
git commit -m "refactor(testonly): use route.Authed() typed constructors"
```

---

## Task 7: Delete legacy `/ws` handler

**Why:** The old `/ws?gameID=X&lobbyID=Y` endpoint is fully replaced by `GET /api/v1/games/{id}/ws` and `GET /api/v1/lobbies/{id}/ws`. The old handler uses manual context extraction, query params, and bypasses typed constructors.

**Files:**
- Delete: `internal/web/rest/websocket_upgrader.go`
- Modify: `internal/web/rest/rest.go` (remove NewWebSocketHandler + wiring)
- Modify: `internal/web/rest/routes.go` (remove wsRoute parameter)

### Step 1: Delete the file

Remove `internal/web/rest/websocket_upgrader.go` entirely.

### Step 2: Update `rest.go` module

In `internal/web/rest/rest.go`, remove the `NewWebSocketHandler` provider and its named parameter:

```go
// Before:
var Module = fx.Options(
    health.Module,
    fx.Provide(
        fx.Annotate(NewWebSocketHandler, fx.ResultTags(`name:"wsRoute"`)),
        fx.Annotate(
            ProvideRoutes,
            fx.ParamTags(`name:"healthRoute"`, `name:"wsRoute"`),
            fx.ResultTags(`group:"routes,flatten"`),
        ),
    ),
)

// After:
var Module = fx.Options(
    health.Module,
    fx.Provide(
        fx.Annotate(
            ProvideRoutes,
            fx.ParamTags(`name:"healthRoute"`),
            fx.ResultTags(`group:"routes,flatten"`),
        ),
    ),
)
```

### Step 3: Update `routes.go`

In `internal/web/rest/routes.go`, remove wsRoute parameter:

```go
// Before:
func ProvideRoutes(healthRoute *route.Route, wsRoute *route.Route) []*route.Route {
    return []*route.Route{healthRoute, wsRoute}
}

// After:
func ProvideRoutes(healthRoute *route.Route) []*route.Route {
    return []*route.Route{healthRoute}
}
```

### Step 4: Verify compilation

```bash
go build ./...
```

### Step 5: Commit

```bash
git add -u internal/web/rest/
git commit -m "refactor(ws): delete legacy /ws handler, use RESTful WS routes only"
```

---

## Task 8: Update perf-test WebSocket client

**Why:** Perf-test still connects to `/ws?gameID=X`. Update to `/api/v1/games/{id}/ws`.

**Files:**
- Modify: `perf-test/internal/client/ws.go`

### Step 1: Update URL construction

In `perf-test/internal/client/ws.go`, line 52:

```go
// Before:
wsURL := fmt.Sprintf("%s/ws?gameID=%d", baseURL, gameID)

// After:
wsURL := fmt.Sprintf("%s/api/v1/games/%d/ws", baseURL, gameID)
```

Auth stays unchanged — perf-test uses `Authorization: Bearer` header directly (handled by auth middleware).

### Step 2: Update any test fixtures

Check `perf-test/internal/runner/protocol_test.go:132` — the test `wsURL` value should also update:

```go
// Before:
wsURL: "ws://localhost",

// After (verify this is just a prefix — the test may construct the full URL):
wsURL: "ws://localhost",
```

If the test constructs the full URL including `/ws?gameID=`, update accordingly.

### Step 3: Run perf-test compilation

```bash
cd perf-test && go build ./... && cd ..
```

### Step 4: Commit

```bash
git add perf-test/internal/client/ws.go
git commit -m "refactor(perf-test): update WS URL to RESTful path"
```

---

## Task 9: Update frontend WebSocket URLs

**Why:** Frontend still connects to `ws://host/ws?gameID=X`. Update to `ws://host/api/v1/games/{id}/ws` and `ws://host/api/v1/lobbies/{id}/ws`.

**Repo:** `go-risk-it-frontend` (separate repo, branch: `svelte-rewrite`)

**Files:**
- Modify: `src/lib/state/websocket.svelte.ts` (game WS URL)
- Modify: `src/lib/state/lobby-state.svelte.ts` (lobby WS URL)
- Possibly modify: `.env` / env config (remove `PUBLIC_WS_URL` or change its meaning)

### Step 1: Update game WS URL

In `src/lib/state/websocket.svelte.ts`:

```typescript
// Before:
url: `${baseUrl}?gameID=${gameId}`
// where baseUrl = PUBLIC_WS_URL = "ws://localhost:8080/ws"

// After:
url: `${baseWsUrl}/api/v1/games/${gameId}/ws`
// where baseWsUrl = "ws://localhost:8080" (no /ws suffix)
```

### Step 2: Update lobby WS URL

In `src/lib/state/lobby-state.svelte.ts`:

```typescript
// Before:
url: `${baseUrl}?lobbyID=${lobbyId}`

// After:
url: `${baseWsUrl}/api/v1/lobbies/${lobbyId}/ws`
```

### Step 3: Update env config

The `PUBLIC_WS_URL` should point to the base (no path), or derive from `PUBLIC_API_URL`:

```env
# Before:
PUBLIC_WS_URL=ws://localhost:8080/ws

# After (option A — separate var):
PUBLIC_WS_URL=ws://localhost:8080

# After (option B — derive from API URL):
# Remove PUBLIC_WS_URL entirely, compute ws:// from PUBLIC_API_URL in code
```

Auth token passing via `Sec-WebSocket-Protocol` subprotocol is unchanged.

### Step 4: Run frontend tests

```bash
cd go-risk-it-frontend
npm run check
npm run test:unit
```

### Step 5: E2E tests (requires both repos updated)

```bash
npm run e2e:up
npm run test:e2e
```

### Step 6: Commit (in frontend repo)

```bash
git add -u src/lib/state/
git commit -m "refactor(ws): update WebSocket URLs to RESTful paths"
```

---

## Task 10: Controller consolidation — decision point

**Investigation findings:**

### Game controllers (9 total, 30 methods)

**REST-facing (called from routes.go):**
| Controller | Methods | Dependencies |
|-----------|---------|-------------|
| GameController | CreateGame, GetUserGames | 3 services |
| AdvancementController | Advance | 3 advancers |
| MoveController | 5x PerformXMove | 5 orchestrators |

**WS-facing only (called from fetchers, not routes):**
| Controller | Methods | Dependencies |
|-----------|---------|-------------|
| BoardController | GetBoardState | 1 service |
| PhaseController | 5x GetXPhaseState | 2 services |
| PlayerController | GetPlayerState | 2 services |
| CardController | GetCardState | 1 service |
| MoveLogController | GetMoveLogs, ConvertMoveLogs | 1 service |
| MissionController | 5x GetXMission | 1 service |

### Lobby controllers (4 total, 6 methods)

| Controller | Methods | Dependencies |
|-----------|---------|-------------|
| CreationController | CreateLobby | 1 service |
| ManagementController | JoinLobby, GetUserLobbies | 1 service |
| StartController | StartGame | 2 deps (GameCreator + service) |
| StateController | GetLobbyState | 1 service |

### Recommendation: **Don't consolidate**

The brainstorm envisioned "Single controller per module." After investigation, this is **counterproductive**:

1. **A single GameController would have 18 dependencies** — a classic god struct
2. **REST controllers and WS-fetcher controllers serve different callers** — they don't belong in the same struct
3. **Each controller has a clear, single responsibility** — AdvancementController advances, MoveController performs moves
4. **The routes.go files already provide the "single readable API surface"** the brainstorm wanted
5. **Lobby controllers are already minimal** (1-2 methods each) — consolidation saves ~50 LOC but loses clarity

The brainstorm's vision was about readability and discoverability. That's already achieved by per-module `routes.go` files + controller-per-domain pattern. The struct consolidation was a means to that end, not the end itself.

**If you still want to consolidate:** The most defensible split would be:
- `GameActionController` (REST-facing: create, advance, moves) — 8 methods, 11 deps
- `GameStateController` (WS-facing: board, phase, player, card, moveLog, mission) — 16 methods, 8 deps
- Leave lobby as-is

This would be Task 10b — a separate, optional follow-up.

---

## Task 11: Clean up `route.New()` usage in `route.go`

**Why:** After all callers are migrated, verify `route.New()` (now `newRoute()`) is only called internally or can be removed entirely.

**Files:**
- Modify: `internal/web/rest/route/route.go`

### Step 1: Check if `newRoute` is still needed

After tasks 1-7, the only remaining caller should be inside `route.go` itself (if constructors use it). If constructors directly create `&Route{}` structs, `newRoute` can be deleted.

### Step 2: Remove if unused

```bash
grep -r "newRoute" internal/web/rest/route/
```

If no callers, delete the function.

### Step 3: Also remove `AsRoute` if unused

`AsRoute` wraps fx annotations. Check if it's still needed after testonly migration:

```bash
grep -r "AsRoute" internal/
```

If testonly still uses it, keep it. Otherwise, delete.

### Step 4: Run full test suite

```bash
go test ./... -count=1
```

### Step 5: Commit

```bash
git add internal/web/rest/route/route.go
git commit -m "chore(route): remove unused newRoute and AsRoute helpers"
```

---

## Execution order and dependencies

```
Task 1 (isWebSocket flag)
  ↓
Task 2 (fix WS auth bug) — depends on isWebSocket from Task 1
  ↓
Task 3 (Route.Wrap + unexport New) — breaks compilation
  ↓
Task 4 (middleware migration) ─┐
Task 5 (health migration)     ├── all fix compilation, can be parallel
Task 6 (testonly migration)   ─┘
  ↓
Task 7 (delete legacy /ws)
  ↓
Task 8 (perf-test URL update)  ─┐── can be parallel
Task 9 (frontend URL update)   ─┘
  ↓
Task 10 (controller decision — recommended: skip)
  ↓
Task 11 (cleanup unused helpers)
```

**Minimum viable commit:** Tasks 1-6 complete the internal cleanup. Tasks 7-9 are the breaking change (delete old `/ws`). Task 11 is polish.
