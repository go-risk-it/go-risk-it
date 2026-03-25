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

  // Helper: returns a Grafana fixed-color object
  fixedColor(hex):: { mode: 'fixed', fixedColor: hex },
}
