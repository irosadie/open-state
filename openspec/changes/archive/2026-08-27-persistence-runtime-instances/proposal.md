## Why

Epic **#3 (Data & Persistence)** requires runtime workflow execution to be persisted so no execution disappears on restart/crash (PRD 128) and history can be replayed (PRD 51-52). This change is the second persistence slice: **runtime instances and their states**. It depends on the definition schema from `persistence-workflow-definitions` (workflows / workflow_versions / states / transitions).

A `workflow_instance` is an executing copy of a published workflow version, pinned to a `workflow_id` + `workflow_version_id` for reproducibility (PRD 58). A `state_instance` is the runtime occurrence of a state inside that instance (PRD 3.6, 11). Both carry optimistic version counters (PRD 31) and are scoped by **tenant + project** (PRD 4, 96, 3.1.1).

## What Changes

- **NEW** — `apps/api/db/migrations/00003_workflow_instances.sql` goose migration creating:
  - `workflow_instances` — runtime execution root (PRD 10, 58); scoped by `tenant_id` + `project_id`; lifecycle `status` (PRD 10), optimistic `version` (PRD 31), pinned `workflow_version_id` (PRD 58), suspension/interruption support (PRD 42-43).
  - `state_instances` — runtime occurrence of a state (PRD 11); lifecycle `status`, optimistic `version`, entered/expired timestamps (PRD 3.6, 25), retry counter persisted (PRD 48).
- **NEW** — `apps/api/db/queries/instance.sql` sqlc query file.
- **NEW** — Domain entities `WorkflowInstance`, `StateInstance` in `apps/api/internal/domain/entities/`.
- **NEW** — Repository interface `apps/api/internal/domain/repositories/instance_repository.go` (`IInstanceRepository`).
- **NEW** — pgx adapter `apps/api/internal/infrastructure/database/pgx_instance_repository.go`.
- **REGENERATED** — sqlc-generated Go under `apps/api/internal/infrastructure/db/`.
- Uses **`db-sqlc-schema`** skill (migration + queries + regen) and **`api-feature`** skill (entity + repository + adapter). `docs-openapi` is **not** touched (no public endpoint).

## Capabilities

### New Capabilities

- `backend/persistence/workflow-instances`: persistent, tenant-isolated storage of running workflow instances and state instances with optimistic locking, exposed through `IInstanceRepository`.

### Modified Capabilities

- `backend/persistence/workflow-definitions`: depends on the `workflow_versions` FK defined there.

## Impact

- **`apps/api/db/migrations/`** — add `00003_workflow_instances.sql`.
- **`apps/api/db/queries/`** — add `instance.sql`.
- **`apps/api/internal/infrastructure/db/`** — sqlc-generated code regenerated.
- **`apps/api/internal/domain/entities/`** — add `workflow_instance.go`, `state_instance.go`.
- **`apps/api/internal/domain/repositories/`** — add `instance_repository.go`.
- **`apps/api/internal/infrastructure/database/`** — add `pgx_instance_repository.go`.
- **No** changes to web, worker, shared packages, OpenAPI, docker.

## Non-Goals

- The runtime engine itself (state machine, guard evaluation, transition execution) — other epic.
- Event system persistence (`events`, inbox/outbox, idempotency) — separate change (`persistence-event-system`).
- Context/memory persistence — separate change.
- HTTP controllers/routes for instances.
- Worker/scheduler integration.

## Dependencies

- `persistence-workflow-definitions` (tables `workflows`, `workflow_versions`) must be merged first (FK target).
- Epic #3.
