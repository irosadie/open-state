## Why

Epic **#3 (Data & Persistence)** requires runtime context and persistent memory to survive restarts and be queryable (PRD 128). This is the fourth persistence slice: **context records and memory references** (PRD 23-24, 43.2).

The platform distinguishes *persistent memory* (customer/user domain, survives workflow expiry, PRD 24) from *workflow data* (state data tied to an instance, PRD 24, 131). `context_records` stores key/value runtime context scoped to an entity (tenant/conversation/workflow instance/state instance), while `memory_references` models persistent memory with explicit references, so deleting workflow state never deletes user memory (PRD 24, 43.2).

## What Changes

- **NEW** — `apps/api/db/migrations/00005_context.sql` goose migration creating:
  - `context_records` — scoped runtime key/value context (PRD 23, 36, 43.2) with typed values and versioning.
  - `memory_references` — persistent memory model with references (PRD 23-24, 43.2).
- **NEW** — `apps/api/db/queries/context.sql` sqlc query file.
- **NEW** — Domain entities `ContextRecord`, `MemoryReference` in `apps/api/internal/domain/entities/`.
- **NEW** — Repository interface `apps/api/internal/domain/repositories/context_repository.go` (`IContextRepository`).
- **NEW** — pgx adapter `apps/api/internal/infrastructure/database/pgx_context_repository.go`.
- **REGENERATED** — sqlc-generated Go under `apps/api/internal/infrastructure/db/`.
- Uses **`db-sqlc-schema`** and **`api-feature`** skills. `docs-openapi` not touched (no public endpoint).

## Capabilities

### New Capabilities

- `backend/persistence/context-memory`: tenant-isolated scoped context records and persistent memory references behind `IContextRepository` and a pgx adapter.

### Modified Capabilities

- None (references `workflow_instances`/`state_instances` FK from `persistence-runtime-instances`).

## Impact

- **`apps/api/db/migrations/`** — add `00005_context.sql`.
- **`apps/api/db/queries/`** — add `context.sql`.
- **`apps/api/internal/infrastructure/db/`** — sqlc-generated code regenerated.
- **`apps/api/internal/domain/entities/`** — add `context_record.go`, `memory_reference.go`.
- **`apps/api/internal/domain/repositories/`** — add `context_repository.go`.
- **`apps/api/internal/infrastructure/database/`** — add `pgx_context_repository.go`.
- **No** changes to web, worker, shared packages, OpenAPI, docker.

## Non-Goals

- The Context Engine (resolution hierarchy, LLM context compilation, PRD 22-23) — separate epic.
- RAG persistence/embeddings (PRD 19) — not owned by this platform.
- Event, capability, audit persistence — separate changes.
- Any HTTP layer.

## Dependencies

- `persistence-runtime-instances` (FK targets `workflow_instances`, `state_instances`).
- Epic #3.
