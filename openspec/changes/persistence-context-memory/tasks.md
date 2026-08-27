## 1. DB schema — context & memory (Skill: db-sqlc-schema)

- [ ] 1.1 Read `.agents/settings.json` and `apps/api/sqlc.yaml` before any change
- [ ] 1.2 Create `apps/api/db/migrations/00005_context.sql` with `context_records` table (id, tenant_id, scope_type, scope_id, key, value JSONB, version, created_at, updated_at) + `UNIQUE(tenant_id, scope_type, scope_id, key)` + `INDEX(tenant_id, scope_type, scope_id)`
- [ ] 1.3 Add `memory_references` table (id, tenant_id, owner_type, owner_id, name, value JSONB, source_workflow_instance_id plain UUID nullable, created_at, updated_at) + `UNIQUE(tenant_id, owner_type, owner_id, name)` + `INDEX(tenant_id, owner_type, owner_id)`
- [ ] 1.4 Verify: `source_workflow_instance_id` is NOT a hard FK (memory survives instance deletion, PRD 24), standard columns, Up + Down blocks

## 2. sqlc queries + regen (Skill: db-sqlc-schema)

- [ ] 2.1 Create `apps/api/db/queries/context.sql`: UpsertContext (optimistic, version bump), FindContextByScope, ListContextByScope, DeleteContext, UpsertMemoryReference, FindMemoryReference, ListMemoryByOwner, DeleteMemoryReference
- [ ] 2.2 Every query filters by `tenant_id` (PRD 4, 96)
- [ ] 2.3 Optimistic upsert uses `ON CONFLICT ... WHERE ... AND version = $n` semantics / `version = version + 1`
- [ ] 2.4 Run `sqlc generate` from `apps/api` — no errors; do NOT edit generated files

## 3. Domain entities (Skill: api-feature)

- [ ] 3.1 Read `.agents/guides/api-entity.md`
- [ ] 3.2 Create `internal/domain/entities/context_record.go` — `ContextRecord` + `ContextScopeType` typed constants (TENANT/CONVERSATION/WORKFLOW_INSTANCE/STATE_INSTANCE)
- [ ] 3.3 Create `internal/domain/entities/memory_reference.go` — `MemoryReference` struct
- [ ] 3.4 Verify: no `interface{}`/`any`; typed Go constants

## 4. Repository interface (Skill: api-feature)

- [ ] 4.1 Read `.agents/guides/api-repository.md`
- [ ] 4.2 Create `internal/domain/repositories/context_repository.go` — `IContextRepository`
- [ ] 4.3 All methods take explicit `ctx` + `tenantID string`
- [ ] 4.4 Methods: UpsertContext (optimistic), FindContextByScope, ListContextByScope, DeleteContext, UpsertMemoryReference, FindMemoryReference, ListMemoryByOwner, DeleteMemoryReference
- [ ] 4.5 Operates on entities only; returns DomainError `NOT_FOUND`/`CONFLICT`

## 5. PostgreSQL adapter (Skill: api-feature)

- [ ] 5.1 Read `.agents/guides/api-db-repository.md`
- [ ] 5.2 Create `internal/infrastructure/database/pgx_context_repository.go` implementing `IContextRepository`
- [ ] 5.3 Constructor `NewPgxContextRepository(pool *pgxpool.Pool) repositories.IContextRepository`
- [ ] 5.4 Upsert maps unique/optimistic violations → `CONFLICT`; `pgx.ErrNoRows` → `NOT_FOUND`

## 6. Quality gate (Skill: api-code-review)

- [ ] 6.1 Run `go build ./...`, `go vet ./...`, `go test ./...` in `apps/api`
- [ ] 6.2 Smoke: `goose up`; upsert context; read scope snapshot; stale version → conflict; create memory reference; delete workflow instance → memory intact
- [ ] 6.3 `sqlc generate` idempotent
- [ ] 6.4 `api-code-review`: tenant-scoped, memory/workflow separation, no business logic in adapter, no edited sqlc files
- [ ] 6.5 All files end with a newline (EOF)
