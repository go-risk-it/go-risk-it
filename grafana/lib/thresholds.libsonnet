// SLO threshold values extracted from existing dashboard JSON files.
// Each entry is a Grafana thresholds object { mode, steps[] }.
{
  // WebSocket dashboard: Active Connections stat panel
  // Source: websocket.json panel 1
  wsConnections: {
    mode: 'absolute',
    steps: [
      { color: 'green', value: null },
      { color: 'yellow', value: 100 },
      { color: 'red', value: 500 },
    ],
  },

  // Game Engine dashboard: Active Games stat panel
  // Source: game-engine.json panel 1
  activeGames: {
    mode: 'absolute',
    steps: [
      { color: 'green', value: null },
      { color: 'yellow', value: 50 },
      { color: 'red', value: 100 },
    ],
  },

  // Database dashboard: Pool Utilization %
  // Source: database.json panel 4
  poolUtil: {
    mode: 'absolute',
    steps: [
      { color: 'green', value: null },
      { color: 'yellow', value: 60 },
      { color: 'red', value: 80 },
    ],
  },

  // Database dashboard: Postgres Cache Hit Rate (inverted — red is low)
  // Source: database.json panel 8
  cacheHit: {
    mode: 'absolute',
    steps: [
      { color: 'red', value: null },
      { color: 'yellow', value: 90 },
      { color: 'green', value: 99 },
    ],
  },

  // Server Golden Signals dashboard: HTTP Error Rate %
  // Source: server-golden-signals.json panel 9
  httpError: {
    mode: 'absolute',
    steps: [
      { color: 'green', value: null },
      { color: 'yellow', value: 1 },
      { color: 'red', value: 5 },
    ],
  },

  // Database dashboard: Canceled Acquires stat panel
  // Source: database.json panel 10
  canceledAcquires: {
    mode: 'absolute',
    steps: [
      { color: 'green', value: null },
      { color: 'red', value: 0.01 },
    ],
  },

  // --- Perf Test Command Center SLOs ---

  // E2E p95 latency. SLO: < 500ms
  // Source: perf-test-command-center.json panel 1
  e2eP95: {
    mode: 'absolute',
    steps: [
      { color: 'green', value: null },
      { color: 'yellow', value: 0.5 },
      { color: 'red', value: 1 },
    ],
  },

  // WS Delivery p95 latency. SLO: < 200ms
  // Source: perf-test-command-center.json panel 2
  wsDeliveryP95: {
    mode: 'absolute',
    steps: [
      { color: 'green', value: null },
      { color: 'yellow', value: 0.2 },
      { color: 'red', value: 0.5 },
    ],
  },

  // DB Transaction p95 latency. SLO: < 50ms
  // Source: perf-test-command-center.json panel 3
  dbTxnP95: {
    mode: 'absolute',
    steps: [
      { color: 'green', value: null },
      { color: 'yellow', value: 0.05 },
      { color: 'red', value: 0.1 },
    ],
  },

  // HTTP Error Rate (percentunit). SLO: < 1%
  // Source: perf-test-command-center.json panel 4
  httpErrorRate: {
    mode: 'absolute',
    steps: [
      { color: 'green', value: null },
      { color: 'yellow', value: 0.01 },
      { color: 'red', value: 0.05 },
    ],
  },

  // Game Completion Rate (inverted — red is low)
  completionRate: {
    mode: 'absolute',
    steps: [
      { color: 'red', value: null },
      { color: 'yellow', value: 0.8 },
      { color: 'green', value: 0.95 },
    ],
  },

  // Fan-out Amplification (WS broadcasts per game move)
  // Normal: ~4 (one broadcast per player per move). Watch: > 10 suggests broadcast storms.
  ccFanOut: {
    mode: 'absolute',
    steps: [
      { color: 'green', value: null },
      { color: 'yellow', value: 10 },
      { color: 'red', value: 20 },
    ],
  },

  // DB Latency Share % (DB p95 as % of HTTP p95)
  // Normal: 30-50%. Watch: > 70% means DB dominates request time.
  ccDbLatencyShare: {
    mode: 'absolute',
    steps: [
      { color: 'green', value: null },
      { color: 'yellow', value: 70 },
      { color: 'red', value: 90 },
    ],
  },

  // --- Perf Test (Client-Side) dashboard ---

  // Active Games stat panel (perf-test scale, higher red threshold than game-engine)
  // Source: perf-test.json panel 1
  perfTestActiveGames: {
    mode: 'absolute',
    steps: [
      { color: 'green', value: null },
      { color: 'yellow', value: 50 },
      { color: 'red', value: 200 },
    ],
  },
}
