## Why

Epic **#4 (MCP & Integrasi)** needs an administrative control plane over the Capability
Registry and its bindings so tenants/operators can register capabilities, bind them to
scopes, view the registry, and test/simulate execution. The data layer
(`persistence-capabilities-policies`) explicitly excluded any HTTP layer, and the
execution layer (`mcp-capability-execution`) is domain-only. This third slice —
**capability-admin-api** — exposes the registry, bindings, policies, and a test/simulate
endpoint over the existing Echo API with proper DTOs, validators, and OpenAPI docs.

It is the Backend contract that the admin UI (`web-capability-admin`) consumes.

## What Changes

- **HTTP endpoints** under `apps/api/internal/interfaces/http/` (routes + controllers)
  for capability management (PRD §59, §60, §107):
  - `GET  /capabilities` — list registry (tenant-scoped, filtered).
  - `POST /capabilities` — register a capability (name, description, provider_type,
    provider_id, input/output schema, status, version, credential_reference).
  - `GET  /capabilities/{id}` — read one capability.
  - `PATCH /capabilities/{id}` — update capability (incl. status/version).
  - `DELETE /capabilities/{id}` — disable/remove a capability.
  - `GET /capabilities/{id}/bindings` — list bindings for a capability.
  - `POST /capabilities/{id}/bindings` — bind to tenant/project/workflow/state scope.
  - `DELETE /bindings/{id}` — remove a binding.
  - `POST /capabilities/{id}/test` — invoke in sandbox/mock mode and return the result
    (PRD §2064, §153) to validate schema + provider wiring without live side effects.
- **Application services** `CapabilityService` (PRD §174) orchestrating
  `ICapabilityRepository` + the execution layer.
- **DTOs & validators** per `.agents/guides/api-dto.md` / `api-validator.md`.
- **OpenAPI docs** — split-per-feature under `docs/openapi/` for capability endpoints.
- **Secrets guard** — never return/store `credential_reference` secrets; only the
  reference string (PRD §61, §91).

## Capabilities

### New Capabilities

- `capability/registry-admin`: HTTP administrative CRUD over the Capability Registry and
  its bindings, tenant-scoped with secret-safe DTOs.
- `capability/test-invocation`: an HTTP endpoint to test/simulate capability execution
  in sandbox/mock mode and observe the normalized result.

### Modified Capabilities

- None (new capabilities introduced by this epic).

## Impact

- **`apps/api/internal/interfaces/http/routes/`** — add capability route group.
- **`apps/api/internal/interfaces/http/controllers/`** — add `capability_controller.go`.
- **`apps/api/internal/application/services/`** — add `capability_service.go`.
- **`apps/api/internal/application/dtos/`** — add capability DTOs.
- **`apps/api/internal/application/validators/`** — add capability validators (if the
  pattern uses dedicated validators).
- **`docs/openapi/`** — add `paths/capabilities.json`, `schemas/capability.json`.
- **`apps/api/internal/domain/`** — reused entities/repository from
  `persistence-capabilities-policies`; reused resolver/invoker from
  `mcp-capability-execution`.
- **No** worker, shared packages, docker changes.
- Quality gate: `go build ./...`, `go vet ./...`, `go test ./...`; OpenAPI validates.

## Non-Goals

- The execution **internals** (resolver, security chain, mock) — in
  `mcp-capability-execution`.
- MCP server/tools for the LLM — in `mcp-server-runtime`.
- Policy CRUD UI (only registry + bindings + test here; policy management may be a later
  slice).
- Frontend admin UI — separate proposal `web-capability-admin`.

## Dependencies

- `persistence-capabilities-policies` (`ICapabilityRepository`, entities).
- `mcp-capability-execution` (resolver/invoker + mock provider for the test endpoint).
- Epic #4.

## Notes

- The `docs-openapi` skill covers the split-per-feature OpenAPI format under
  `docs/openapi/`.
