## Context

The platform must persist workflow *definitions* as the source of truth (PRD 128), with a versioned, publish-immutable lifecycle (PRD 9, 55, 56) and tenant isolation (PRD 4, 96). Per ADR-001, the engine depends on a repository interface; PostgreSQL is the primary adapter. This change is the first persistence slice of epic #3.

## Goals / Non-Goals

**Goals:**
- Create production-grade PostgreSQL schema for workflow definitions, versions, states, transitions, and guards.
- Define domain entities + `IWorkflowRepository` interface.
- Implement the pgx adapter backed by sqlc-generated queries.
- Follow `db-sqlc-schema` (migration → query → regen) and `api-feature` (entity → repository → adapter) conventions already established by auth.

**Non-Goals:**
- Runtime instance persistence (separate change).
- Any HTTP layer (controller/route/use-case/service/DTO).
- Event, context, capability, audit persistence (separate changes).
- Frontend changes.

## Decisions

### D1: Relational tables + full definition JSONB snapshot
The epic and PRD 68 name concrete tables (`workflows`, `workflow_versions`, `states`, `transitions`, `transition_guards`). Each `workflow_versions.definition` also stores the complete `WorkflowDefinition` envelope as `JSONB` (same shape used by the builder, PRD 75.1). Relational rows give queryable access; JSONB preserves the authoritative authoring artifact and future portability to other adapters. Both are written atomically in the same transaction.

### D2: Workflow lifecycle as VARCHAR + Go typed constants
Per `db-sqlc-schema` prohibition (no PostgreSQL ENUM) and `api-feature` (typed Go constants), `workflows.status` and `workflow_versions.status` are `VARCHAR` columns validated via Go typed constants:
- `workflow_status`: `DRAFT`, `VALIDATING`, `VALID`, `PUBLISHED`, `ARCHIVED` (PRD 9).

### D3: Versioned, immutable published versions
`workflow_versions.version_no` is unique per `workflow_id` (PRD 3.3). A published version's rows are immutable: `states`, `transitions`, `transition_guards` reference `workflow_version_id` and are never mutated after publish; editing creates a new draft version (PRD 55, 56). `is_current` marks the active published version used by new instances (PRD 58).

### D4: Optimistic locking on the definition root
`workflows.version` (integer, default 0) enables optimistic concurrency (PRD 31). Every update runs `... WHERE id=$n AND version=$old` and bumps `version=version+1`; a 0-row result signals a conflict. Instance-level locking is a later change.

### D5: Tenant isolation via composite unique + scoped queries
Every definition query filters by `tenant_id` (PRD 4, 96). `workflows` has `UNIQUE (tenant_id, slug)` (PRD 5: slug unique inside tenant). `IWorkflowRepository` methods take an explicit `tenantID string` parameter so cross-tenant reads are impossible at the data-access layer.

### D6: Entity identifiers as strings
Following the auth pattern (`User.ID string`), entity IDs are `string` UUIDs; the adapter parses/serializes to `uuid.UUID` via pgx/sqlc. This keeps the domain DB-agnostic (ADR-001).

## Schema Outline

```
workflows
  id            UUID PK
  tenant_id     UUID NOT NULL
  slug          VARCHAR NOT NULL
  name          VARCHAR NOT NULL
  description   TEXT
  status        VARCHAR NOT NULL DEFAULT 'DRAFT'
  current_version INT NOT NULL DEFAULT 0
  version       INT NOT NULL DEFAULT 0           -- optimistic lock (PRD 31)
  created_at / updated_at
  UNIQUE(tenant_id, slug)

workflow_versions
  id            UUID PK
  workflow_id   UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE
  tenant_id     UUID NOT NULL                     -- denormalized for scoped reads
  version_no    INT NOT NULL
  definition    JSONB NOT NULL                    -- full WorkflowDefinition
  status        VARCHAR NOT NULL DEFAULT 'DRAFT'
  is_current    BOOLEAN NOT NULL DEFAULT FALSE
  created_at / updated_at
  UNIQUE(workflow_id, version_no)

states
  id            UUID PK
  workflow_version_id UUID NOT NULL REFERENCES workflow_versions(id) ON DELETE CASCADE
  key           VARCHAR NOT NULL                  -- stable node key (e.g. PAYMENT)
  kind          VARCHAR NOT NULL                  -- START/STATE/DECISION/WAIT/END/EVENT
  name          VARCHAR NOT NULL
  description   TEXT
  instructions  TEXT
  required_context JSONB NOT NULL DEFAULT '[]'
  capabilities  JSONB NOT NULL DEFAULT '[]'
  policy        JSONB NOT NULL DEFAULT '{}'
  is_terminal   BOOLEAN NOT NULL DEFAULT FALSE
  position      JSONB NOT NULL DEFAULT '{}'        -- x/y for builder
  created_at / updated_at
  UNIQUE(workflow_version_id, key)

transitions
  id            UUID PK
  workflow_version_id UUID NOT NULL REFERENCES workflow_versions(id) ON DELETE CASCADE
  key           VARCHAR NOT NULL                  -- stable transition id
  source_state_id UUID NOT NULL REFERENCES states(id) ON DELETE CASCADE
  target_state_id UUID NOT NULL REFERENCES states(id) ON DELETE CASCADE
  event         VARCHAR NOT NULL
  priority      INT NOT NULL DEFAULT 1            -- PRD 34: lower = evaluated first
  is_active     BOOLEAN NOT NULL DEFAULT TRUE
  created_at / updated_at
  UNIQUE(workflow_version_id, key)

transition_guards
  id            UUID PK
  transition_id UUID NOT NULL REFERENCES transitions(id) ON DELETE CASCADE
  workflow_version_id UUID NOT NULL REFERENCES workflow_versions(id) ON DELETE CASCADE
  logic         VARCHAR NOT NULL DEFAULT 'AND'    -- AND/OR
  conditions    JSONB NOT NULL DEFAULT '[]'
  created_at / updated_at
```

Indexes: FK columns (`tenant_id` on every table; `workflow_id` on versions; `workflow_version_id` on states/transitions; `transition_id` on guards), plus `workflows(tenant_id, slug)` unique and `workflow_versions(workflow_id, version_no)` unique.

## Risks / Trade-offs

- **Risk: JSONB definition vs relational rows drift** → Mitigation: both written atomically in one transaction; relational rows are the queryable projection, JSONB the authoritative snapshot. Seed data lives in migrations/seed and is reviewed.
- **Risk: Denormalized `tenant_id` on child tables** → Mitigation: intentional for scoped, indexed tenant reads (PRD 96) and to avoid join-per-query; kept consistent via application invariants (child rows always inherit the workflow's tenant).
- **Risk: Optimistic lock conflicts under load** → Mitigation: expected for authoring; clients retry with the latest version. Instance-level concurrency is a separate change (PRD 31).

## Migration Plan

1. Branch `feature/epic3-persistence-workflow-definitions`.
2. Add migration `00002_workflows.sql` (Up + Down).
3. Add `db/queries/workflow.sql`.
4. Run `sqlc generate` in `apps/api`.
5. Add domain entities and `IWorkflowRepository`.
6. Implement `pgx_workflow_repository.go`.
7. `go build ./...`, `go vet ./...`, `go test ./...`.
8. Optional smoke: `goose up`, insert a workflow, query it.
9. PR → review → merge.

**Rollback**: migration `Down` drops the tables; feature branch isolates changes.

## Open Questions

- Whether `states`/`transitions` relational tables should be kept for the first slice or derived purely from JSONB. Decision: keep them relational now (PRD 68 "responsibilities separable"); revisit if unused by the runtime engine after instance persistence lands.
