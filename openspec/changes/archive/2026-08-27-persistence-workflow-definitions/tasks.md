## 1. DB schema — workflow definitions (Skill: db-sqlc-schema)

- [x] 1.1 Read `.agents/settings.json` and `apps/api/sqlc.yaml` before any change
- [x] 1.2 Create `apps/api/db/migrations/00002_workflows.sql` with `projects` table (id, tenant_id, name, slug, status) + `UNIQUE(tenant_id, slug)`; and `workflows` table (id, tenant_id, project_id, slug, name, description, status, current_version, version, created_at, updated_at) + `UNIQUE(tenant_id, project_id, slug)` + `-- +goose Down`
- [x] 1.3 Add `workflow_versions` table (id, workflow_id FK, tenant_id, project_id, version_no, definition JSONB, status, is_current, created_at, updated_at) + `UNIQUE(workflow_id, version_no)` + indexes
- [x] 1.4 Add `states` table (id, workflow_version_id FK, key, kind, name, description, instructions, required_context JSONB, capabilities JSONB, policy JSONB, is_terminal, position JSONB) + `UNIQUE(workflow_version_id, key)`
- [x] 1.5 Add `transitions` table (id, workflow_version_id FK, key, source_state_id FK, target_state_id FK, event, priority, is_active) + `UNIQUE(workflow_version_id, key)`
- [x] 1.6 Add `transition_guards` table (id, transition_id FK, workflow_version_id FK, logic, conditions JSONB)
- [x] 1.7 Verify: all tables use `id UUID PK DEFAULT gen_random_uuid()`, `created_at`, `updated_at` standard columns; snake_case; FK `ON DELETE CASCADE`; indexes on FK/tenant/project columns

## 2. sqlc queries + regen (Skill: db-sqlc-schema)

- [x] 2.1 Create `apps/api/db/queries/workflow.sql` with: CreateProject, FindProjectByID, FindProjectBySlug, ListProjectsByTenant, CreateWorkflow, FindWorkflowByID, FindWorkflowBySlug, ListWorkflowsByTenant, UpdateWorkflowStatus (optimistic), UpdateWorkflowVersion (optimistic), CreateWorkflowVersion, FindWorkflowVersionByNumber, FindCurrentWorkflowVersion, ListWorkflowVersions, CreateState, ListStatesByVersion, CreateTransition, ListTransitionsByVersion, CreateTransitionGuard, ListGuardsByTransition
- [x] 2.2 Every query filters by `tenant_id` (and `project_id` for workflow-scoped queries) (PRD 4, 96, 3.1.1)
- [x] 2.3 Optimistic-lock updates use `WHERE ... AND version = $n` and `SET version = version + 1`
- [x] 2.4 Run `sqlc generate` from `apps/api` — no errors
- [x] 2.5 Do NOT edit generated files in `internal/infrastructure/db/`

## 3. Domain entities (Skill: api-feature)

- [x] 3.1 Read `.agents/guides/api-entity.md`
- [x] 3.2 Create `internal/domain/entities/project.go` — `Project` struct + `ProjectStatus` typed constants (ACTIVE/ARCHIVED); create `workflow.go` — `Workflow` struct (incl. `ProjectID`) + `WorkflowStatus` typed constants (DRAFT/VALIDATING/VALID/PUBLISHED/ARCHIVED)
- [x] 3.3 Create `internal/domain/entities/workflow_version.go` — `WorkflowVersion` struct (incl. `ProjectID`) + `VersionStatus` constants
- [x] 3.4 Create `internal/domain/entities/state.go` — `State` struct + `StateKind` typed constants (START/STATE/DECISION/WAIT/END/EVENT); `RequiredContext`, `Capabilities`, `Policy`, `Position` as JSON-typed fields (use `[]byte` or `json.RawMessage` / defined structs)
- [x] 3.5 Create `internal/domain/entities/transition.go` — `Transition` struct + `TransitionGuard` struct
- [x] 3.6 Verify: no `interface{}`/`any` without justification; typed Go constants for fixed value sets

## 4. Repository interface (Skill: api-feature)

- [x] 4.1 Read `.agents/guides/api-repository.md`
- [x] 4.2 Create `internal/domain/repositories/workflow_repository.go` — `IWorkflowRepository` interface; create `internal/domain/repositories/project_repository.go` — `IProjectRepository` interface
- [x] 4.3 All methods accept explicit `ctx` + `tenantID string`; workflow methods also accept `projectID string`
- [x] 4.4 Interface operates on domain entities only (no sqlc rows); returns `error` (DomainError from `packages/go-shared` on not-found/conflict)
- [x] 4.5 Methods: Create, FindByID, FindBySlug, ListByTenant, UpdateStatus (optimistic), CreateVersion, FindCurrentVersion, ListVersions, FindVersionByNumber, ListStatesByVersion, ListTransitionsByVersion, ListGuardsByTransition

## 5. PostgreSQL adapter (Skill: api-feature)

- [x] 5.1 Read `.agents/guides/api-db-repository.md`
- [x] 5.2 Create `internal/infrastructure/database/pgx_workflow_repository.go` implementing `IWorkflowRepository` via sqlc `Queries`; create `internal/infrastructure/database/pgx_project_repository.go` implementing `IProjectRepository`
- [x] 5.3 Constructors `NewPgxWorkflowRepository(pool *pgxpool.Pool)` and `NewPgxProjectRepository(pool *pgxpool.Pool)` (mirror `NewPgxAuthRepository`)
- [x] 5.4 Map sqlc rows → entities (UUID → string, JSONB → typed fields); handle `pgx.ErrNoRows` → DomainError `NOT_FOUND`
- [x] 5.5 Transactional publish: insert version + set `is_current` + bump workflow `current_version` + bump `version` atomically
- [x] 5.6 Map optimistic-lock zero-row updates → DomainError `CONFLICT`

## 6. Quality gate (Skill: api-code-review)

- [x] 6.1 Run `go build ./...`, `go vet ./...`, `go test ./...` in `apps/api`
- [x] 6.2 Run `goose -dir db/migrations postgres "$DATABASE_URL" up` smoke test (workflow insert → query)
- [x] 6.3 Run `sqlc generate`; confirm no diff to generated code is required after (idempotent)
- [x] 6.4 `api-code-review`: no business logic in adapter, no plain strings for fixed-value fields, no edited sqlc files, tenant-scoped everywhere
- [x] 6.5 All files end with a newline (EOF)
