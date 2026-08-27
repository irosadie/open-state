## 1. DB schema — runtime instances (Skill: db-sqlc-schema)

- [x] 1.1 Read `.agents/settings.json` and `apps/api/sqlc.yaml` before any change
- [x] 1.2 Create `apps/api/db/migrations/00003_workflow_instances.sql` with `workflow_instances` table (id, tenant_id, workflow_id FK, workflow_version_id FK, correlation_key, status, version, started_at, completed_at, expires_at, created_at, updated_at) + `-- +goose Down`
- [x] 1.3 Add `state_instances` table (id, tenant_id, workflow_instance_id FK CASCADE, workflow_version_id FK, state_key, state_id FK nullable, status, version, retry_count, entered_at, expires_at, exited_at, created_at, updated_at)
- [x] 1.4 Add `ALTER TABLE workflow_instances ADD CONSTRAINT ... FOREIGN KEY (current_state_instance_id) REFERENCES state_instances(id)` (nullable, after both tables exist) + indexes (tenant_id/status, correlation_key, workflow_instance_id)
- [x] 1.5 Verify: standard columns, snake_case, FKs with ON DELETE, indexes on FK/tenant columns, Up + Down blocks

## 2. sqlc queries + regen (Skill: db-sqlc-schema)

- [x] 2.1 Create `apps/api/db/queries/instance.sql`: CreateWorkflowInstance, FindWorkflowInstanceByID, ListWorkflowInstancesByTenant, UpdateWorkflowInstanceStatus (optimistic), IncrementWorkflowInstanceVersion (optimistic), CreateStateInstance, FindStateInstanceByID, UpdateStateInstanceStatus (optimistic), UpdateStateInstanceRetry (optimistic), TransitionState (transactional: exit old + enter new + bump parent version), SetCurrentStateInstance
- [x] 2.2 Every query filters by `tenant_id` (PRD 4, 96)
- [x] 2.3 Optimistic-lock updates use `WHERE ... AND version = $n` + `SET version = version + 1`
- [x] 2.4 `TransitionState` runs in a single transaction (sqlc `.sql` comment + adapter tx)
- [x] 2.5 Run `sqlc generate` from `apps/api` — no errors; do NOT edit generated files

## 3. Domain entities (Skill: api-feature)

- [x] 3.1 Read `.agents/guides/api-entity.md`
- [x] 3.2 Create `internal/domain/entities/workflow_instance.go` — `WorkflowInstance` + `WorkflowInstanceStatus` typed constants (CREATED/RUNNING/WAITING/COMPLETED/CANCELLED/FAILED/EXPIRED/ABORTED/SUSPENDED)
- [x] 3.3 Create `internal/domain/entities/state_instance.go` — `StateInstance` + `StateInstanceStatus` typed constants (ENTERING/ACTIVE/WAITING/EXITING/COMPLETED/FAILED/EXPIRED/CANCELLED)
- [x] 3.4 Verify: no `interface{}`/`any` without justification; typed Go constants

## 4. Repository interface (Skill: api-feature)

- [x] 4.1 Read `.agents/guides/api-repository.md`
- [x] 4.2 Create `internal/domain/repositories/instance_repository.go` — `IInstanceRepository`
- [x] 4.3 All methods take explicit `ctx` + `tenantID string`
- [x] 4.4 Methods: Create, FindByID, ListByTenant, UpdateStatus (optimistic), Transition (atomic), CreateStateInstance, FindStateInstanceByID, UpdateStateInstanceStatus (optimistic), IncrementRetry (optimistic)
- [x] 4.5 Operates on entities only; returns DomainError `NOT_FOUND`/`CONFLICT`

## 5. PostgreSQL adapter (Skill: api-feature)

- [x] 5.1 Read `.agents/guides/api-db-repository.md`
- [x] 5.2 Create `internal/infrastructure/database/pgx_instance_repository.go` implementing `IInstanceRepository` via sqlc `Queries`
- [x] 5.3 Constructor `NewPgxInstanceRepository(pool *pgxpool.Pool) repositories.IInstanceRepository`
- [x] 5.4 Implement `Transition` as a real DB transaction: exit old state + insert new state + set current + bump parent instance version
- [x] 5.5 Map `pgx.ErrNoRows` → `NOT_FOUND`; optimistic zero-row → `CONFLICT`

## 6. Quality gate (Skill: api-code-review)

- [x] 6.1 Run `go build ./...`, `go vet ./...`, `go test ./...` in `apps/api`
- [x] 6.2 Smoke: `goose up`; create instance + state instances; optimistic-lock update; stale update → conflict
- [x] 6.3 `sqlc generate` idempotent
- [x] 6.4 `api-code-review`: tenant-scoped, optimistic locking, no business logic in adapter, no edited sqlc files
- [x] 6.5 All files end with a newline (EOF)
