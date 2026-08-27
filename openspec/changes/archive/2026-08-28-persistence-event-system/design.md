## Context

Events are the durable driver of the state machine (PRD 27-32) and must survive restarts (PRD 128) while enabling replay (PRD 51-52). Reliable delivery uses the outbox (PRD 65) and inbound dedup uses the inbox (PRD 66) + idempotency (PRD 30). This change adds the event-system persistence slice of epic #3.

## Goals / Non-Goals

**Goals:**
- Schema for `events`, `event_inbox`, `event_outbox`, `idempotency_records`.
- Domain entities + `IEventRepository`.
- pgx adapter backed by sqlc.
- Tenant scoping + per-instance ordering + dedup.

**Non-Goals:**
- Event processing/transition engine (separate epic).
- Outbox publisher / message-bus integration (PRD 71) — later.
- Context, capability, audit persistence (separate changes).
- Any HTTP layer.

## Decisions

### D1: Immutable append-only event history
`events` is append-only. Rows are never updated/deleted during normal operation (PRD 51, 52). It stores the full PRD 27 event model: `event_id` (logical, unique per tenant), `type`, `source`, `aggregate_id`, `workflow_instance_id`, `timestamp`, `payload` (JSONB), `correlation_id`, `causation_id`, `idempotency_key`.

### D2: Event source as VARCHAR + Go constants
`events.source` mirrors PRD 28: `USER`, `LLM`, `MCP`, `WEBHOOK`, `SYSTEM`, `SCHEDULER`, `ADMIN`, `API`. VARCHAR + Go typed constants (no PG ENUM).

### D3: Outbox for reliable emit (PRD 65)
`event_outbox` persists events that must be published to the bus. Written **atomically in the same transaction** as the DB state change they accompany (PRD 65, 69). `status` transitions `PENDING → PUBLISHED` (→ `FAILED` with retry). The publisher worker (later change) reads `PENDING`, publishes, and marks `PUBLISHED`.

### D4: Inbox for inbound dedup (PRD 66)
`event_inbox` stores inbound external events before processing. A unique constraint on `idempotency_key` (per tenant) enforces dedup (PRD 30, 66). `status` `RECEIVED → PROCESSING → PROCESSED` (→ `FAILED`).

### D5: Idempotency ledger
`idempotency_records` is a tenant-scoped ledger keyed by `idempotency_key`, storing the result status so repeated external deliveries are ignored (PRD 30). It is the durable check the engine consults before applying an event.

### D6: Per-instance ordering
Events for the same `workflow_instance_id` must process deterministically (PRD 32). The schema stores `sequence` (global per-tenant auto-increment) and a `sequence` on `events` for replay order. Ordering enforcement happens in the engine/worker (later), but the schema supports it via `(workflow_instance_id, sequence)` index.

### D7: Tenant isolation
All four tables carry `tenant_id`; all queries filter by it (PRD 4, 96).

## Schema Outline

```
events                                  -- append-only history (PRD 27, 51)
  id            UUID PK
  tenant_id     UUID NOT NULL
  event_id      VARCHAR NOT NULL         -- logical id (unique per tenant)
  type          VARCHAR NOT NULL         -- e.g. payment.success
  source        VARCHAR NOT NULL         -- PRD 28
  aggregate_id  VARCHAR
  workflow_instance_id UUID REFERENCES workflow_instances(id)
  sequence      BIGSERIAL NOT NULL
  timestamp     TIMESTAMP NOT NULL DEFAULT NOW()
  payload       JSONB NOT NULL DEFAULT '{}'
  correlation_id VARCHAR
  causation_id  VARCHAR
  idempotency_key VARCHAR
  created_at
  UNIQUE(tenant_id, event_id)
  INDEX(workflow_instance_id, sequence)  -- PRD 32

event_inbox                             -- inbound dedup (PRD 66)
  id            UUID PK
  tenant_id     UUID NOT NULL
  idempotency_key VARCHAR NOT NULL
  event_type    VARCHAR NOT NULL
  source        VARCHAR NOT NULL
  payload       JSONB NOT NULL
  status        VARCHAR NOT NULL DEFAULT 'RECEIVED'   -- RECEIVED/PROCESSING/PROCESSED/FAILED
  attempt_count INT NOT NULL DEFAULT 0
  received_at   TIMESTAMP NOT NULL DEFAULT NOW()
  processed_at  TIMESTAMP
  created_at / updated_at
  UNIQUE(tenant_id, idempotency_key)

event_outbox                            -- reliable emit (PRD 65)
  id            UUID PK
  tenant_id     UUID NOT NULL
  event_id      UUID REFERENCES events(id)
  payload       JSONB NOT NULL
  topic         VARCHAR NOT NULL
  status        VARCHAR NOT NULL DEFAULT 'PENDING'    -- PENDING/PUBLISHED/FAILED
  attempt_count INT NOT NULL DEFAULT 0
  published_at  TIMESTAMP
  created_at / updated_at
  INDEX(tenant_id, status)

idempotency_records                     -- dedup ledger (PRD 30)
  id            UUID PK
  tenant_id     UUID NOT NULL
  idempotency_key VARCHAR NOT NULL
  scope         VARCHAR NOT NULL DEFAULT 'event'
  result_status VARCHAR NOT NULL        -- PROCESSED/IGNORED/FAILED
  payload       JSONB
  created_at / updated_at
  UNIQUE(tenant_id, idempotency_key)
```

## Risks / Trade-offs

- **Risk: outbox grows unbounded** → Mitigation: `PUBLISHED` rows are eligible for archival/purge (later ops concern); index on `(tenant_id, status)` keeps the worker scan fast.
- **Risk: inbox/outbox status races across workers** → Mitigation: optimistic `attempt_count` + status transitions; a later change adds a worker lease/claim. Adapter exposes safe claim queries (`UPDATE ... WHERE status='PENDING' LIMIT n RETURNING ...`).
- **Risk: idempotency vs inbox duplication** → Mitigation: both keyed on `idempotency_key`; the inbox is the inbound gate, `idempotency_records` the durable outcome ledger. Kept separate per PRD 66 (dedup) vs PRD 30 (idempotent processing).

## Migration Plan

1. Branch `feature/epic3-persistence-event-system`.
2. Add migration `00004_events.sql` (Up + Down).
3. Add `db/queries/event.sql`.
4. Run `sqlc generate`.
5. Add entities + `IEventRepository`.
6. Implement `pgx_event_repository.go`.
7. `go build ./...`, `go vet ./...`, `go test ./...`.
8. Optional smoke: `goose up`; append event; enqueue outbox; claim inbox.
9. PR → review → merge.

**Rollback**: migration `Down` drops the four tables; feature branch isolates changes.

## Open Questions

- Whether `event_id` on outbox should reference `events.id` (hard FK) or be a plain logical id. Decision: `event_id UUID REFERENCES events(id)` for referential integrity; payload denormalized for independent publishing.
