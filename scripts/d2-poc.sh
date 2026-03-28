#!/usr/bin/env bash
set -euo pipefail

# d2-poc.sh — Generate a D2 architecture diagram from representative packages.
# Produces docs/d2-poc.svg showing layer containers and key dependency arrows.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
D2_SRC="$(mktemp)"
OUTPUT="$ROOT_DIR/docs/d2-poc.svg"

trap 'rm -f "$D2_SRC"' EXIT

cat > "$D2_SRC" << 'D2'
direction: right

# Layer containers with representative packages

infrastructure: Infrastructure {
  style.fill: "#e8f5e9"
  config
  rand
  metrics
  tracing
  slog
}

ctx: Ctx {
  style.fill: "#fff3e0"
  ctx
}

events: Events {
  style.fill: "#e3f2fd"
  events.bus: "events (Bus)"
  events.logger: "events/logger"
  events.game: "events/game"
  events.lobby: "events/lobby"
}

data: Data {
  style.fill: "#fce4ec"
  data.db: "data/db"
  data.pool: "data/pool"
  data.migration: "data/migration"
}

shared: Shared {
  style.fill: "#f3e5f5"
  logic.errors: "logic/errors"
}

logic: Logic {
  style.fill: "#e8eaf6"
  orchestration: "move/orchestration"
  board: "game/board"
  phase: "game/phase"
  snapshot: "game/snapshot"
  creation: "game/creation"
  lobby.start: "lobby/start"
}

api: API {
  style.fill: "#efebe9"
  api.game.msg: "game/messaging"
  api.game.req: "game/rest/request"
  api.lobby.msg: "lobby/messaging"
}

web: Web {
  style.fill: "#fffde7"
  controller: "game/controller"
  publisher: "game/publisher"
  ws: "game/ws"
  middleware: "middleware"
  rest.route: "rest/route"
}

# Key dependency arrows (layer boundaries)
web.controller -> logic.orchestration: "move request"
web.publisher -> events.bus: "subscribe"
logic.orchestration -> data.db: "transaction"
logic.orchestration -> events.bus: "emit events"
logic.orchestration -> logic.phase: "walk/advance"
logic.board -> data.db: "query regions"
web.controller -> api.game.req: "parse DTOs"
web.ws -> ctx.ctx: "game context"
logic.orchestration -> shared.logic.errors: "domain errors"
data.db -> infrastructure.config: {style.opacity: 0}
D2

if ! command -v d2 &> /dev/null; then
    echo "ERROR: d2 is not installed. Install with: brew install d2" >&2
    echo "Generating D2 source only at: $D2_SRC" >&2
    cat "$D2_SRC"
    exit 1
fi

echo "Generating D2 diagram..."
d2 --theme 200 --layout elk "$D2_SRC" "$OUTPUT"
echo "Generated: $OUTPUT"
