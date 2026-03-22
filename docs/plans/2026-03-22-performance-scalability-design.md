# Performance & Scalability Design

## Overview & Goals

**Goal:** Scale go-risk-it to thousands of simultaneous games with full observability,
running entirely locally. The strategy is "find the ceiling" — instrument everything,
systematically increase load until something breaks, then fix what the data shows.

**Approach:** Observability as the enabler, not an end in itself. Full OTel instrumentation
(traces, metrics, logs) with custom game-level, WebSocket, and database metrics. A single
`grafana/otel-lgtm` container provides Prometheus + Tempo + Loki + Grafana. Four dashboards
(server golden signals, game engine, WebSocket, database) make bottlenecks visible. Load
testing uses a combined strategy: binary search to find the ballpark, continuous ramp to find
the exact breaking point, soak test to verify stability. Optimization is evidence-based — no
pre-planned fixes, let the dashboards decide.

**Non-goals:** Horizontal scaling (multi-instance, sharding) is explicitly deferred until the
single-instance ceiling is proven. No third-party SaaS dependencies. No premature optimization.

---

## Observability Architecture

### Local Observability Stack

A single `grafana/otel-lgtm` container bundles Prometheus, Tempo, Loki, Grafana, and an OTel
Collector. The Go backend sends all telemetry via OTLP HTTP to port 4318. If the all-in-one
image becomes a bottleneck at high load, it can be split into individual containers with zero
changes to application code.

```
┌─────────────┐     OTLP/HTTP      ┌──────────────────────────┐
│  risk-it    │ ──────────────────► │  grafana/otel-lgtm       │
│  (Go app)   │     :4318           │                          │
└─────────────┘                     │  ┌─ OTel Collector       │
                                    │  ├─ Prometheus (metrics)  │
┌─────────────┐     OTLP/HTTP      │  ├─ Tempo (traces)        │
│  perf-test  │ ──────────────────► │  ├─ Loki (logs)           │
│  (harness)  │     :4318           │  └─ Grafana (dashboards)  │
└─────────────┘                     │         :3000             │
                                    └──────────────────────────┘
```

### Backend Instrumentation (4 layers)

| Layer | What we instrument | Metrics exposed |
|-------|-------------------|-----------------|
| **HTTP/REST** | Request handling via OTel middleware | `http_request_duration_seconds`, `http_requests_total`, `http_request_size_bytes` |
| **Game engine** | Custom counters/gauges in logic layer | `game_active_count`, `game_moves_total{phase}`, `game_phase_duration_seconds{phase}`, `game_completed_total`, `game_timed_out_total` |
| **WebSocket** | Instrumentation in `ws/` package | `ws_connections_active`, `ws_broadcast_duration_seconds`, `ws_messages_sent_total`, `ws_connection_lifetime_seconds` |
| **Database** | pgx tracing hooks + pool stats | `db_query_duration_seconds{query}`, `db_transaction_duration_seconds`, `db_pool_active_connections`, `db_pool_idle_connections`, `db_pool_wait_duration_seconds` |

Additionally, Go runtime metrics (goroutines, GC, heap) are exported automatically via the
OTel SDK's runtime instrumentation.

---

## Grafana Dashboards

### Dashboard 1: Server Golden Signals

The first-look dashboard when something is wrong. RED metrics (Rate, Errors, Duration) for
all HTTP endpoints, plus Go runtime health.

- Request rate by endpoint (req/s)
- Error rate by endpoint and status code (%)
- Latency percentiles (p50, p95, p99) by endpoint
- Goroutine count over time
- Heap allocation rate and total heap size
- GC pause duration and frequency
- Open file descriptors

### Dashboard 2: Game Engine

Application-level view — how the game system behaves under load. No pre-built dashboard
exists for this; it's custom to go-risk-it.

