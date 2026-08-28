## 1. Audit actions + AuditWriter (Skill: api-feature)

- [x] 1.1 Extend `internal/domain/entities/audit_log.go`: `workflow.published`, `capability.invoked`, `capability.denied`, `binding.created`, `binding.deleted`, `rbac.role_assigned/updated/removed`, `authorization.denied`
- [x] 1.2 Create `internal/application/services/audit_writer.go` (`AuditWriter`, best-effort)

## 2. Wire audit into operations (Skill: api-feature)

- [x] 2.1 `BuilderService.Publish` → `workflow.published`
- [x] 2.2 `CapabilityService.TestInvoke` → `capability.invoked`/`capability.denied`
- [x] 2.3 `CapabilityService.Bind`/`Unbind` → `binding.created`/`binding.deleted`
- [x] 2.4 Controllers pass actor (user id) from JWT context

## 3. Filtered + paginated queries (Skill: db-sqlc-schema)

- [x] 3.1 Add `ListAuditFiltered`/`CountAuditFiltered` to `apps/api/db/queries/audit.sql`
- [x] 3.2 Run `sqlc generate`
- [x] 3.3 Add `ListFiltered`/`CountFiltered` + `AuditFilter` to `IAuditRepository`
- [x] 3.4 Implement in `pgx_audit_repository.go`

## 4. Audit query API (Skill: api-feature)

- [x] 4.1 Create `AuditService` (`application/services/audit_service.go`) with filters + pagination
- [x] 4.2 Create `AuditController` + DTOs
- [x] 4.3 Register `GET /api/audit` behind `audit:read`
- [x] 4.4 Update server wiring (`CreateApp`/`main.go`)

## 5. OpenAPI (Skill: docs-openapi)

- [x] 5.1 Add `docs/openapi/paths/audit.json` + `docs/openapi/schemas/audit.json`
- [x] 5.2 Update merged `docs/openapi.json` (path, schemas, tag)

## 6. Frontend audit UI (Skill: web-api-integrated)

- [x] 6.1 `packages/schemas/audit.ts` (Zod + action labels)
- [x] 6.2 `packages/types/audit-response.ts`
- [x] 6.3 Constants (`api-routers.ts`, `query-keys.ts`)
- [x] 6.4 `useAuditList` hook
- [x] 6.5 `/admin/audit` page + content + table

## 7. Verify

- [x] 7.1 `go build ./...` + `go vet ./...` pass
- [x] 7.2 Backend tests pass
- [x] 7.3 Frontend `tsc --noEmit` + biome pass
- [x] 7.4 `gofmt` clean on changed files
