# Operations Guide

This document explains how to run, verify, and maintain this starter in conditions that are truly supported by the current codebase.

## Scope

This repo currently supports the following baseline:
- `apps/web`: Next.js App Router on port `3020`
- web auth foundation: NextAuth credentials + internal BFF proxy route
- `apps/api`: Hono API on port `8020`
- `apps/worker`: worker scaffold connected to Redis with no active default queue
- `docker-compose.yml`: provides local PostgreSQL and Redis
- `scripts/`: repo-level helper executables for local bootstrap and supporting workflows
- `Prisma`: CLI and local database infra ready, though default schema is still generic scaffold

Structure conventions:
- `docs/` only for documentation and templates
- `scripts/` for repo-level executable files like shell helpers

## Local Topology

| Component | URL / Port | Purpose |
| --- | --- | --- |
| Web | `http://localhost:3020` | UI starter with login, panel, and BFF proxy |
| Web Auth | `http://localhost:3020/api/auth/*` | NextAuth route handler for session and credentials flow |
| Web Proxy | `http://localhost:3020/api/proxy/*` | Internal proxy for all browser requests to backend |
| API | `http://localhost:8020` | HTTP API for root and health check |
| Worker | n/a | Idle worker scaffold for background runtime |
| PostgreSQL | `postgresql://postgres:postgres@127.0.0.1:5432/openstate?schema=public` | Local database for next Prisma features |
| Redis | `redis://127.0.0.1:6379` | Ready-to-use broker for next queue features |
| Prisma Studio | `http://localhost:5555` | Browser UI for local database when schema is in use |

## Supported Baseline

Local baseline considered healthy for this repo:
1. `bun install` succeeds.
2. Local PostgreSQL and Redis are active.
3. `bun run dev` runs web, api, and worker.
4. `bun run check` passes.
5. Starter login page can be opened at `http://localhost:3020/login`.
6. Default protected route `/panel` redirects to `/login` when no session exists.

## Environment Matrix

### Root

Repo does not use root `.env` as source of truth. Env is managed per app.

### `apps/web/.env.example`

| Variable | Required | Default | Notes |
| --- | --- | --- | --- |
| `NEXT_PUBLIC_APP_URL` | no | `http://localhost:3020` | Web app base URL |
| `NEXT_PUBLIC_API_URL` | no | `http://localhost:8020` | Fallback backend base URL for server-side auth config |
| `NEXT_PUBLIC_API_PROXY_BASE_URL` | no | `/api/proxy` | Internal proxy base path used by axios in browser |
| `NEXTAUTH_URL` | no | `http://localhost:3020` | Canonical web URL for NextAuth |
| `NEXTAUTH_SECRET` | yes for production | none | Secret for sign/encrypt session cookie |
| `API_URL` | no | `http://localhost:8020` | Backend base URL used by proxy route server-side |
| `AUTH_LOGIN_PATH` | no | `/auth/login` | Backend path for credentials login |
| `AUTH_REFRESH_PATH` | no | `/auth/refresh` | Backend path for token refresh |
| `AUTH_LOGOUT_PATH` | no | `/auth/logout` | Backend path for logout token/session |

### `apps/api/.env.example`

| Variable | Required | Default | Notes |
| --- | --- | --- | --- |
| `API_PORT` | no | `8020` | HTTP API port |
| `DATABASE_URL` | no | `postgresql://postgres:postgres@127.0.0.1:5432/openstate?schema=public` | Default local PostgreSQL for Prisma |

### `apps/worker/.env.example`

| Variable | Required | Default | Notes |
| --- | --- | --- | --- |
| `REDIS_URL` | no | `redis://127.0.0.1:6379` | Redis for BullMQ worker |

## Fast Start

### One-time bootstrap

```bash
bun install
bun run bootstrap
```

`bun run bootstrap` will:
1. create `.env` files from `.env.example` if not yet exists
2. start PostgreSQL and Redis via Docker Compose
3. wait for both services to be ready
4. generate Prisma client
5. generate merged OpenAPI spec