- Active games gauge
- Moves/sec total and broken down by phase (deploy, attack, conquer, reinforce, cards)
- Phase duration heatmap (which phases take longest)
- Game completion rate vs timeout rate
- Average game duration over time
- Moves per game distribution

### Dashboard 3: WebSocket

Real-time delivery health — the critical path between "move executed" and "all players see
the update."

- Active WebSocket connections gauge
- Broadcast latency per game (time to send state to all players)
- Messages sent/sec
- Connection churn (opens and closes per second)
- Failed broadcasts / closed connection errors

### Dashboard 4: Database

The layer most likely to be the first bottleneck. Every move is a RepeatableRead transaction
touching multiple tables.

- Connection pool utilization (used / idle / max)
- Query latency by query name (p50, p95, p99)
- Transaction duration distribution
- Queries/sec
- Pool wait time (time blocked acquiring a connection)
- Slowest queries ranking

---

## Load Testing Strategy

Combined approach: Binary Search → Continuous Ramp → Soak Test. Each phase answers a
different question. All phases use the existing perf-test harness with Grafana dashboards
open.

### Phase A: Binary Search (find the ballpark)

Start at a known-good baseline and increase games at each step. Observe dashboards for 2-3
minutes at each plateau before moving up.

| Step | Games | Players | Purpose |
|------|-------|---------|---------|
| 1 | 1 | 4 | Verify stack works |
| 2 | 10 | 40 | Baseline metrics |
| 3 | 25 | 100 | Early contention signals |
| 4 | 50 | 200 | Current "heavy" preset |
| 5 | 100 | 400 | Moderate scale |
| 6 | 250 | 1,000 | Approaching limits |
| 7 | 500 | 2,000 | Likely stress zone |
| 8 | 1,000 | 4,000 | Pushing ceiling |
| ... | 2x | ... | Until failure |

At each step, record: CPU/memory utilization, p99 latency, error rate, DB pool usage, WS
connection count, goroutine count. Stop when error rate exceeds 5% or p99 latency degrades
>10x from baseline.

### Phase B: Continuous Ramp (find the exact breaking point)

Once the ballpark is known (e.g., "between 250 and 500 games"), ramp linearly through that
range. Requires a new harness feature: continuous game spawning at a configurable rate (e.g.,
+10 games/minute) while existing games continue playing.

The inflection point — where latency curves upward or error rate starts climbing — is the
ceiling.

### Phase C: Soak Test (verify stability)

Run at 80% of the discovered ceiling for 2+ hours. Watch for:

- Memory leaks (heap growing without bound)
- Goroutine leaks (count climbing over time)
- Connection pool drift (idle connections not returning)
- Latency creep (p99 slowly increasing)
- Game completion rate degradation

A clean soak test means the ceiling is real. A degrading soak test means there's a slow-burn
issue that only manifests over time.

---

## Candidate Bottlenecks & Optimization Catalog

We don't pre-commit to an optimization order. The dashboards will reveal what breaks first.
However, we document the known candidate bottlenecks and the optimizations available for each,
so when the data points to a layer we can act quickly.

### Database Layer

- **Connection pool exhaustion** → increase pool size, tune max connections
- **Transaction contention** (RepeatableRead conflicts) → reduce transaction scope, separate
  read and write paths
- **Slow queries under load** → add missing indexes, optimize hot queries, analyze
  `pg_stat_statements`
- **move_log JSONB growth** → partition by game, limit history depth, async writes

### WebSocket Layer

- **Broadcast fan-out latency** → pre-serialize shared messages (one marshal per game, not
  per player)
- **Full state on every move** → delta updates (send only what changed)
- **Message volume** → batch multiple state updates into a single WebSocket frame
- **Connection overhead** → tune nbio buffer sizes, connection-level compression

### Go Runtime

- **GC pressure from allocations** → `sync.Pool` for hot-path objects (JSON buffers, state
  snapshots)
- **Goroutine explosion** → bounded worker pools instead of goroutine-per-request for
  broadcasts
