#!/usr/bin/env bash
# run-baseline.sh — Run a perf-test baseline with all observability enabled.
#
# Usage:
#   ./run-baseline.sh                    # baseline preset, auto-named
#   ./run-baseline.sh --preset light     # different preset
#   ./run-baseline.sh --games 2          # override specific flags
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
    --output text
    --save-baseline
)

# If no --preset and no --games passed, default to baseline preset.
HAS_PRESET=false
HAS_GAMES=false
for arg in "$@"; do
    case "$arg" in
        --preset) HAS_PRESET=true ;;
        --games) HAS_GAMES=true ;;
    esac
done

if ! $HAS_PRESET && ! $HAS_GAMES; then
    DEFAULT_ARGS+=(--preset baseline)
fi

# If no --baseline-name passed, auto-generate from preset or "run".
HAS_NAME=false
for arg in "$@"; do
    if [[ "$arg" == "--baseline-name" ]]; then
        HAS_NAME=true
        break
    fi
done

if ! $HAS_NAME; then
    # Extract preset name from args or default.
    PRESET_NAME="baseline"
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
    DEFAULT_ARGS+=(--baseline-name "$PRESET_NAME")
fi

echo "=== Perf Test Baseline ==="
echo "  OTel:    $OTEL_ENDPOINT"
echo "  Grafana: $GRAFANA_URL"
echo "  Args:    ${DEFAULT_ARGS[*]} $*"
echo ""

cd "$SCRIPT_DIR"
exec go run ./cmd/loadtest/ "${DEFAULT_ARGS[@]}" "$@"
