## Context

Epic #4 (MCP & Integrasi) needs an admin control plane over the Capability Registry,
bindings, and sandbox testing. The data layer (`persistence-capabilities-policies`)
provides `ICapabilityRepository`; the execution layer (`mcp-capability-execution`)
provides the resolver/invoker + mock provider. This slice exposes those over the Echo
HTTP API with DTOs, validators, and split OpenAPI docs. See proposal.md for motivation.

## Goals / Non-Goals

**Goals:**
- Expose tenant-scoped CRUD for capabilities and bindings.
- Expose a sandbox/mock test-invocation endpoint reusing the execution layer.
- Keep secrets out of DTOs (only `credential_reference`).
- Document the API with split OpenAPI docs.

**Non-Goals:**
- Execution internals (already in `mcp-capability-execution`).
- MCP server/tools for LLM (in `mcp-server-runtime`).
- Policy CRUD UI (later slice).
- Frontend admin UI (separate proposal `web-capability-admin`).

## Decisions

### D1. Route/controller/service layering
Follow the existing Echo + Clean Architecture pattern in `apps/api`:
```
routes.go                        → register /capabilities group (auth + tenant-scoped)
capability_controller.go         → parse request, call service, format response
capability_service.go            → orchestrate ICapabilityRepository + resolver/invoker
dtos/capability_*.go             → request/response mapping
```
No business logic in controllers; no DB access in services (PRD §74, §73).

### D2. Tenant scoping
Every query and mutation filters by the tenant from the auth session middleware
(PRD §4, §96). DTOs carry `tenantID` from the authenticated context, never from the body.

### D3. Secrets
DTOs expose only `credential_reference`. A capability create/update accepts the
reference string; the resolved secret is never returned, stored, or logged (PRD §61,
§91). The controller strips any raw secret field before responding.

### D4. Test-invocation endpoint
`POST /capabilities/{id}/test` reuses `capability.Invoker` in **mock/sandbox** mode
(PRD §2064). The service forces the mock provider and returns the normalized
`InvocationResult` (with `fromMock=true`) or a classified `CapabilityError`. It never
hits a live MCP provider.

### D5. Provider types & status
Reuse typed constants from the domain entities (`ProviderType`, `CapabilityStatus`,
`BindingScopeType`, `BindingPermission`) — never strings at the boundary (enum discipline).

### D6. OpenAPI
Split-per-feature docs under `docs/openapi/`:
- `paths/capabilities.json` — all capability + binding + test paths.
- `schemas/capability.json` — Capability, CapabilityBinding, InvocationResult,
  CapabilityError.
Updated via the `docs-openapi` skill.

## Risks / Trade-offs

- [Test endpoint could mask live integration issues] → Sandbox is explicit (`fromMock`);
  live execution is a separate, privileged path.
- [Binding deletion removes policy unexpectedly] → Scoped to tenant; UI confirms deletes.
- [Schema JSONB free-form] → Validated against a JSON Schema at the service boundary.

## Migration Plan

- Land after `mcp-capability-execution` (and `persistence-capabilities-policies`).
- Add route group behind existing auth middleware; additive, no breaking changes.
- Rollback: remove the route group; no data migration.

## Open Questions

None.
