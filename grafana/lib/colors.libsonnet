// Semantic boundary colors for go-risk-it dashboards.
// Each color maps to an architectural boundary in the system.
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

  // Helper: returns a Grafana fixed-color object
  fixedColor(hex):: { mode: 'fixed', fixedColor: hex },
}
