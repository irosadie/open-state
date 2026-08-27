## 1. DB schema — event system (Skill: db-sqlc-schema)

- [ ] 1.1 Read `.agents/settings.json` and `apps/api/sqlc.yaml` before any change
- [ ] 1.2 Create `apps/api/db/migrations/00004_events.sql` with `events` table (id, tenant_id, event_id, type, source, aggregate_id, workflow_instance_id FK nullable, sequence BIGSERIAL, timestamp, payload JSONB, correlation_id, causation_id, idempotency_key, created_at) + `UNIQUE(tenant_id, event_id)` + `INDEX(workflow_instance_id, sequence)`
- [ ] 1.3 Add `event_inbox` table (id, tenant_id, idempotency_key, event_type, source, payload JSONB, status, attempt_count, received_at, processed_at, created_at, updated_at) + `UNIQUE(tenant_id, idempotency_key)`
- [ ] 1.4 Add `event_outbox` table (id, tenant_id, event_id FK, payload JSONB, topic, status, attempt_count, published_at, created_at, updated_at) + `INDEX(tenant_id, status)`
- [ ] 1.5 Add `idempotency_records` table (id, tenant_id, idempotency_key, scope, result_status, payload JSONB, created_at, updated_at) + `UNIQUE(tenant_id, idempotency_key)`
- [ ] 1.6 Verify: standard columns, snake_case, FKs ON DELETE, indexes, Up + Down blocks

## 2. sqlc queries + regen (Skill: db-sqlc-schema)

- [ ] 2.1 Create `apps/api/db/queries/event.sql`: AppendEvent, FindEventByID, ListEventsByInstance (ordered by sequence), ListEventsByTenant, InsertInboxEvent, ClaimInboxEvents (batch atomic), MarkInboxProcessed, InsertOutboxEvent, ClaimOutboxEvents (batch atomic), MarkOutboxPublished, UpsertIdempotencyRecord, FindIdempotencyRecord
- [ ] 2.2 Every query filters by `tenant_id` (PRD 4, 96); ordering by `sequence` (PRD 32)
- [ ] 2.3 Dedup enforced via `UNIQUE(tenant_id, idempotency_key)`
- [ ] 2.4 Run `sqlc generate` from `apps/api` — no errors; do NOT edit generated files

## 3. Domain entities (Skill: api-feature)

- [ ] 3.1 Read `.agents/guides/api-entity.md`
- [ ] 3.2 Create `internal/domain/entities/event.go` — `Event` struct + `EventSource` typed constants (USER/LLM/MCP/WEBHOOK/SYSTEM/SCHEDULER/ADMIN/API)
- [ ] 3.3 Create `internal/domain/entities/inbox_event.go` — `InboxEvent` + `InboxEventStatus` constants (RECEIVED/PROCESSING/PROCESSED/FAILED)
- [ ] 3.4 Create `internal/domain/entities/outbox_event.go` — `OutboxEvent` + `OutboxEventStatus` constants (PENDING/PUBLISHED/FAILED)
- [ ] 3.5 Create `internal/domain/entities/idempotency_record.go` — `IdempotencyRecord` + `IdempotencyResultStatus` constants (PROCESSED/IGNORED/FAILED)
- [ ] 3.6 Verify: no `interface{}`/`any`; typed Go constants

## 4. Repository interface (Skill: api-feature)

- [ ] 4.1 Read `.agents/guides/api-repository.md`
- [ ] 4.2 Create `internal/domain/repositories/event_repository.go` — `IEventRepository`
- [ ] 4.3 All methods take explicit `ctx` + `tenantID string`
- [ ] 4.4 Methods: Append, FindEventByID, ListEventsByInstance, ListEventsByTenant, InsertInbox, ClaimInbox, MarkInboxProcessed, InsertOutbox, ClaimOutbox, MarkOutboxPublished, UpsertIdempotency, FindIdempotency
- [ ] 4.5 Operates on entities only; returns DomainError `NOT_FOUND`/`CONFLICT`

## 5. PostgreSQL adapter (Skill: api-feature)

- [ ] 5.1 Read `.agents/guides/api-db-repository.md`
- [ ] 5.2 Create `internal/infrastructure/database/pgx_event_repository.go` implementing `IEventRepository`
- [ ] 5.3 Constructor `NewPgxEventRepository(pool *pgxpool.Pool) repositories.IEventRepository`
- [ ] 5.4 Claim queries use atomic `UPDATE ... WHERE status='PENDING' ... RETURNING` for safe worker batching
- [ ] 5.5 Map `pgx.ErrNoRows` → `NOT_FOUND`; unique violation → `CONFLICT`

## 6. Quality gate (Skill: api-code-review)

- [ ] 6.1 Run `go build ./...`, `go vet ./...`, `go test ./...` in `apps/api`
- [ ] 6.2 Smoke: `goose up`; append event; enqueue + claim outbox; dedup second inbound
- [ ] 6.3 `sqlc generate` idempotent
- [ ] 6.4 `api-code-review`: tenant-scoped, dedup, no business logic in adapter, no edited sqlc files
- [ ] 6.5 All files end with a newline (EOF)
