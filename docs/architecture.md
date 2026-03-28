# Architecture

This document describes the architecture of the go-risk-it system — a multiplayer online Risk board game.

## System Overview

```mermaid
graph LR
    subgraph Browser
        FE[Frontend :5173]
    end

    subgraph Docker Compose
        BE[Go Backend :8080]
        Kong[Kong API Gateway :8000]
        GoTrue[GoTrue Auth :9999]
        PG[(PostgreSQL :5432)]
        Jaeger[Jaeger :16686]
    end

    FE -- "/api/* (REST)\n/ws/* (WebSocket)" --> BE
    FE -- "/auth/*" --> Kong
    Kong --> GoTrue
    GoTrue --> PG
    BE --> PG
    BE -- "OTLP" --> Jaeger
```

| Service | Port | Purpose |
|---------|------|---------|
| Frontend (Vite dev server) | 5173 | Svelte 5 SPA |
| Go Backend | 8080 | Game engine, REST API, WebSocket |
| Kong | 8000 | API gateway for auth routes |
| GoTrue | 9999 | Supabase auth (JWT issuance) |
| PostgreSQL | 5432 | Persistent storage (game + lobby schemas) |
| Jaeger | 16686 | Distributed tracing UI |

## Go Package Architecture

The backend follows a strict **4-layer architecture** enforced through Go packages and Uber Fx dependency injection. Each layer only depends on the layers below it.

```mermaid
graph TD
    subgraph Web["Web Layer (internal/web)"]
        direction LR
        rest["rest/\nHTTP handlers"]
        wsm["ws/\nWebSocket manager"]
        ctrl["controller/\nRequest → domain"]
        mw["middleware/\nJWT auth"]
        sig_w["signals/\nWS broadcast"]
    end

    subgraph Logic["Logic Layer (internal/logic)"]
        direction LR
        orch["move/orchestration/\nTransaction + pipeline"]
        move["move/{attack,deploy,...}\nPerform + Walk + Advance"]
        phase["phase/\nPhase management"]
        mission["mission/\nVictory checking"]
        card["card/\nCard logic"]
        board["board/\nRegion adjacency"]
        player["player/\nPlayer state"]
        adv["advancement/\nPhase advancement"]
        creation["creation/\nGame setup"]
    end

    subgraph Data["Data Layer (internal/data)"]
        direction LR
        sqlc["game/\nSQLC queries"]
        ldb["lobby/\nSQLC queries"]
        pool["pool/\nConnection pool"]
        mig["migration/\nSchema migrations"]
    end

    subgraph Config["Config Layer (internal/config)"]
        direction LR
        koanf["Koanf\nYAML + env vars"]
    end

    Web -->|"calls"| Logic
    Logic -->|"queries"| Data
    Data -->|"reads"| Config
    Web -->|"reads"| Config
```

### Uber Fx Dependency Injection

