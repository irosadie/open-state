# Deployment Guide

This guide covers how to run OpenState locally, via Docker, and (as a reference)
on Kubernetes. It focuses on the Go API + Next.js web + worker, backed by
PostgreSQL and Redis.

## Components

| Component | Tech | Purpose |
| --- | --- | --- |
| `apps/api` | Go (Echo, Clean Architecture) | HTTP API + MCP server, port `8020` |
| `apps/web` | Next.js (App Router) | State Builder + admin console, port `3020` |
| `apps/worker` | Go (asynq) | Background jobs (timeouts, outbox, delayed events) |
| `packages/go-shared` | Go module | Shared DomainError / types |
| PostgreSQL | `postgres:16` | Source of truth, port `5437` |
| Redis | `redis:7` | Worker broker, port `6381` |

## Local Development

### Prerequisites

- [Go](https://golang.org/dl/) 1.26+
- [Bun](https://bun.sh) 1.3+
- [Docker](https://www.docker.com/) + [Docker Compose](https://docs.docker.com/compose/)
- [goose](https://github.com/pressly/goose) for migrations

### 1. Start infrastructure

```bash
docker compose up -d
```

Starts PostgreSQL (`5437`) and Redis (`6381`).

### 2. Configure environment

```bash
cp apps/api/.env.example apps/api/.env
cp apps/worker/.env.example apps/worker/.env
cp apps/web/.env.example apps/web/.env
```

Edit secrets as needed (`JWT_SECRET`, `MCP_API_KEY_PEPPER`, `DATABASE_URL`, `REDIS_URL`).

### 3. Install deps, migrate, seed

```bash
bun install
cd apps/api && goose -dir db/migrations postgres "$DATABASE_URL" up
DATABASE_URL="..." go run ./cmd/seed
```

The `seed` command is **idempotent** — re-running it upserts the demo example
workflows (`padel-court-booking`, `order-food`, `order-doctor`) under a dedicated
demo tenant and never duplicates workflow/project rows. See
[Seed data](#seed-data) below.

### 4. Run the stack

```bash
# API (8020)
cd apps/api && go run ./cmd/server/main.go

# Worker (Redis queue)
cd apps/worker && go run ./cmd/worker/main.go

# Web (3020)
cd apps/web && bun run dev

# State MCP (8030; required API key authentication)
cd apps/api && go run ./cmd/mcp-server/main.go
```

Open:

- **State Builder**: <http://localhost:3020/state-builder>
- **Web app**: <http://localhost:3020>
- **API health**: <http://localhost:8020/health>

## Environment Variables

| Env | Default | Description |
| --- | --- | --- |
| `API_PORT` | `8020` | HTTP API + MCP server port |
| `DATABASE_URL` | — | PostgreSQL connection string |
| `JWT_SECRET` | — | JWT signing secret (API) |
| `MCP_API_KEY_PEPPER` | — | Required 32+ character verifier pepper for State MCP API keys |
| `MCP_PORT` | `8030` | State MCP endpoint port (`/mcp`) |
| `MCP_GATEWAY_MODE` | `advisory` | `advisory` preserves direct two-MCP compatibility; `secure` enforces provider calls through State MCP project bindings |

For the gateway rollout, secure-mode behavior, and rollback sequence, see
[`MCP-GATEWAY.md`](MCP-GATEWAY.md).
| `REDIS_URL` | `redis://127.0.0.1:6381` | Redis URL (worker) |
| `NEXT_PUBLIC_APP_URL` | `http://localhost:3020` | Web app base URL |
| `NEXT_PUBLIC_API_URL` | `http://localhost:8020` | Backend base URL |
| `NEXT_PUBLIC_TENANT_ID` | `00000000-0000-0000-0000-000000000001` | Default tenant id for the admin UI |

## Docker

The API ships with a Linux-ready `Dockerfile` (`apps/api/Dockerfile`) and a
separate MCP server image (`apps/api/Dockerfile.mcp`). Build and run the API:

```bash
docker build -t openstate-api -f apps/api/Dockerfile apps/api
docker run -p 8020:8020 --env-file apps/api/.env openstate-api
```

For a full local stack, `docker compose up -d` (PostgreSQL + Redis) is the
recommended path for development.

## Kubernetes (reference)

> This is a reference topology, not a full k8s operator/manifest set. It
> documents the deployment model so operators can author their own manifests.

Deploy three workloads behind an ingress:

1. **`openstate-api`** (Deployment + Service, port `8020`) — env from a Secret
   (`DATABASE_URL`, `JWT_SECRET`) + ConfigMap (`PORT`).
2. **`openstate-web`** (Deployment + Service, port `3020`) — env
   `NEXT_PUBLIC_API_URL`/`NEXT_PUBLIC_APP_URL`.
3. **`openstate-worker`** (Deployment) — same `DATABASE_URL` + `REDIS_URL`
   env, no Service.

External dependencies (PostgreSQL, Redis) are provisioned separately and exposed
via DNS or a managed service.

### Minimal example

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: openstate-api
spec:
  replicas: 2
  selector:
    matchLabels: { app: openstate-api }
  template:
    metadata:
      labels: { app: openstate-api }
    spec:
      containers:
        - name: api
          image: openstate-api:latest
          ports: [{ containerPort: 8020 }]
          env:
            - name: PORT
              value: "8020"
          envFrom:
            - secretRef: { name: openstate-db }
          livenessProbe:
            httpGet: { path: /health, port: 8020 }
```

### Migrations & seed

Run migrations and the seed as a Kubernetes **Job** before (or alongside) the
first API rollout, so the demo workflows exist when the app starts:

```bash
kubectl create job openstate-migrate --image=openstate-api:latest \
  -- goose -dir db/migrations postgres "$DATABASE_URL" up
kubectl create job openstate-seed --image=openstate-api:latest \
  -- go run ./cmd/seed
```

## Seed Data

The `apps/api/cmd/seed` command upserts the three canonical example workflows
and their projects under a fixed demo tenant (`00000000-0000-0000-0000-000000000001`):

| Intent | Workflow | Project |
| --- | --- | --- |
| `BOOKING_PADEL` | `padel-court-booking` | `padel` |
| `ORDER_FOOD` | `order-food` | `retail` |
| `ORDER_DOCTOR` | `order-doctor` | `health` |

Canonical definitions live in `docs/*.workflow.json`. Seeding is scoped to the
demo tenant so it never pollutes other tenants (PRD §4). Re-running the seed
upserts rather than duplicating workflow/project/intent rows.

The MCP routing sequence is `list_intents` → select the canonical intent key →
`resolve_intent` → `start_workflow`. For example, the seeded `padel` project
includes `BOOKING_PADEL` with examples such as `saya mau order lapangan` and
`saya mau booking lapangan padel`.

## Verification & CI

- **API**: `cd apps/api && go build ./... && go vet ./... && go test ./...`
- **Web**: `cd apps/web && bun run lint && bun run typecheck && bun run test && bun run build`
- **Golden/E2E**: covered by `go test ./...` (no real LLM, PRD 170)
- **CI**: `.github/workflows/app-ci.yml` (frontend) and
  `.github/workflows/go-ci.yml` (Go backend build/vet/test/bench).

## Rollback

All additions in this repo (seed data, tests, docs, CI) are additive and
idempotent. Removing them restores prior behavior; the seed can be dropped
without a data migration because it only upserts demo-scoped rows.
