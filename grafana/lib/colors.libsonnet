// Semantic boundary colors for go-risk-it dashboards.
// Each color maps to an architectural boundary in the system.
//
// COLORING RULES:
// 1. Single-series panels → color=colors.fixedColor(colors.XX) — boundary identification
// 2. Multi-series data-driven panels (by route, by phase, by label) → no color param (palette-classic)
// 3. Multi-series percentile panels (p50/p95/p99) → + modifiers.withPercentileColors('XX')
{
  // Architectural boundaries
  db: '#3274D9',        // Database — blue
  ws: '#8F3BB8',        // WebSocket — purple
  gameLogic: '#56A64B', // Game Logic — green
  http: '#FF9830',      // HTTP — amber
  errors: '#E02F44',    // Errors — red
  client: '#73BF69',    // Client/perf-test — cyan-green
  eventBus: '#00BCD4',  // Event Bus — teal (post-response async)

  // Postgres lock mode colors — keyed by pg_locks mode label values
  lockModes: {
    AccessShareLock: '#73BF69',       // light green — read-only, harmless
    RowShareLock: '#56A64B',          // green
    RowExclusiveLock: '#FF9830',      // amber — writes
    ShareUpdateExclusiveLock: '#FADE2A',  // yellow
    ShareLock: '#3274D9',             // blue
    ShareRowExclusiveLock: '#8F3BB8', // purple
    ExclusiveLock: '#FF780A',         // dark orange — contention risk
    AccessExclusiveLock: '#E02F44',   // red — DDL, blocks everything
  },

  // Event type colors — keyed by event_type metric label values
  eventTypes: {
    move_executed: '#56A64B',       // gameLogic green — dominant heartbeat
    phase_transitioned: '#73BF69',  // light green
    game_created: '#73BF69',        // light green
    game_completed: '#3274D9',      // db blue
    player_connected: '#8AB8FF',    // light blue
    player_eliminated: '#E02F44',   // errors red
    continent_captured: '#FF9830',  // http amber
    continent_lost: '#FF780A',      // dark orange
  },

  // Game phase colors — keyed by phase name
  phase: {
    deploy: '#3274D9',     // blue — placing troops
    attack: '#E02F44',     // red — combat
    cards: '#FADE2A',      // gold — card exchange
    conquer: '#56A64B',    // green — claiming territory
    reinforce: '#8F3BB8',  // purple — troop movement
  },

  // Operational signal colors — severity/status semantics
  signal: {
    ok: '#56A64B',         // green — healthy/passing
    warning: '#FF9830',    // amber — degraded/attention
    'error': '#E02F44',    // red — failure/breach
    info: '#3274D9',       // blue — informational
    muted: '#8D8D8D',      // gray — inactive/disabled
  },

  // Event headline colors — notable game moments
  headline: {
    continent: '#FADE2A',  // gold — continent captured/lost
    elimination: '#E02F44', // red — player eliminated
    victory: '#56A64B',    // green — game won
  },

  // Shade variants for percentile panels (p50=light, p95=medium, p99=dark).
  // Used by modifiers.withPercentileColors().
  shades: {
    db: { light: '#73A9F2', medium: '#3274D9', dark: '#1F4FA0' },
    ws: { light: '#B57ADB', medium: '#8F3BB8', dark: '#6A2A8A' },
    gameLogic: { light: '#8CCF82', medium: '#56A64B', dark: '#3D7A35' },
    http: { light: '#FFBE73', medium: '#FF9830', dark: '#CC7A26' },
    errors: { light: '#F07A8B', medium: '#E02F44', dark: '#A82233' },
    client: { light: '#A8DB9F', medium: '#73BF69', dark: '#52914A' },
    eventBus: { light: '#4DD9E8', medium: '#00BCD4', dark: '#008C9E' },
  },

  // Helper: returns a Grafana fixed-color object
  fixedColor(hex):: { mode: 'fixed', fixedColor: hex },
}
