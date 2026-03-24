#!/usr/bin/env bash
# run-staircase.sh — Run a staircase load test with all observability enabled.
#
# Usage:
#   ./run-staircase.sh                          # default staircase preset
#   ./run-staircase.sh --preset staircase-light # lighter preset
#   ./run-staircase.sh --steps 5,10,20          # custom steps (no preset)
#
# This script ensures OTel metrics export and Grafana annotations are always
# enabled when the docker compose stack is running. Any extra flags are passed
# through to the loadtest binary.
#
# Prerequisites:
#   1. Docker compose stack running: docker compose --env-file component-test/.env up -d
#   2. ANON_KEY set in component-test/.env
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
ENV_FILE="$PROJECT_ROOT/component-test/.env"

# --- Resolve ANON_KEY ---
if [[ -z "${ANON_KEY:-}" ]]; then
    if [[ -f "$ENV_FILE" ]]; then
        ANON_KEY=$(grep '^ANON_KEY=' "$ENV_FILE" | cut -d= -f2)
    fi
fi

if [[ -z "${ANON_KEY:-}" ]]; then
    echo "ERROR: ANON_KEY not found. Set it in $ENV_FILE or export ANON_KEY." >&2
    exit 1
fi

# --- Check docker stack ---
if ! docker compose --env-file "$ENV_FILE" -f "$PROJECT_ROOT/docker-compose.yml" \
    ps --format '{{.Name}}' 2>/dev/null | grep -q "lgtm"; then
    echo "ERROR: LGTM container not running. OTel metrics will not be collected." >&2
    echo "  Start the stack: cd $PROJECT_ROOT && docker compose --env-file component-test/.env up -d" >&2
    exit 1
fi

# --- Defaults (overridable via flags) ---
OTEL_ENDPOINT="${OTEL_ENDPOINT:-localhost:4318}"
GRAFANA_URL="${GRAFANA_URL:-http://localhost:3000}"

# --- Build default args ---
DEFAULT_ARGS=(
    --anon-key "$ANON_KEY"
    --otel-endpoint "$OTEL_ENDPOINT"
    --grafana-url "$GRAFANA_URL"
    --save-journal
)

# If no --preset and no --steps passed, default to staircase preset.
HAS_PRESET=false
HAS_STEPS=false
for arg in "$@"; do
    case "$arg" in
        --preset) HAS_PRESET=true ;;
        --steps) HAS_STEPS=true ;;
    esac
done

if ! $HAS_PRESET && ! $HAS_STEPS; then
    DEFAULT_ARGS+=(--preset staircase)
fi

# If no --journal-name passed, auto-generate from preset or "staircase".
HAS_NAME=false
for arg in "$@"; do
    if [[ "$arg" == "--journal-name" ]]; then
        HAS_NAME=true
        break
    fi
done

if ! $HAS_NAME; then
    # Extract preset name from args or default.
    PRESET_NAME="staircase"
    NEXT_IS_PRESET=false
    for arg in "$@"; do
        if $NEXT_IS_PRESET; then
            PRESET_NAME="$arg"
            break
        fi
        if [[ "$arg" == "--preset" ]]; then
            NEXT_IS_PRESET=true
        fi
    done
    DEFAULT_ARGS+=(--journal-name "$PRESET_NAME")
fi

# --- Wait for API readiness through Kong ---
source "$SCRIPT_DIR/wait-for-api.sh"

echo "=== Staircase Load Test ==="
echo "  OTel:    $OTEL_ENDPOINT"
echo "  Grafana: $GRAFANA_URL"
echo "  Args:    ${DEFAULT_ARGS[*]} $*"
echo ""

cd "$SCRIPT_DIR"
exec go run ./cmd/loadtest/ "${DEFAULT_ARGS[@]}" "$@"