- **CPU saturation in game logic** → profile with `pprof`, optimize mission checking / board
  traversal

### Player Strategy / Game Behavior

- **Suboptimal moves prolong games** → games take more turns than typical human play,
  amplifying DB writes, WS broadcasts, and phase transitions per game
- **Bad attack decisions** (losing battles) → more attack/conquer cycles per turn, more dice
  rolls, more state updates
- **Strategy computation time** → `think-time` flag exists but the heuristic itself may be
  CPU-heavy at scale with board traversal
- **Mitigation:** track moves-per-game and average game duration on the Game Engine dashboard.
  If games are unrealistically long, improve the strategy or add a move cap. Compare "strategy
  overhead" (time in strategy code) vs "server overhead" (time in transaction + broadcast) to
  separate the two

This is an interesting bottleneck because it's **load-profile-shaped, not server-shaped** —
the fix is a smarter test client, not a faster server. The dashboards should make it obvious:
if phase duration is low but games take 200+ turns, the strategy is the problem, not the
server.

### Application Architecture (last resort)

- **Single-instance limits reached** → game state sharding across instances (sticky sessions
  by game ID)
- **Database as bottleneck even after optimization** → read replicas, event sourcing,
  in-memory game state with periodic DB sync
- **WebSocket connection limits** → dedicated WebSocket gateway service

Each optimization should be validated by re-running the load test and comparing dashboards
before and after. No optimization is "done" without quantitative evidence of improvement.

---

## Phased Implementation Roadmap

### Phase 1: Observability (estimated 1-2 weeks)

Independently useful — even without load testing, having dashboards makes development and
debugging better.

1. Add `grafana/otel-lgtm` container to `docker-compose.yml`
   - Ports: 3000 (Grafana), 4317 (OTLP gRPC), 4318 (OTLP HTTP)
   - Volume for data persistence and dashboard provisioning
2. Enable existing OTel tracing in the backend
   - Uncomment metrics exporter in `internal/web/otel/otel.go`
   - Point OTLP endpoint at the LGTM container
   - Add Go runtime metrics instrumentation
3. Add custom game engine metrics
   - Active games gauge, moves counter by phase, phase duration histogram
   - Game completion/timeout counters
4. Add WebSocket metrics
   - Connection gauge, broadcast duration histogram, messages counter
5. Add database metrics
   - pgx tracing hooks for query/transaction duration
   - Connection pool stats export
6. Build 4 Grafana dashboards (provisioned as JSON files in repo)
   - Server golden signals, game engine, WebSocket, database
7. Verify with smoke test (1 game) — confirm all dashboards populate

### Phase 2: Systematic Load Testing (estimated 1 week)

Depends on Phase 1 — dashboards must be working before load testing has value.

1. Run binary search sequence (1 → 10 → 25 → 50 → 100 → 250 → 500 → 1000 games)
   - Record dashboard snapshots at each step
   - Document findings in a results table
2. Add continuous ramp feature to perf-test harness
   - New flag: `--ramp-rate` (games per minute, sustained)
   - Games keep spawning while existing games play
3. Run continuous ramp through the identified bottleneck zone
   - Identify exact inflection point
4. Run soak test at 80% of ceiling for 2+ hours
   - Watch for memory/goroutine/connection leaks
5. Produce bottleneck report: what broke, at what scale, with dashboard evidence

### Phase 3: Targeted Optimization (ongoing)

Driven entirely by Phase 2 findings. No pre-committed work items.

1. Pick the #1 bottleneck from the Phase 2 report
2. Implement the simplest fix from the optimization catalog
3. Re-run the load test at the same scale
4. Compare dashboards before/after — quantify the improvement
5. If ceiling moved, find the new ceiling (repeat Phase 2 steps)
6. Repeat until satisfied or until architectural limits are reached

Each optimization cycle should be a single PR with before/after metrics in the PR description.
