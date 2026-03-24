#!/usr/bin/env bash
# common.sh — Shared functions for perf-test runner scripts.
# Source this file from run-baseline.sh and run-staircase.sh.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
ENV_FILE="$PROJECT_ROOT/component-test/.env"

# resolve_anon_key sets ANON_KEY from the env file if not already exported.
resolve_anon_key() {
    if [[ -z "${ANON_KEY:-}" ]]; then
        if [[ -f "$ENV_FILE" ]]; then
            ANON_KEY=$(grep '^ANON_KEY=' "$ENV_FILE" | cut -d= -f2)
        fi
    fi

    if [[ -z "${ANON_KEY:-}" ]]; then
        echo "ERROR: ANON_KEY not found. Set it in $ENV_FILE or export ANON_KEY." >&2
        exit 1
    fi

    export ANON_KEY
}

# check_docker_stack verifies the LGTM container is running.
check_docker_stack() {
    if ! docker compose --env-file "$ENV_FILE" -f "$PROJECT_ROOT/docker-compose.yml" \
        ps --format '{{.Name}}' 2>/dev/null | grep -q "lgtm"; then
        echo "ERROR: LGTM container not running. OTel metrics will not be collected." >&2
        echo "  Start the stack: cd $PROJECT_ROOT && docker compose --env-file component-test/.env up -d" >&2
        exit 1
    fi
}

# has_flag checks if any of the passed arguments matches the given flag name.
# Usage: has_flag "--preset" "$@"
has_flag() {
    local flag="$1"
    shift
    for arg in "$@"; do
        if [[ "$arg" == "$flag" ]]; then
            return 0
        fi
    done
    return 1
}

# extract_flag_value extracts the value following a flag, or returns a default.
# Usage: extract_flag_value "--preset" "default_value" "$@"
extract_flag_value() {
    local flag="$1"
    local default_val="$2"
    shift 2
    local next_is_value=false
    for arg in "$@"; do
        if $next_is_value; then
            echo "$arg"
            return
        fi
        if [[ "$arg" == "$flag" ]]; then
            next_is_value=true
        fi
    done
    echo "$default_val"
}

# build_default_args populates DEFAULT_ARGS with common observability flags.
build_default_args() {
    OTEL_ENDPOINT="${OTEL_ENDPOINT:-localhost:4318}"
    GRAFANA_URL="${GRAFANA_URL:-http://localhost:3000}"

    DEFAULT_ARGS=(
        --anon-key "$ANON_KEY"
        --otel-endpoint "$OTEL_ENDPOINT"
        --grafana-url "$GRAFANA_URL"
    )
}

# print_banner prints the run header.
# Usage: print_banner "Perf Test Baseline"
print_banner() {
    local title="$1"
    echo "=== $title ==="
    echo "  OTel:    $OTEL_ENDPOINT"
    echo "  Grafana: $GRAFANA_URL"
    echo "  Args:    ${DEFAULT_ARGS[*]} $*"
    echo ""
}
