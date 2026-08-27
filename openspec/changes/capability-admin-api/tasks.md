## 1. DTOs & validators (Skill: api-feature)

- [ ] 1.1 Read `.agents/settings.json`, `.agents/guides/api-dto.md`, `.agents/guides/api-validator.md`
- [ ] 1.2 Create capability DTOs in `apps/api/internal/application/dtos/` (create/update/read/list, binding, test-invocation request+response, error shape)
- [ ] 1.3 DTOs expose only `credential_reference`, never resolved secrets (PRD §61, §91)
- [ ] 1.4 Add input validators for create/update/binding/test payloads (PRD §62 schema validation reuse)
- [ ] 1.5 Unit test DTO/validator edge cases (missing fields, invalid provider_type/status/scope)

## 2. Application service (Skill: api-feature)

- [ ] 2.1 Read `.agents/guides/api-service.md`
- [ ] 2.2 Create `apps/api/internal/application/services/capability_service.go` — `CapabilityService` orchestrating `ICapabilityRepository` (PRD §174)
- [ ] 2.3 Methods: Create, FindByID, List, Update, Delete/Disable, ListBindings, Bind, Unbind, TestInvoke
- [ ] 2.4 All queries filter by tenant (PRD §4, §96); map repo DomainError → application errors
- [ ] 2.5 `TestInvoke` forces mock/sandbox provider and returns normalized result (PRD §2064, §153)

## 3. Controller (Skill: api-feature)

- [ ] 3.1 Read `.agents/guides/api-controller.md`
- [ ] 3.2 Create `apps/api/internal/interfaces/http/controllers/capability_controller.go`
- [ ] 3.3 Handlers: list, get, create, update, delete, list-bindings, create-binding, delete-binding, test-invoke
- [ ] 3.4 Derive `tenantID` from auth context; never from request body; no business logic in controller (PRD §74)

## 4. Routes (Skill: api-feature)

- [ ] 4.1 Read `.agents/guides/api-route.md`
- [ ] 4.2 Register `/capabilities` route group in `apps/api/internal/interfaces/http/routes/routes.go` behind auth + tenant middleware
- [ ] 4.3 Wire sub-paths: `/capabilities`, `/capabilities/{id}`, `/capabilities/{id}/bindings`, `/bindings/{id}`, `/capabilities/{id}/test`
- [ ] 4.4 Ensure error handler returns classified status codes (400/401/403/404/409/422/500) (PRD §74 error mapping)

## 5. OpenAPI docs (Skill: docs-openapi)

- [ ] 5.1 Read `docs/openapi/base.json` and existing `paths/`/`schemas/` conventions
- [ ] 5.2 Add `docs/openapi/paths/capabilities.json` — all capability/binding/test paths
- [ ] 5.3 Add `docs/openapi/schemas/capability.json` — Capability, CapabilityBinding, InvocationResult, CapabilityError
- [ ] 5.4 Verify the split OpenAPI validates (paths + schemas resolve against base.json)

## 6. Quality gate (Skills: api-code-review, meta-skill-hygiene)

- [ ] 6.1 `go build ./...`, `go vet ./...`, `go test ./...` pass in `apps/api`
- [ ] 6.2 Smoke: auth required; register capability; bind to scope; test-invoke returns mock result; delete binding
- [ ] 6.3 Secret-safety review: no credential value returned/logged anywhere in DTO/controller
- [ ] 6.4 `api-code-review` passes on the capability API
- [ ] 6.5 All files end with a newline (EOF)
