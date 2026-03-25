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

# ensure_fresh_stack tears down the existing Docker stack, checks out the
# requested branch (default: main), rebuilds, and starts all containers.
# Skipped if --skip-rebuild is passed to the parent script.
# Use --branch <name> to build from a specific branch (default: main).
ensure_fresh_stack() {
    local target_branch
    target_branch=$(extract_flag_value "--branch" "main" "$@")

    if has_flag "--skip-rebuild" "$@"; then
        echo "[stack] --skip-rebuild: skipping stack rebuild"
        # Still verify the stack is running.
        if ! docker compose --env-file "$ENV_FILE" -f "$PROJECT_ROOT/docker-compose.yml" \
            ps --format '{{.Name}}' 2>/dev/null | grep -q "lgtm"; then
            echo "ERROR: Docker stack not running. Remove --skip-rebuild or start the stack manually." >&2
            exit 1
        fi
        return 0
    fi

    local current_branch
    current_branch=$(git -C "$PROJECT_ROOT" branch --show-current)

    echo "[stack] tearing down existing containers..."
    docker compose --env-file "$ENV_FILE" -f "$PROJECT_ROOT/docker-compose.yml" down 2>&1 | tail -1

    if [[ "$target_branch" != "$current_branch" ]]; then
        echo "[stack] checking out $target_branch..."
        git -C "$PROJECT_ROOT" checkout "$target_branch" --quiet
        git -C "$PROJECT_ROOT" pull --quiet 2>/dev/null || true
    else
        echo "[stack] already on $target_branch"
    fi

    echo "[stack] rebuilding and starting containers from $(git -C "$PROJECT_ROOT" log --oneline -1)..."
    docker compose --env-file "$ENV_FILE" -f "$PROJECT_ROOT/docker-compose.yml" up -d --build 2>&1 | tail -3

    echo "[stack] waiting for containers to be healthy..."
    local max_wait=60
    local waited=0
    while [[ $waited -lt $max_wait ]]; do
        local healthy
        healthy=$(docker compose --env-file "$ENV_FILE" -f "$PROJECT_ROOT/docker-compose.yml" \
            ps --format '{{.Status}}' 2>/dev/null | grep -c "healthy" || true)
        if [[ "$healthy" -ge 3 ]]; then
            echo "[stack] ready ($healthy healthy containers)"
            return 0
        fi
        sleep 2
        waited=$((waited + 2))
    done

    echo "WARNING: not all containers healthy after ${max_wait}s, proceeding anyway" >&2
}

# strip_stack_flags removes --skip-rebuild and --branch <value> from args
# so they don't get forwarded to the Go binary.
strip_stack_flags() {
    local result=()
    local skip_next=false
    for arg in "$@"; do
        if $skip_next; then
            skip_next=false
            continue
        fi
        if [[ "$arg" == "--skip-rebuild" ]]; then
            continue
        fi
        if [[ "$arg" == "--branch" ]]; then
            skip_next=true
            continue
        fi
        result+=("$arg")
    done
    echo "${result[@]}"
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
