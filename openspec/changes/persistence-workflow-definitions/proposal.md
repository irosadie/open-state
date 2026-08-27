## Why

Epic **#3 (Data & Persistence)** requires PostgreSQL to become the source of truth for workflow *definitions* (PRD 128: no data lost on restart/crash). Today the platform has only auth tables (`00001_init_auth.sql`); there is no persistence for workflows. Per **ADR-001**, the engine must talk to a **repository interface** while PostgreSQL is the primary adapter. This change delivers the workflow-definition persistence slice of that epic: schema, domain entities, repository interface, sqlc queries, and the pgx adapter — all following the established `db-sqlc-schema` + `api-feature` patterns (users/auth already use).

Workflows are *definitions* (authoring artifacts), distinct from *runtime instances* (handled in `persistence-runtime-instances`). Definitions are versioned and published-immutable (PRD 3.3, 9, 55, 56), so the schema must preserve that model.

## What Changes

- **NEW** — `apps/api/db/migrations/00002_workflows.sql` goose migration creating:
  - `projects` — tenant-scoped business area (PRD 3.1.1); slug unique per tenant.
  - `workflows` — project-scoped definition root (PRD 4, 96), `slug` unique per
    (tenant, project) (PRD 5, 3.1.1), lifecycle `status`, optimistic `version` (PRD 31).
  - `workflow_versions` — immutable published snapshots (PRD 3.3, 9, 55); stores the full `WorkflowDefinition` as `JSONB` (PRD 68).
  - `states` — relational, queryable view of definition nodes (PRD 12, 14), immutable per version.
  - `transitions` — relational transitions (PRD 33, 34), immutable per version.
  - `transition_guards` — relational guard records per transition (PRD 35).
- **NEW** — `apps/api/db/queries/workflow.sql` sqlc query file for the above tables.
- **NEW** — Domain entities in `apps/api/internal/domain/entities/`: `Workflow`, `WorkflowVersion`, `State`, `Transition`, `TransitionGuard`.
- **NEW** — Repository interface `apps/api/internal/domain/repositories/workflow_repository.go` (`IWorkflowRepository`).
- **NEW** — pgx adapter `apps/api/internal/infrastructure/database/pgx_workflow_repository.go`.
- **REGENERATED** — sqlc-generated Go under `apps/api/internal/infrastructure/db/` (never hand-edited).
- Uses **`db-sqlc-schema`** skill (migrations + queries + regen) and **`api-feature`** skill (entity + repository + adapter layers).

No HTTP endpoints are added in this change — this is a persistence-contract change only, so `docs-openapi` is intentionally **not** touched (no public API surface changes). The API feature skill is used for the repository/entity portion; controller/route layers are out of scope until runtime features exist.

## Capabilities

### New Capabilities

- `backend/persistence/workflow-definitions`: persistent storage for workflow definitions and their immutable versions, exposed through `IWorkflowRepository` behind a pgx/PostgreSQL adapter.

### Modified Capabilities

- None.

## Impact

- **`apps/api/db/migrations/`** — add `00002_workflows.sql` (sequential filename, goose Up/Down blocks).
- **`apps/api/db/queries/`** — add `workflow.sql`.
- **`apps/api/internal/infrastructure/db/`** — sqlc-generated code regenerated (do not hand-edit).
- **`apps/api/internal/domain/entities/`** — add `workflow.go`, `workflow_version.go`, `state.go`, `transition.go`, `transition_guard.go`.
- **`apps/api/internal/domain/repositories/`** — add `workflow_repository.go`.
- **`apps/api/internal/infrastructure/database/`** — add `pgx_workflow_repository.go`.
- **No** changes to `apps/web`, `apps/worker`, shared packages, OpenAPI, or docker.

## Non-Goals

- Runtime instance persistence (`workflow_instances`, `state_instances`) — separate change.
- Event system, context/memory, capabilities/policies, audit persistence — separate changes.
- Any HTTP controller/route/use-case for workflow management — out of scope here.
- Cross-tenant authorization middleware — repository queries are tenant-scoped by design (PRD 4/96) but auth enforcement is out of scope.
- Frontend (`PGlite` draft persistence, PRD 75.1) — tracked under the Frontend epic (#5).

## Dependencies

- Requires completed Go backend + `packages/go-shared` DomainError (epic #1 / `migrate-backend-to-golang`, already merged).
- Epic #3 as a whole; this is the first slice.
