#!/usr/bin/env bash
set -Eeuo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
compose_file="$repo_root/scripts/e2e/docker-compose.yml"
compose_project="openstate-e2e"
compose=(docker compose -p "$compose_project" -f "$compose_file")

export E2E_PG_PORT="${E2E_PG_PORT:-55437}"
export E2E_REDIS_PORT="${E2E_REDIS_PORT:-56381}"
export E2E_API_PORT="${E2E_API_PORT:-8021}"
export E2E_WEB_PORT="${E2E_WEB_PORT:-3020}"
export DATABASE_URL="${DATABASE_URL:-postgresql://postgres:postgres@127.0.0.1:${E2E_PG_PORT}/openstate_e2e}"
export REDIS_URL="${REDIS_URL:-redis://127.0.0.1:${E2E_REDIS_PORT}}"
export E2E_FIXTURE_PASSWORD="${E2E_FIXTURE_PASSWORD:-openstate-e2e-fixture-password}"
export E2E_TEST_SECRET="${E2E_TEST_SECRET:-openstate-e2e-only-secret-0123456789-abcdef}"
export API_PORT="$E2E_API_PORT"
export PORT="$E2E_API_PORT"
export E2E_FIXTURES=1

status=0
cleanup() {
  status=$?
  if [ "$status" -ne 0 ]; then
    mkdir -p "$repo_root/apps/web/test-results/e2e"
    "${compose[@]}" logs --no-color --tail=200 > "$repo_root/apps/web/test-results/e2e/stack.log" 2>&1 || true
  fi
  "${compose[@]}" down --volumes --remove-orphans >/dev/null 2>&1 || true
  exit "$status"
}
trap cleanup EXIT

"${compose[@]}" up -d --wait

if ! command -v goose >/dev/null 2>&1; then
  printf 'goose is required for browser E2E migrations. Install it before running test:e2e.\n' >&2
  exit 1
fi

(cd "$repo_root/apps/api" && goose -dir db/migrations postgres "$DATABASE_URL" up)
(cd "$repo_root/apps/api" && go run ./cmd/e2e-fixtures --mode seed)
(cd "$repo_root/apps/api" && go run ./cmd/e2e-fixtures --mode verify)

bun run --cwd "$repo_root/apps/web" test:e2e
