# OpenState — Enterprise Conversation State Orchestration Platform

> **The LLM can suggest. The State Engine decides. The MCP executes. The RAG informs. PostgreSQL remembers.**

<div align="center">

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](./LICENSE)
[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white)](https://golang.org)
[![Next.js](https://img.shields.io/badge/Next.js-16-000000?logo=next.js&logoColor=white)](https://nextjs.org)
[![TypeScript](https://img.shields.io/badge/TypeScript-blue?logo=typescript&logoColor=white)](https://www.typescriptlang.org)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](./CONTRIBUTING.md)

**Multi-tenant · Versioned workflow engine · LLM-driven conversations · MCP + RAG integrated**

![OpenState Banner](./docs/assets/og-image.svg)

</div>

OpenState is an open-source, enterprise-grade platform for defining, publishing,
executing, observing, and versioning **conversation state** and **business
workflows**. It sits between the user's conversation and external AI/tooling
systems, giving the LLM authoritative runtime context while keeping the LLM
**out of the loop** when it comes to workflow state.

## Why OpenState

LLMs are great at understanding language, inferring intent, and generating
responses — but they must **never** become the source of truth for workflow
state. OpenState gives you a deterministic state engine that controls:

- Which business process the conversation is currently executing
- What state that process is in
- What context is available / required
- What capabilities are allowed
- What transitions are valid

```
User
  ↓
LLM
  ↓
Conversation Orchestrator
  ├── Workflow Resolver
  ├── State Engine
  ├── Context Engine
  ├── Policy Engine
  ├── Event Engine
  └── Capability Resolver
        ├── MCP
        └── RAG
  ↓
LLM
  ↓
User
```

RAG and MCP are **not owned** by this platform — they are integrated through
well-defined contracts.

## Core Principle

> **The LLM can suggest. The State Engine decides. The MCP executes. The RAG informs. PostgreSQL remembers.**

| Component          | Responsibility                                                        |
| ------------------ | --------------------------------------------------------------------- |
| LLM                | Understand language, infer intent, generate responses                 |
| State Orchestrator | Authoritative workflow/state decision                                  |
| State Builder      | Define workflows visually                                             |
| RAG                | Retrieve knowledge                                                     |
| MCP                | Execute external capabilities/tools                                   |
| PostgreSQL         | Persistent source of truth                                            |

## Features

- **Multi-tenant & multi-project** — isolation at API, service, repository,
  cache, event, and capability layers. Domain hierarchy:
  `Perusahaan (Tenant) → Project → Intent → Workflow → State`
- **Multiple workflows per tenant** — unlimited workflows across projects
  (ORDER, BOOKING, CONSULTATION, etc.)
- **Workflow versioning** — immutable published versions, rollback for new
  instances
- **State Builder** — visual drag-and-drop flowchart editor (React Flow)
- **Intent Registry** — conversation → intent → workflow → state resolution,
  scoped per project
- **Deterministic guards** — JSON-based, no arbitrary code execution
- **Events & transitions** — with priority, guards, and concurrency control
- **State & workflow timeouts** — processed through the normal event pipeline
- **Idempotency & optimistic concurrency** — safe external event processing
- **Capability Registry** — logical capabilities mapped to MCP providers
- **LLM context compilation** — minimal context per turn, PII redaction
- **Suspension / resume** — mid-flow interruption without losing context
- **Audit trail & replay** — full operational history
- **Simulation & validation** — test workflows before publishing
- **Import / Export** — versioned JSON, git-friendly

## Architecture

OpenState is a **modular monolith** (Go) with clean internal boundaries
(domain / application / infrastructure / interfaces), per PRD section 175.
It scales horizontally when operational requirements justify it.

```
apps/
  web/           Next.js + React Flow (State Builder, Admin Console)
  api/           Go (Echo, Clean Architecture) + sqlc + goose + pgx
  worker/        Go (asynq) + Redis (timeout, outbox, delayed events)
packages/
  go-shared/     Shared Go module
  schemas/       Shared validation schemas
  types/         Shared TypeScript types
  utils/         Shared utilities
```

### Domain Hierarchy

```
Perusahaan (Tenant)
  └── Project (business area: resto | padel | dokter)
        └── Intent
              └── Workflow (state machine)
                    └── State
```

- **Tenant** owns many **Projects**
- **Project** owns many **Intents**
- **Intent** resolves to a **Workflow** within the same project
- **Workflow** contains many **States**
- All persistence & runtime operations are scoped by `tenant_id` + `project_id`

### Persistence

PostgreSQL is the primary persistent database (the source of truth). The
engine talks to **repository interfaces** — the Postgres adapter is the primary
implementation, and MySQL/SQLite/MongoDB adapters can be added later without
rewriting the engine. See [ADR-001](./docs/adr/ADR-001-persistence.md).

## Quick Start

### Prerequisites

- [Go](https://golang.org/dl/) 1.26+
- [Bun](https://bun.sh) 1.3+
- [Docker](https://www.docker.com/) (for local PostgreSQL + Redis)
- [goose](https://github.com/pressly/goose) (migrations)

### 1. Start infrastructure (PostgreSQL + Redis)

```bash
docker compose up -d
```

> Default ports: PostgreSQL `5437`, Redis `6381` (adjusted to avoid conflicts
> with common local setups). See `docker-compose.yml`.

### 2. Configure environment

```bash
cp apps/api/.env.example apps/api/.env
cp apps/worker/.env.example apps/worker/.env
cp apps/web/.env.example apps/web/.env
```

Edit the `.env` files with your own secrets.

### 3. Install dependencies, run migrations & seed

```bash
bun install
cd apps/api && goose -dir db/migrations postgres "$DATABASE_URL" up
DATABASE_URL="..." go run ./cmd/seed
```

The seed (idempotent) registers the demo example workflows under a dedicated
demo tenant so they can be resolved and executed end-to-end. See
[`docs/DEPLOYMENT.md`](./docs/DEPLOYMENT.md#seed-data).

### 4. Run the stack

```bash
# API (port 8020)
cd apps/api && go run ./cmd/server/main.go

# Worker (Redis queue)
cd apps/worker && go run ./cmd/worker/main.go

# Web (port 3020)
cd apps/web && bun run dev

# State MCP (port 8030, required Bearer API key)
cd apps/api && go run ./cmd/mcp-server/main.go
```

Open:

- **State Builder**: <http://localhost:3020/state-builder>
- **Web app**: <http://localhost:3020>
- **API health**: <http://localhost:8020/health>

## Configuration

| Env | Default | Description |
| --- | --- | --- |
| `API_PORT` | `8020` | HTTP API port |
| `DATABASE_URL` | — | PostgreSQL connection string |
| `JWT_SECRET` | — | JWT signing secret (API) |
| `MCP_API_KEY_PEPPER` | — | 32+ character server secret used to verify State MCP API keys |
| `MCP_PORT` | `8030` | State MCP HTTP port (`/mcp`) |
| `MCP_PROVIDER_MOCK_PORT` | `8031` | Development provider MCP mock port (`/mcp`) |
| `REDIS_URL` | `redis://127.0.0.1:6381` | Redis URL (worker) |
| `NEXT_PUBLIC_APP_URL` | `http://localhost:3020` | Web app base URL |
| `NEXT_PUBLIC_API_URL` | `http://localhost:8020` | Backend base URL |

## State MCP authentication

State MCP runs separately at `http://localhost:8030/mcp`. The development
provider mock runs at `http://localhost:8031/mcp`; the LLM host connects to both
servers as separate MCP sessions. It requires
`Authorization: Bearer osk_...`; tenant identity comes from that API key, not a
tool argument. Create a key using an authenticated admin session and the target
tenant header:

```bash
curl -X POST http://localhost:8020/api/api-keys \
  -H "Authorization: Bearer <admin-jwt>" \
  -H "X-Tenant-ID: <tenant-id>" \
  -H "Content-Type: application/json" \
  -d '{"name":"local Claude","projectIds":["<project-id>"],"defaultProjectId":"<project-id>","scopes":["state:read","state:write","capability:invoke"]}'
```

Copy `data.key` immediately; it is never returned again. Configure the MCP
client with that value as a Bearer token. The `project` tool argument is
optional only when the key has a default project, and it must be in the key's
allowlist.

Provider alias configuration belongs to the LLM/MCP host, not workflow JSON:

```json
{
  "openstate": { "url": "http://localhost:8030/mcp", "authorization": "Bearer osk_..." },
  "padel-provider-mock": { "url": "http://localhost:8031/mcp" }
}
```

When State MCP returns `providerServer` and `providerTool`, call that exact
tool on the already-connected alias, then report the normalized result with
`report_capability_result` before calling `propose_event`. Run
`bun run mcp:check` to detect a wrong or stale listener on either port.

## Developer Tooling

The repository ships with configuration for the following agent tooling. They are
**developer tools**, not runtime dependencies — install them in your own
environment as needed.

### OpenSpec (spec-driven planning)

OpenSpec is used for lightweight, spec-driven feature planning. Configuration
lives in [`openspec/`](./openspec).

```bash
npm install -g @fission-ai/openspec@latest
openspec validate
```

- Specs: `openspec/specs/`
- Proposed changes: `openspec/changes/`
- Skills: `.agents/skills/openspec-*`

### Serena (semantic code retrieval / editing)

Serena is an MCP toolkit that gives the agent IDE-level symbol understanding
of the codebase. Configuration lives in [`.serena/`](./.serena).

```bash
# prerequisite: uv
uv tool install -p 3.13 serena-agent
serena init
```

Connect `serena` to your MCP client (Claude Code, Codex, OpenCode, etc.) per
the [Serena docs](https://github.com/oraios/serena).

### Graphify (knowledge graph)

Graphify builds a queryable knowledge graph of the codebase (code + docs +
diagrams) for the agent. Configuration lives in [`.graphify/`](./.graphify).

```bash
# install via uv (provides `graphify` + `graphify-mcp` binaries)
uv tool install graphifyy

# verify
graphify --help

# run the local stdio MCP server
graphify-mcp serve --transport stdio
```

> Set your AI model API key (e.g. `OPENAI_API_KEY` / `ANTHROPIC_API_KEY`) for
> semantic extraction. Graphify only sends semantic descriptions, never raw source.

Docs: <https://graphify.net/>

## MCP Integration

OpenState's **output** is an **MCP server (HTTP/SSE)** that a **3rd-party LLM /
RAG** (owned by the customer) calls to query and drive workflow state:

```text
3rd-party LLM (customer-owned)
    ↓ invokes MCP tools
MCP Server (OpenState output)
    ↓
State Orchestrator
```

The platform does not call an LLM internally — intent classification, entity
extraction, and response generation are done by the 3rd-party LLM via these
tools:

- `list_intents` — list canonical intents, descriptions, and example utterances for a tenant/project
- `resolve_intent` — resolve the selected canonical intent → workflow
- `get_active_workflow` — active workflow + current state + allowed events
- `get_context` — available + missing context (PII-redacted)
- `get_allowed_capabilities` — authorized capabilities per state
- `propose_event` — LLM suggests, engine validates & transitions
- `invoke_capability` — authorized capability execution
- `start_workflow`, `suspend_workflow`, `resume_workflow`, `cancel_workflow`
- `get_workflow_instances`, `get_history`, `replay_workflow`

### Intent routing flow

The LLM should discover the available choices before resolving a user request:

1. Call `list_intents` with the tenant and project.
2. Compare the user's message with each intent's description and examples.
3. For a message such as `saya mau order lapangan`, select `BOOKING_PADEL`.
4. Call `resolve_intent` with `BOOKING_PADEL`.
5. Use the returned workflow with `start_workflow`, then continue through
   `get_context` and `propose_event`.

OpenState does not classify text internally. The LLM suggests the canonical
intent, while the tenant/project-scoped catalog and State Engine validate the
mapping and execution path.

> The full MCP tool contract is tracked in the GitHub issue
> "[MCP & Integration] Server MCP + capability".

## Example Workflows

- **PADEL Booking** — availability check, overlap recommendation, 50% down payment
- **Order Food** — product selection, stock check, recommendation, payment, mid-flow product change
- **Order Doctor** — schedule/queue check, needs-based recommendation, doctor switch

Load them in the State Builder via the "Examples" dropdown.

## Quality & Testing

The engine and integration surface are covered by deterministic tests that run
in CI **without a real LLM** (PRD 170):

- **Deterministic runtime tests** — every guard operator, AND/OR grouping,
  priority ordering, and rejection, driven directly on the engine (PRD 126).
- **Golden conversation tests** — per-workflow fixtures replaying user turns and
  asserting the resolved state, as AI-behavior regression (PRD 125).
- **End-to-end test** — a deterministic LLM/MCP mock drives the real MCP tool
  handlers (`resolve_intent`, `start_workflow`, `propose_event`) through the
  engine to a persisted state transition.
- **Load test** — in-memory state-transition throughput baseline with a loose
  lower bound, plus a Go benchmark (`go test -bench=BenchmarkProcessEvent`).

Run them:

```bash
# Go backend (unit + golden + deterministic + E2E + load)
cd apps/api && go test ./...

# throughput benchmark
cd apps/api && go test ./internal/domain/engine -bench=BenchmarkProcessEvent -benchtime=1s

# frontend
cd apps/web && bun run test
```

CI gates both toolchains: `.github/workflows/app-ci.yml` (frontend) and
`.github/workflows/go-ci.yml` (Go build/vet/test/bench).

## Documentation

- **PRD**: [`MAIN_PRD.md`](./MAIN_PRD.md) — the product definition & engineering
  specification (source of truth)
- **Architecture Decision Records**: [`docs/adr/`](./docs/adr/)
- **Deployment**: [`docs/DEPLOYMENT.md`](./docs/DEPLOYMENT.md) — local, Docker, and Kubernetes
- **Operations**: [`docs/OPERATION.md`](./docs/OPERATION.md)
- **Changelog**: [`CHANGELOG.md`](./CHANGELOG.md)
- **Contributing**: [`CONTRIBUTING.md`](./CONTRIBUTING.md)
- **Security**: [`SECURITY.md`](./SECURITY.md)
- **Code of Conduct**: [`CODE_OF_CONDUCT.md`](./CODE_OF_CONDUCT.md)

## Roadmap

Tracked as GitHub issues (per PRD phases):

1. **Runtime Engine** — state machine, guard evaluation, context store
2. **Data & Persistence** — PostgreSQL schema + repository abstraction
3. **MCP & Integration** — MCP server, capability execution, RAG integration
4. **Frontend** — State Builder production + Admin Console
5. **Security & Ops** — multi-tenant, RBAC, audit, observability, deployment
6. **Quality** — example workflows, testing, documentation

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Follow the coding conventions (Biome for frontend, `go vet` for backend)
4. Open a pull request

See [`MAIN_PRD.md`](./MAIN_PRD.md) and the GitHub issues for the roadmap.

## License

[MIT](./LICENSE)
