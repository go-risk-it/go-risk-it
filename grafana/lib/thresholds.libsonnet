// SLO threshold values for go-risk-it Grafana dashboards.
// Each entry is a Grafana thresholds object { mode, steps[] }.
{
  // Active WS Connections
  wsConnections: {
    mode: 'absolute',
    steps: [
      { color: 'green', value: null },
      { color: 'yellow', value: 100 },
      { color: 'red', value: 500 },
    ],
  },

  // Active Games
  activeGames: {
    mode: 'absolute',
    steps: [
      { color: 'green', value: null },
      { color: 'yellow', value: 50 },
      { color: 'red', value: 100 },
    ],
  },

  // DB Pool Utilization %
  poolUtil: {
    mode: 'absolute',
    steps: [
      { color: 'green', value: null },
      { color: 'yellow', value: 60 },
      { color: 'red', value: 80 },
    ],
  },

  // Postgres Cache Hit Rate (inverted — red is low)
  cacheHit: {
    mode: 'absolute',
    steps: [
      { color: 'red', value: null },
      { color: 'yellow', value: 90 },
      { color: 'green', value: 99 },
    ],
  },

  // HTTP Error Rate %
  httpError: {
    mode: 'absolute',
    steps: [
      { color: 'green', value: null },
      { color: 'yellow', value: 1 },
      { color: 'red', value: 5 },
    ],
  },

  // Canceled Acquires
  canceledAcquires: {
    mode: 'absolute',
    steps: [
      { color: 'green', value: null },
      { color: 'red', value: 0.01 },
    ],
  },

  // E2E p95 latency. SLO: < 500ms
  e2eP95: {
    mode: 'absolute',
    steps: [
      { color: 'green', value: null },
      { color: 'yellow', value: 0.5 },
      { color: 'red', value: 1 },
    ],
  },

  // WS Delivery p95 latency. SLO: < 200ms
  wsDeliveryP95: {
    mode: 'absolute',
    steps: [
      { color: 'green', value: null },
      { color: 'yellow', value: 0.2 },
      { color: 'red', value: 0.5 },
    ],
  },

  // DB Transaction p95 latency. SLO: < 50ms
  dbTxnP95: {
    mode: 'absolute',
    steps: [
      { color: 'green', value: null },
      { color: 'yellow', value: 0.05 },
      { color: 'red', value: 0.1 },
    ],
  },

  // HTTP Error Rate (percentunit). SLO: < 1%
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
  ccFanOut: {
    mode: 'absolute',
    steps: [
      { color: 'green', value: null },
      { color: 'yellow', value: 10 },
      { color: 'red', value: 20 },
    ],
  },

  // DB Latency Share % (DB p95 as % of HTTP p95)
  ccDbLatencyShare: {
    mode: 'absolute',
    steps: [
      { color: 'green', value: null },
      { color: 'yellow', value: 70 },
      { color: 'red', value: 90 },
    ],
  },

  // Active Games (perf-test scale, higher red threshold)
  perfTestActiveGames: {
    mode: 'absolute',
    steps: [
      { color: 'green', value: null },
      { color: 'yellow', value: 50 },
      { color: 'red', value: 200 },
    ],
  },
}