### Manual setup

```bash
cp apps/web/.env.example apps/web/.env
cp apps/api/.env.example apps/api/.env
cp apps/worker/.env.example apps/worker/.env
bun run stack:up
bun run prisma:generate
bun run openapi:generate
```

## Daily Commands

### Stack management

```bash
bun run stack:up
bun run stack:down
bun run stack:logs
bun run stack:ps
bun run stack:reset
```

### App runtime

```bash
bun run dev
bun run dev:web
bun run dev:api
bun run dev:worker
```

## Auth and Proxy Flow

Current frontend auth baseline uses BFF pattern:
1. User submits form at `/login`.
2. NextAuth credentials in `apps/web/auth.ts` sends request to `/api/proxy/auth/login`.
3. Route `apps/web/app/api/proxy/[...path]/route.ts` forwards that request to backend `API_URL`.
4. Access token and refresh token are stored in NextAuth session cookie based on JWT.
5. Subsequent browser requests from hooks/frontend always hit `/api/proxy/*`, not backend directly.
6. Internal proxy adds bearer token, attempts token refresh when nearly expired, then retries if backend returns `401`.
7. `apps/web/proxy.ts` guards default protected route `/panel` and redirects guests to `/login`.

### Main quality gate

```bash
bun run check
```

This runs:
- skill assets validation
- lint
- typecheck
- test
- build

### Partial commands

```bash
bun run session:status
bun run lint
bun run typecheck
bun run test
bun run build
```

`bun run session:status` is used by onboarding flow when user only types `Mulai` or `Start` to agent.

## OpenAPI Workflow

Current OpenAPI source of truth:
- `docs/openapi/base.json`
- `docs/openapi/paths/*.json`
- `docs/openapi/schemas/*.json`

Operational command:

```bash
bun run openapi:generate
```

Operational rules:
1. Modify split files per feature, not `docs/openapi.json` directly.
2. Run `bun run openapi:generate` after changes.
3. Commit both split files and merged `docs/openapi.json` so tooling that reads merged file stays in sync.

## Prisma Workflow

Local Prisma infra is now available, though default schema is still generic scaffold.

Operational commands:

```bash
bun run prisma:generate
bun run prisma:migrate:dev
bun run prisma:seed
bun run prisma:studio
```

Current correct status:
- `apps/api/prisma/schema.prisma` is still generic scaffold
- `apps/api/prisma/seed.ts` is only placeholder
- Local PostgreSQL is already available in Docker Compose
- Prisma client can be generated as part of bootstrap

Meaning:
- local database infra is ready
- but first business schema must be defined before migrate/seed becomes meaningful

## Queue and Worker Workflow

Worker currently runs as idle scaffold:
1. worker process can still be run
2. Redis connection is available for next queue features
3. no active queue or job handler by default

Meaning:
- background worker stack is available
- first queue needs to be added when feature actually requires it

## Seed Status

Current seed behavior is intentional: file `apps/api/prisma/seed.ts` only writes placeholder message. No default seed data should be assumed to exist.

## CI and Branch Hygiene

Active workflows:
- app CI: `.github/workflows/app-ci.yml`
- skill hygiene: `.github/workflows/skills-hygiene.yml`

Expectations before merge:
1. `bun run check` passes locally.
2. OpenAPI is regenerated if split files changed.
3. Feature docs and registry are synced if new feature exists.

## Troubleshooting

### Bootstrap fails because Docker not running
- ensure Docker Desktop / daemon is active
- rerun `bun run bootstrap`

### PostgreSQL not ready
- check `bun run stack:ps`
- check logs `bun run stack:logs`
- check if port `5432` is used by another service

### Worker fails to start
- ensure Redis is active: `bun run stack:ps`
- check `REDIS_URL` in `apps/worker/.env`
- rerun `bun run dev:worker`

### OpenAPI merge fails
- ensure all `$ref` point to existing schema in `docs/openapi/schemas/`
- rerun `bun run openapi:generate` to see specific validation errors
