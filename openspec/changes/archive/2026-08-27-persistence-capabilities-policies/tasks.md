## 1. DB schema — capabilities & policies (Skill: db-sqlc-schema)

- [x] 1.1 Read `.agents/settings.json` and `apps/api/sqlc.yaml` before any change
- [x] 1.2 Create `apps/api/db/migrations/00006_capabilities.sql` with `capabilities` table (id, tenant_id, name, description, provider_type, provider_id, input_schema JSONB, output_schema JSONB, status, version, credential_reference, created_at, updated_at) + `UNIQUE(tenant_id, name)`
- [x] 1.3 Add `capability_bindings` table (id, tenant_id, capability_id FK CASCADE, scope_type, scope_id, permission, created_at, updated_at) + `UNIQUE(tenant_id, capability_id, scope_type, scope_id)` + `INDEX(tenant_id, capability_id)`
- [x] 1.4 Add `policies` table (id, tenant_id, scope_type, scope_id, type, content JSONB, created_at, updated_at) + `UNIQUE(tenant_id, scope_type, scope_id, type)` + `INDEX(tenant_id, scope_type, scope_id)`
- [x] 1.5 Verify: no secrets stored (only `credential_reference`), standard columns, Up + Down blocks

## 2. sqlc queries + regen (Skill: db-sqlc-schema)

- [x] 2.1 Create `apps/api/db/queries/capability.sql`: CreateCapability, FindCapabilityByID, FindCapabilityByName, ListCapabilitiesByTenant, UpdateCapabilityStatus, BindCapability, ListBindingsByCapability, ListBindingsByScope, UpsertPolicy, FindPolicyByType, ListPoliciesByScope
- [x] 2.2 Every query filters by `tenant_id` (PRD 4, 96)
- [x] 2.3 Run `sqlc generate` from `apps/api` — no errors; do NOT edit generated files

## 3. Domain entities (Skill: api-feature)

- [x] 3.1 Read `.agents/guides/api-entity.md`
- [x] 3.2 Create `internal/domain/entities/capability.go` — `Capability` + `ProviderType` constants (MCP/INTERNAL/HTTP/FUTURE) + `CapabilityStatus` constants (ACTIVE/INACTIVE/DISABLED)
- [x] 3.3 Create `internal/domain/entities/capability_binding.go` — `CapabilityBinding` + `BindingScopeType` constants (TENANT/WORKFLOW/STATE) + `BindingPermission` constants (ALLOW/DENY)
- [x] 3.4 Create `internal/domain/entities/policy.go` — `Policy` + `PolicyScopeType` constants (TENANT/WORKFLOW/STATE)
- [x] 3.5 Verify: no `interface{}`/`any`; typed Go constants; no secrets in entities

## 4. Repository interface (Skill: api-feature)

- [x] 4.1 Read `.agents/guides/api-repository.md`
- [x] 4.2 Create `internal/domain/repositories/capability_repository.go` — `ICapabilityRepository`
- [x] 4.3 All methods take explicit `ctx` + `tenantID string`
- [x] 4.4 Methods: Create, FindByID, FindByName, ListByTenant, UpdateStatus, Bind, ListBindingsByCapability, ListBindingsByScope, UpsertPolicy, FindPolicyByType, ListPoliciesByScope
- [x] 4.5 Operates on entities only; returns DomainError `NOT_FOUND`/`CONFLICT`

## 5. PostgreSQL adapter (Skill: api-feature)

- [x] 5.1 Read `.agents/guides/api-db-repository.md`
- [x] 5.2 Create `internal/infrastructure/database/pgx_capability_repository.go` implementing `ICapabilityRepository`
- [x] 5.3 Constructor `NewPgxCapabilityRepository(pool *pgxpool.Pool) repositories.ICapabilityRepository`
- [x] 5.4 Map `pgx.ErrNoRows` → `NOT_FOUND`; unique violation → `CONFLICT`

## 6. Quality gate (Skill: api-code-review)

- [x] 6.1 Run `go build ./...`, `go vet ./...`, `go test ./...` in `apps/api`
- [x] 6.2 Smoke: `goose up`; register capability; bind to scopes; upsert policy
- [x] 6.3 `sqlc generate` idempotent
- [x] 6.4 `api-code-review`: tenant-scoped, no secrets, no business logic in adapter, no edited sqlc files
- [x] 6.5 All files end with a newline (EOF)
