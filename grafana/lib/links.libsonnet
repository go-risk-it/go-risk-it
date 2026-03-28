// Dashboard data link helpers for go-risk-it Grafana dashboards.
// Generates uid-based links with time range interpolation.
{
  // Named dashboard UIDs — one constant per dashboard.
  dashboardUids: {
    websocket: 'websocket',
    database: 'database',
    gameEngine: 'game-engine',
    serverGoldenSignals: 'server-golden-signals',
    perfTest: 'perf-test',
    perfTestCommandCenter: 'perf-test-command-center',
    requestLifecycle: 'request-lifecycle',
  },

  // Build a Grafana data link to another dashboard, preserving time range.
  // title: string — link display text
  // uid: string — target dashboard UID (use dashboardUids.* constants)
  toDashboard(title, uid):: {
    title: title,
    url: '/d/' + uid + '?from=${__from}&to=${__to}',
    targetBlank: true,
  },
}
