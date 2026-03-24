#!/usr/bin/env bash
# run-staircase.sh — Run a staircase load test with all observability enabled.
#
# Usage:
#   ./run-staircase.sh                          # default staircase preset
#   ./run-staircase.sh --preset staircase-light # lighter preset
#   ./run-staircase.sh --steps 5,10,20          # custom steps (no preset)
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

DEFAULT_ARGS+=(--save-journal)

# If no --preset and no --steps passed, default to staircase preset.
if ! has_flag "--preset" "$@" && ! has_flag "--steps" "$@"; then
    DEFAULT_ARGS+=(--preset staircase)
fi

# If no --journal-name passed, auto-generate from preset or "staircase".
if ! has_flag "--journal-name" "$@"; then
    PRESET_NAME=$(extract_flag_value "--preset" "staircase" "$@")
    DEFAULT_ARGS+=(--journal-name "$PRESET_NAME")
fi

# --- Wait for API readiness through Kong ---
source "$PERF_DIR/wait-for-api.sh"

print_banner "Staircase Load Test" "$@"

cd "$PERF_DIR"
exec go run ./cmd/loadtest/ "${DEFAULT_ARGS[@]}" "$@"
