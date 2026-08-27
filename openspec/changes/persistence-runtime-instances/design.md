## Context

Runtime execution must be the source of truth in PostgreSQL (PRD 128), with instances pinned to immutable workflow versions (PRD 58), lifecycle-managed (PRD 10, 11), concurrency-safe via optimistic locking (PRD 31), and tenant-isolated (PRD 4, 96). This change adds the instance/state persistence slice of epic #3, building on the definition schema.

## Goals / Non-Goals

**Goals:**
- Schema for `workflow_instances` and `state_instances`.
- Domain entities + `IInstanceRepository`.
- pgx adapter backed by sqlc.
- Optimistic locking and tenant scoping on every query.

**Non-Goals:**
- The state-machine / transition engine (separate epic).
- Event system, context/memory, capability, audit persistence (separate changes).
- Any HTTP layer.

## Decisions

### D1: Workflow instance lifecycle as VARCHAR + Go constants
`workflow_instances.status` uses VARCHAR + Go typed constants mirroring PRD 10:
`CREATED`, `RUNNING`, `WAITING`, `COMPLETED`, `CANCELLED`, `FAILED`, `EXPIRED`, `ABORTED`, plus `SUSPENDED` for interruption (PRD 42-43).

### D2: State instance lifecycle as VARCHAR + Go constants
`state_instances.status` mirrors PRD 11: `ENTERING`, `ACTIVE`, `WAITING`, `EXITING`, `COMPLETED`, `FAILED`, `EXPIRED`, `CANCELLED`.

### D3: Instance pinned to immutable version
`workflow_instances.workflow_version_id` references `workflow_versions(id)` (PRD 58). Behavior is reproducible: the instance always executes against the version it was created from, even after newer versions publish (PRD 55, 56).

### D4: Optimistic locking at instance and state level
Both tables have a `version INT NOT NULL DEFAULT 0`. All mutations run `WHERE ... AND version = $n` and `SET version = version + 1` (PRD 31). A 0-row result is a conflict → DomainError `CONFLICT`. This is the primary concurrency mechanism (PRD 130).

### D5: State version counter and retry persistence
`state_instances.retry_count INT NOT NULL DEFAULT 0` persists retry state across restarts (PRD 48). `entered_at`, `expires_at` support timeout scheduling (PRD 25, 3.6); `exited_at` for lifecycle.

### D6: Version pinning + parent-child tenancy
`workflow_instances.tenant_id` required; `state_instances.tenant_id` denormalized for scoped reads. `state_instances.workflow_instance_id` FK to the parent; `current_state_instance_id` on the workflow instance points to the active state instance for fast "current state" resolution (PRD 7).

### D7: Transactional atomic transitions
Repository exposes a transition-scoped operation that, within one transaction, updates the current state instance and the workflow instance version counter together (PRD 69: "State transitions must atomically persist ... version increment"). This is the adapter-level unit; the engine decides *what* to transition, the adapter persists it atomically.

## Schema Outline

```
workflow_instances
  id               UUID PK
  tenant_id        UUID NOT NULL
  workflow_id      UUID NOT NULL REFERENCES workflows(id)
  workflow_version_id UUID NOT NULL REFERENCES workflow_versions(id)   -- PRD 58
  correlation_key  VARCHAR                         -- business/conversation correlation (PRD 6)
  status           VARCHAR NOT NULL DEFAULT 'CREATED'
  version          INT NOT NULL DEFAULT 0          -- optimistic lock (PRD 31)
  current_state_instance_id UUID REFERENCES state_instances(id)
  started_at       TIMESTAMP
  completed_at     TIMESTAMP
  expires_at       TIMESTAMP                        -- workflow timeout (PRD 26)
  created_at / updated_at
  INDEX (tenant_id, status), INDEX (tenant_id, correlation_key)

state_instances
  id               UUID PK
  tenant_id        UUID NOT NULL
  workflow_instance_id UUID NOT NULL REFERENCES workflow_instances(id) ON DELETE CASCADE
  workflow_version_id UUID NOT NULL REFERENCES workflow_versions(id)
  state_key        VARCHAR NOT NULL                -- reference to states.key
  state_id         UUID REFERENCES states(id)
  status           VARCHAR NOT NULL DEFAULT 'ENTERING'
  version          INT NOT NULL DEFAULT 0          -- optimistic lock (PRD 31)
  retry_count      INT NOT NULL DEFAULT 0          -- PRD 48
  entered_at       TIMESTAMP NOT NULL DEFAULT NOW()
  expires_at       TIMESTAMP                        -- state timeout (PRD 25)
  exited_at        TIMESTAMP
  created_at / updated_at
  INDEX (workflow_instance_id), INDEX (tenant_id, status)
```

Note: `current_state_instance_id` circular FK to `state_instances` is resolved by adding the FK after `state_instances` is created (PostgreSQL allows `ALTER TABLE ADD CONSTRAINT` post-create), and is nullable.

## Risks / Trade-offs

- **Risk: circular FK `workflow_instances ↔ state_instances`** → Mitigation: `current_state_instance_id` is nullable and added via `ALTER TABLE` after both tables exist; no INSERT-time ordering problem.
- **Risk: optimistic-lock retries from engine** → Mitigation: expected; adapter returns `CONFLICT`, engine re-reads latest version (handled in a later engine change).
- **Risk: denormalized tenant on state_instances** → Mitigation: intentional for indexed, join-free tenant-scoped reads (PRD 96).

## Migration Plan

1. Branch `feature/epic3-persistence-runtime-instances`.
2. Add migration `00003_workflow_instances.sql` (Up + Down, incl. ALTER for circular FK).
3. Add `db/queries/instance.sql`.
4. Run `sqlc generate`.
5. Add entities + `IInstanceRepository`.
6. Implement `pgx_instance_repository.go` (incl. transactional transition op).
7. `go build ./...`, `go vet ./...`, `go test ./...`.
8. Optional smoke: `goose up`, create instance + state instances, update with optimistic lock.
9. PR → review → merge.

**Rollback**: migration `Down` drops tables (and removes the circular FK first); feature branch isolates changes.

## Open Questions

- Whether `current_state_instance_id` should be a hard FK or a plain UUID pointer. Decision: nullable FK for referential integrity; revisit if the circular reference complicates adapter transactions.
