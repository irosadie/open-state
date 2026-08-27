## Why

Epic **#3 (Data & Persistence)** requires the event model to be the durable driver of state transitions (PRD 27-32) with reliable delivery and replay (PRD 51-52, 65-66). This is the third persistence slice: the **event system** — `events` (immutable history), `event_inbox` (inbound dedup, PRD 66), `event_outbox` (reliable emit, PRD 65), and `idempotency_records` (PRD 30).

Events must survive restarts (PRD 128), be ordered per workflow instance (PRD 32), carry tenant + project + correlation/causation (PRD 27), and be deduplicated (PRD 30, 66). Outbox/inbox guarantee atomic consistency with DB state changes (PRD 65-66).

## What Changes

- **NEW** — `apps/api/db/migrations/00004_events.sql` goose migration creating:
  - `events` — append-only immutable event history (PRD 27, 51); full event model fields.
  - `event_inbox` — inbound event dedup/queue (PRD 66) with processing state.
  - `event_outbox` — outbound reliable delivery queue (PRD 65) with published state.
  - `idempotency_records` — dedup ledger keyed by idempotency_key (PRD 30).
- **NEW** — `apps/api/db/queries/event.sql` sqlc query file.
- **NEW** — Domain entities `Event`, `InboxEvent`, `OutboxEvent`, `IdempotencyRecord` in `apps/api/internal/domain/entities/`.
- **NEW** — Repository interface `apps/api/internal/domain/repositories/event_repository.go` (`IEventRepository`).
- **NEW** — pgx adapter `apps/api/internal/infrastructure/database/pgx_event_repository.go`.
- **REGENERATED** — sqlc-generated Go under `apps/api/internal/infrastructure/db/`.
- Uses **`db-sqlc-schema`** and **`api-feature`** skills. `docs-openapi` not touched (no public endpoint).

## Capabilities

### New Capabilities

- `backend/persistence/events`: append-only event history, inbound inbox, reliable outbox, and idempotency records — behind `IEventRepository` and a pgx adapter.

### Modified Capabilities

- `backend/persistence/workflow-instances`: `events.workflow_instance_id` references `workflow_instances` defined there.

## Impact

- **`apps/api/db/migrations/`** — add `00004_events.sql`.
- **`apps/api/db/queries/`** — add `event.sql`.
- **`apps/api/internal/infrastructure/db/`** — sqlc-generated code regenerated.
- **`apps/api/internal/domain/entities/`** — add `event.go`, `inbox_event.go`, `outbox_event.go`, `idempotency_record.go`.
- **`apps/api/internal/domain/repositories/`** — add `event_repository.go`.
- **`apps/api/internal/infrastructure/database/`** — add `pgx_event_repository.go`.
- **No** changes to web, worker, shared packages, OpenAPI, docker.

## Non-Goals

- The event processing pipeline (event engine, guard eval, transition execution) — separate epic.
- Actual message-bus publish/subscribe (NATS/Kafka/Redis Streams, PRD 71) — this change persists to the outbox; the publisher worker is a later change.
- Context/memory, capability, audit persistence — separate changes.
- HTTP controllers/routes.

## Dependencies

- `persistence-workflow-definitions` and `persistence-runtime-instances` (FK targets `workflow_versions`, `workflow_instances`).
- Epic #3.
