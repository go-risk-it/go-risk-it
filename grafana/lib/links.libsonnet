// Dashboard data link helpers for go-risk-it Grafana dashboards.
// Generates uid-based links with time range interpolation.
//
// Cross-dashboard variable contract
// ──────────────────────────────────
// Dashboard variables referenced in toDashboard() URLs or expected by
// target dashboards. Any dashboard that defines one of these variables
// MUST use the canonical name and type listed here.
//
//   Variable   | Type | Defined on          | Purpose
//   -----------|------|---------------------|--------------------------------------------
//   $gameId    | text | game-engine,        | Filters panels to a single game instance
//              |      | game-theater        |
//   $traceId   | text | system-health,      | Deep-links into Tempo trace view
//              |      | game-theater        |
//
// Dashboard UIDs (used as link targets):
//   system-health  — infrastructure & RED overview
//   game-engine    — per-game operational detail
//   game-theater   — single-game deep dive (phase tapestry, event chronicle, trace forensics)
//   perf-test      — load test analysis
//
// When adding a new dashboard, register its UID in dashboardUids below
// and document any new variables in the table above.
//
// Constructors:
//   toDashboard(title, uid) — basic cross-dashboard link with time range
//   toDashboardWithVar(title, uid, varName, varValue) — link with variable pre-population
{
  // Named dashboard UIDs — one constant per dashboard.
  dashboardUids: {
    systemHealth: 'system-health',
    gameEngine: 'game-engine',
    gameTheater: 'game-theater',
    perfTest: 'perf-test',
  },

  // Build a Grafana data link to another dashboard, preserving time range.
  // title: string — link display text
  // uid: string — target dashboard UID (use dashboardUids.* constants)
  toDashboard(title, uid):: {
    title: title,
    url: '/d/' + uid + '?from=${__from}&to=${__to}',
    targetBlank: true,
  },

  // Build a Grafana data link to another dashboard, preserving time range
  // and pre-populating a template variable.
  // title: string — link display text
  // uid: string — target dashboard UID
  // varName: string — variable name without $ prefix (e.g. 'gameId')
  // varValue: string — Grafana interpolation expression (e.g. '${__field.labels.game_id}')
  toDashboardWithVar(title, uid, varName, varValue):: {
    title: title,
    url: '/d/' + uid + '?from=${__from}&to=${__to}&var-' + varName + '=' + varValue,
    targetBlank: true,
  },
}
