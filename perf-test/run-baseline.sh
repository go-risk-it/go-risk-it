#!/usr/bin/env bash
# run-baseline.sh — Run a perf-test baseline with all observability enabled.
#
# Usage:
#   ./run-baseline.sh                    # baseline preset, auto-named
#   ./run-baseline.sh --preset light     # different preset
#   ./run-baseline.sh --games 2          # override specific flags
#
# Prerequisites:
#   1. Docker compose stack running: docker compose --env-file component-test/.env up -d
#   2. ANON_KEY set in component-test/.env
set -euo pipefail

PERF_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$PERF_DIR/scripts/common.sh"

resolve_anon_key
check_docker_stack
build_default_args

DEFAULT_ARGS+=(
    --output text
    --save-baseline
)

# If no --preset and no --games passed, default to baseline preset.
if ! has_flag "--preset" "$@" && ! has_flag "--games" "$@"; then
    DEFAULT_ARGS+=(--preset baseline)
fi

# If no --baseline-name passed, auto-generate from preset or "baseline".
if ! has_flag "--baseline-name" "$@"; then
    PRESET_NAME=$(extract_flag_value "--preset" "baseline" "$@")
    DEFAULT_ARGS+=(--baseline-name "$PRESET_NAME")
fi

# --- Wait for API readiness through Kong ---
source "$PERF_DIR/wait-for-api.sh"

print_banner "Perf Test Baseline" "$@"

cd "$PERF_DIR"
exec go run ./cmd/loadtest/ "${DEFAULT_ARGS[@]}" "$@"
