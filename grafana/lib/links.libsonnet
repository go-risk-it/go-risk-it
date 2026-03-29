// Dashboard data link helpers for go-risk-it Grafana dashboards.
// Generates uid-based links with time range interpolation.
{
  // Named dashboard UIDs — one constant per dashboard.
  dashboardUids: {
    systemHealth: 'system-health',
    gameEngine: 'game-engine',
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
}