The application uses [Uber Fx](https://github.com/uber-go/fx) for compile-time-safe dependency injection. The module hierarchy mirrors the package layers:

```
app.Module
├── config.Module        # Koanf config loading
├── loggerfx.Module      # Zap logger setup
├── data.Module          # DB pools, SQLC queries, migrations
│   ├── game.Module
│   └── lobby.Module
├── logic.Module         # Business logic services
│   ├── game.Module      # Game engine (move/, phase/, mission/, card/, board/, ...)
│   └── lobby.Module     # Lobby management
├── web.Module           # HTTP/WS servers
│   ├── game.Module      # Game REST + WS + signal handlers
│   ├── lobby.Module     # Lobby REST + WS + signal handlers
│   ├── middleware.Module # JWT auth middleware
│   ├── mux.Module       # Route registration
│   ├── nbio.Module      # nbio HTTP engine
│   ├── otel.Module      # OpenTelemetry setup
│   └── rest.Module      # Health check + route helpers
└── rand.Module          # Random number generation
```

Each package exposes a `var Module = fx.Options(...)` that declares its providers. Fx wires everything together at startup — if a dependency is missing or circular, the app fails to start with a clear error message.

## Move Execution Flow

This sequence shows how a move request (e.g., an attack) flows through the code layers:

```mermaid
sequenceDiagram
    participant Client
    participant Handler as AttackHandler<br/>(web/rest/move)
    participant Controller as MoveController<br/>(web/controller)
    participant Orchestrator as Orchestrator[T]<br/>(logic/move/orchestration)
    participant Service as AttackService<br/>(logic/move/attack)
    participant DB as PostgreSQL
    participant Signal as GameStateChangedSignal
    participant WS as WebSocket Clients

    Client->>Handler: POST /moves/attacks
    Handler->>Handler: Decode request body
    Handler->>Controller: PerformAttackMove()
    Controller->>Orchestrator: OrchestrateMove()

    Note over Orchestrator,DB: RepeatableRead Transaction
    Orchestrator->>DB: Begin TX
    Orchestrator->>Service: ValidateQ() — pre-move checks
    Orchestrator->>Service: PerformQ() — roll dice, update troops
    Service->>DB: UPDATE region SET troops = ...
    Service-->>Orchestrator: MoveResult (casualties, outcome)
    Orchestrator->>DB: LogMoveQ() — record to move_log
    Orchestrator->>Orchestrator: IsMissionAccomplishedQ()
    Orchestrator->>Service: WalkQ() — determine next phase
    Service-->>Orchestrator: CONQUER (if conquered) or ATTACK/REINFORCE
    Orchestrator->>Service: AdvanceQ() — insert new phase record
    Service->>DB: INSERT INTO phase ...
    Orchestrator->>DB: Commit TX

    Orchestrator--)Signal: Emit(GameStateChanged)
    Signal--)WS: Broadcast updated state
    WS--)Client: gameState, boardState, playerState
```

### Service Interface Pattern

Each move type (deploy, attack, conquer, reinforce, cards) implements a generic `Service[T, R]` interface composed of three roles:

| Role | Method | Responsibility |
|------|--------|----------------|
| **Performer** | `PerformQ(ctx, querier, move T) → R` | Execute the move logic |
| **PhaseWalker** | `WalkQ(ctx, querier, voluntary bool) → PhaseType` | Determine next phase |
| **Advancer** | `AdvanceQ(ctx, querier, targetPhase, result R)` | Transition to new phase |

The orchestrator calls these three methods in sequence inside a single database transaction with `RepeatableRead` isolation, ensuring atomic state transitions.

## Board Topology

The game board consists of **42 regions** organized into **6 continents**. Controlling all regions in a continent grants bonus troops each turn.

| Continent | Regions | Bonus Troops |
|-----------|---------|--------------|
| Asia | 12 | 7 |
| North America | 9 | 5 |
| Europe | 7 | 5 |
| Africa | 6 | 3 |
| South America | 4 | 2 |
| Oceania | 4 | 2 |

Region adjacency (including cross-continent links) is defined in [`map.json`](../map.json) and loaded at game creation. The board topology is a static graph — adjacency is used for attack validation, conquer movement, and BFS-based reinforcement pathfinding.

## Authentication Flow

```mermaid
sequenceDiagram
    participant User
    participant Frontend
    participant Kong
    participant GoTrue
    participant Backend

    User->>Frontend: Enter credentials
    Frontend->>Kong: POST /auth/v1/token
    Kong->>GoTrue: Forward request
    GoTrue-->>Kong: JWT + refresh token
    Kong-->>Frontend: JWT + refresh token
    Frontend->>Frontend: Store JWT in memory

    Note over Frontend,Backend: REST calls
    Frontend->>Backend: GET /api/v1/games/summary<br/>Authorization: Bearer {JWT}
    Backend->>Backend: Validate JWT, extract user ID
    Backend-->>Frontend: Response

    Note over Frontend,Backend: WebSocket
    Frontend->>Backend: Upgrade /api/v1/games/{id}/ws<br/>Sec-WebSocket-Protocol: risk-it.websocket.auth.token, {JWT}
    Backend->>Backend: Validate JWT from subprotocol
    Backend-->>Frontend: Connection established
    Backend-->>Frontend: Push game state messages
```

## Game Lifecycle

```mermaid
sequenceDiagram
    participant P1 as Player 1
    participant P2 as Player 2
    participant FE as Frontend
    participant BE as Backend
    participant DB as PostgreSQL

    Note over P1,DB: Lobby Phase
    P1->>BE: POST /lobbies (create)
    P2->>BE: POST /lobbies/{id}/join
    P1->>BE: POST /lobbies/{id}/start
    BE->>DB: Create game, assign regions, missions, cards
    BE-->>P1: Lobby WebSocket: game started
    BE-->>P2: Lobby WebSocket: game started

    Note over P1,DB: Game Phase
    P1->>BE: Connect WebSocket /games/{id}/ws
    P2->>BE: Connect WebSocket /games/{id}/ws
    BE-->>P1: gameState, boardState, playerState, cardState, missionState
    BE-->>P2: gameState, boardState, playerState, cardState, missionState

    loop Turn Loop
        P1->>BE: POST /moves/deployments
        BE-->>P1: Updated boardState, gameState
        BE-->>P2: Updated boardState, gameState
        P1->>BE: POST /moves/attacks
        P1->>BE: POST /moves/conquers
        P1->>BE: POST /moves/reinforcements
        BE->>BE: Check mission completion
        BE->>BE: Advance to next turn
    end

    BE-->>P1: gameState (winner declared)
    BE-->>P2: gameState (winner declared)
```

## Game Phase State Machine

```mermaid
stateDiagram-v2
    [*] --> Cards: Turn starts

    Cards --> Deploy: Play cards or skip
    Note right of Cards: Auto-skipped if < 3 cards<br/>Forced if >= 5 cards

    Deploy --> Attack: All troops placed

    Attack --> Conquer: Successful attack
    Conquer --> Attack: Troops moved to conquered region

    Attack --> Reinforce: Skip or no valid attacks

    Reinforce --> Cards: End turn → next player

    Note right of Reinforce: Dead players are<br/>automatically skipped
```

### Phase Details

| Phase | Action | Endpoint |
|-------|--------|----------|
| **Cards** | Trade card combinations for bonus troops | `POST /moves/cards` |
| **Deploy** | Place troops on owned regions | `POST /moves/deployments` |
| **Attack** | Attack adjacent enemy regions | `POST /moves/attacks` |
| **Conquer** | Move troops into newly conquered region | `POST /moves/conquers` |
| **Reinforce** | Move troops between connected owned regions | `POST /moves/reinforcements` |

## REST API Reference

All endpoints except `/healthz` require a valid JWT in the `Authorization: Bearer {token}` header.

### Health

| Method | Path | Description |
|--------|------|-------------|
| GET | `/healthz` | Health check (no auth) |

### Lobby

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/lobbies` | Create a new lobby |
| GET | `/api/v1/lobbies/summary` | List lobbies for current user |
| POST | `/api/v1/lobbies/{id}/join` | Join a lobby |
| POST | `/api/v1/lobbies/{id}/start` | Start game from lobby |

### Game

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/games` | Create a new game |
| GET | `/api/v1/games/summary` | List games for current user |
| POST | `/api/v1/games/{id}/advancements` | Advance phase (dead player skip) |

### Moves

| Method | Path | Phase | Description |
|--------|------|-------|-------------|
| POST | `/api/v1/games/{id}/moves/cards` | Cards | Trade card combinations |
| POST | `/api/v1/games/{id}/moves/deployments` | Deploy | Place troops on a region |
| POST | `/api/v1/games/{id}/moves/attacks` | Attack | Attack an adjacent region |
| POST | `/api/v1/games/{id}/moves/conquers` | Conquer | Move troops after conquest |
| POST | `/api/v1/games/{id}/moves/reinforcements` | Reinforce | Redistribute troops |

### Test-Only (Component Tests)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/reset` | Reset all database state |
| POST | `/api/v1/setup-near-win` | Setup game near victory for E2E testing |

## WebSocket Protocol

### Connection

- **Game**: `ws://host/api/v1/games/{id}/ws`
- **Lobby**: `ws://host/api/v1/lobbies/{id}/ws`
- **Auth**: JWT passed via WebSocket subprotocol header: `Sec-WebSocket-Protocol: risk-it.websocket.auth.token, {JWT}`

### Message Envelope

All messages follow this format:

```json
{
  "type": "<messageType>",
  "data": { ... }
}
```

### Message Types

| Type | Direction | Description |
|------|-----------|-------------|
| `gameState` | Server → Client | Current turn, phase type, game ID |
| `boardState` | Server → Client | All regions with owner and troop count |
| `playerState` | Server → Client | All players with status, card count, connection |
| `cardState` | Server → Client | Current player's cards (private per player) |
| `missionState` | Server → Client | Current player's secret mission (private) |
| `moveHistory` | Server → Client | Log of all moves in the game |
| `lobbyState` | Server → Client | Lobby participants and readiness |

## Database

Two PostgreSQL schemas, auto-migrated on startup via [golang-migrate](https://github.com/golang-migrate/migrate):

### `lobby` Schema

| Table | Purpose |
|-------|---------|
| `lobby` | Lobby instances (owner, linked game) |
| `participant` | Players in a lobby |

### `game` Schema

| Table | Purpose |
|-------|---------|
| `game` | Game instances (current phase, winner) |
| `player` | Players in a game (turn order, status) |
| `region` | Board regions (owner, troop count) |
| `card` | Risk cards (type, owner) |
| `phase` | Phase records (type, turn number) |
| `deploy_phase` | Deploy-specific state (remaining troops) |
| `conquer_phase` | Conquer-specific state (source/target regions, min troops) |
| `mission` | Player missions |
| `two_continents_mission` | Mission: conquer 2 specific continents |
| `two_continents_plus_one_mission` | Mission: conquer 2 continents + 1 of choice |
| `eliminate_player_mission` | Mission: eliminate a specific player |
| `move_log` | Move history (JSONB data + results) |

### Enums

- **phase_type**: `CARDS`, `DEPLOY`, `ATTACK`, `CONQUER`, `REINFORCE`
- **card_type**: `CAVALRY`, `INFANTRY`, `ARTILLERY`, `JOLLY`
- **mission_type**: `EIGHTEEN_TERRITORIES_TWO_TROOPS`, `TWENTY_FOUR_TERRITORIES`, `TWO_CONTINENTS`, `TWO_CONTINENTS_PLUS_ONE`, `ELIMINATE_PLAYER`

## Observability

The backend emits traces via [OpenTelemetry](https://opentelemetry.io/) to Jaeger:

- **Protocol**: OTLP over HTTP (`http://jaeger:4318`)
- **Service name**: `risk-it`
- **Jaeger UI**: http://localhost:16686

All REST handlers and database operations are instrumented with spans for request tracing.

## Additional Documentation

- **[Component Architecture](architecture-components.md)** — Mermaid flowchart showing subsystem groupings and layer relationships
- **[Auto-Generated Diagram](architecture-diagram.svg)** — D2 diagram generated from the import graph, CI-checked for freshness
- **[Testing Philosophy](testing-philosophy.md)** — Invariant-based testing strategy with 12 game-state properties
- **[doc.go Specification](doc-go-spec.md)** — Living documentation format and layer taxonomy
