## Context

Runtime context and persistent memory are distinct concerns (PRD 23-24, 43.2). Context is scoped and versioned; memory is long-lived user/customer data that must survive workflow expiry. This change adds the context/memory persistence slice of epic #3.

## Goals / Non-Goals

**Goals:**
- Schema for `context_records` and `memory_references`.
- Domain entities + `IContextRepository`.
- pgx adapter backed by sqlc.
- Tenant scoping + context versioning + memory/workflow separation.

**Non-Goals:**
- The Context Engine resolution logic (separate epic).
- RAG/embeddings (not owned by the platform).
- Event, capability, audit persistence (separate changes).
- Any HTTP layer.

## Decisions

### D1: Scoped, typed, versioned context
`context_records` stores key/value pairs scoped by `scope_type` (tenant/conversation/workflow_instance/state_instance) and `scope_id`. Value is JSONB (PRD 131: the engine stores references and relevant snapshots). `version` enables optimistic updates and change tracking (PRD 31).

### D2: Explicit scope model
`scope_type` VARCHAR + Go constants: `TENANT`, `CONVERSATION`, `WORKFLOW_INSTANCE`, `STATE_INSTANCE`. This mirrors the PRD 23 hierarchy (tenant → conversation → workflow → state → turn) but turn/transient data is not persisted here.

### D3: Persistent memory vs workflow data separation
`memory_references` models persistent memory (PRD 24, 43.2): `owner_type`/`owner_id` (e.g., customer) with named references storing values/snapshots. It is **not** cascade-deleted when a workflow instance is removed — deleting workflow state never deletes user memory (PRD 24). A `source_workflow_instance_id` is optional (provenance) but is a soft reference.

### D4: Tenant isolation everywhere
Both tables carry `tenant_id`; all queries filter by it (PRD 4, 96).

### D5: Repo interface returns current + version
`IContextRepository` lets the engine upsert context and read the full scope snapshot; memory operations are owner-scoped so the context engine can resolve persistent memory without touching workflow data.

## Schema Outline

```
context_records
  id            UUID PK
  tenant_id     UUID NOT NULL
  scope_type    VARCHAR NOT NULL        -- TENANT/CONVERSATION/WORKFLOW_INSTANCE/STATE_INSTANCE
  scope_id      VARCHAR NOT NULL        -- tenant-id / conversation-id / instance-id / state-instance-id
  key           VARCHAR NOT NULL        -- e.g. booking.time_start
  value         JSONB NOT NULL          -- typed value / snapshot (PRD 131)
  version       INT NOT NULL DEFAULT 0  -- optimistic lock (PRD 31)
  created_at / updated_at
  UNIQUE(tenant_id, scope_type, scope_id, key)

memory_references                       -- persistent memory (PRD 24, 43.2)
  id            UUID PK
  tenant_id     UUID NOT NULL
  owner_type    VARCHAR NOT NULL        -- e.g. CUSTOMER / USER
  owner_id      VARCHAR NOT NULL
  name          VARCHAR NOT NULL        -- e.g. address / preferences
  value         JSONB NOT NULL
  source_workflow_instance_id UUID       -- optional provenance, NOT cascade-deleted
  created_at / updated_at
  UNIQUE(tenant_id, owner_type, owner_id, name)
```

Indexes: `context_records(tenant_id, scope_type, scope_id)`, `memory_references(tenant_id, owner_type, owner_id)`.

## Risks / Trade-offs

- **Risk: JSONB value is loosely typed** → Mitigation: intentional for flexibility (snapshots of business data, PRD 131); the Context Engine enforces shape. No DB-level type constraint beyond JSONB.
- **Risk: memory reference provenance soft FK** → Mitigation: `source_workflow_instance_id` is a plain UUID (no hard FK), so memory rows survive instance deletion (PRD 24). Document this so it is not turned into a hard FK later.
- **Risk: scope_id is VARCHAR but instance ids are UUID** → Mitigation: VARCHAR supports both UUID and logical ids (conversation ids may be non-UUID). Indexed for lookups.

## Migration Plan

1. Branch `feature/epic3-persistence-context-memory`.
2. Add migration `00005_context.sql` (Up + Down).
3. Add `db/queries/context.sql`.
4. Run `sqlc generate`.
5. Add entities + `IContextRepository`.
6. Implement `pgx_context_repository.go`.
7. `go build ./...`, `go vet ./...`, `go test ./...`.
8. Optional smoke: `goose up`; upsert context; read scope snapshot; create memory reference; delete workflow instance → memory intact.
9. PR → review → merge.

**Rollback**: migration `Down` drops the two tables; feature branch isolates changes.

## Open Questions

- Whether a conversation/scope registry table is needed (e.g., to validate scope_id). Decision: no — scope ids are validated by the referencing engine; a separate registry is out of scope for persistence slice.
