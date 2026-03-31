# go-risk-it — Personal Project

This is a personal project (Risk board game in Go), **not** related to Booking.com work.

## Overrides to global instructions

- **No session logging** — Do NOT write entries to `~/claude-work-log/`. Skip the "Automatic Session Logging" section from global CLAUDE.md entirely.
- **No domain knowledge persistence** — Do NOT write to `~/.claude/knowledge/`. Skip the "Knowledge Management" rules.
- **No Jira conventions** — No ticket prefixes in branch names or commit messages. Use simple descriptive names.
- **No WBSO** — Skip any WBSO-related skills or time tracking.
- **No standup/wrap-up** — Skip work log based skills (standup, weekly-summary, follow-ups, wrap-up, today).

## What still applies

- All coding behavior rules (execute don't plan, no over-engineering, security practices)
- Git best practices (but without Jira ticket conventions)
- Project MEMORY.md in `.claude/projects/` for project-specific context

## Observability

This project uses a **spanmetrics-first** observability architecture. Three documentation locations cover the full stack:

- **`internal/kernel/observe/doc.go`** — API reference for the 5-function observe package (`Span`, `SpanEvent`, `Info`, `Warn`, `Error`), the `done(error)` lifecycle pattern, span naming conventions, three-signal decision framework, and OTel initialization chain
- **`internal/kernel/metrics/doc.go`** — Why only state gauges are manual instruments, and what NOT to add (latency histograms, per-operation counters, error rates — all derived from spans)
- **`grafana/README.md`** — Full pipeline documentation: OTel setup, three-signal model, spanmetrics connector, collector dimensions, `spans::` catalog, RED metric helpers, dashboard toolchain, and how to add a new metric

**Key rules:**
- Business logic (`game/logic/`, `lobby/logic/`, `game/consumers/`, `lobby/consumers/`) must use `kernel/observe` — never import `log/slog` or `go.opentelemetry.io/otel` directly (enforced by arch_test rules O1 and O2)
- RED metrics (Rate, Errors, Duration) are derived from spans by the OTel Collector's spanmetrics connector — do not create manual RED instruments
- Only 4 manual OTel instruments exist: `ws.connections.active`, `db.transaction.retries.total`, `game.active`, `game.duration` (resource state with no span equivalent)
