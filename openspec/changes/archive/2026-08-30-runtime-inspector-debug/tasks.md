## 1. Trace persistence and runtime contracts

- [x] 1.1 Read the runtime, event, context, audit, authorization, database-schema, and OpenAPI project guides; map the existing repository and DTO conventions.
- [x] 1.2 Define typed runtime trace entities, stage/source/status enums, sanitized attribute envelope, and a tenant-scoped repository port.
- [x] 1.3 Add a goose migration and sqlc queries for append-only runtime trace entries, including tenant/instance/turn/sequence indexes and ordered list queries.
- [x] 1.4 Implement the Postgres trace repository and expose it from the composed adapter without leaking sqlc types outside infrastructure.
- [x] 1.5 Implement shared redaction/allowlist behavior that removes secrets, credentials, sensitive PII, raw prompts/responses, and raw RAG documents before trace persistence.

## 2. Runtime inspection application API

- [x] 2.1 Add runtime-inspector DTOs and an application service that composes instance, state, event, context, audit correlation, and trace repositories into tenant-scoped list/detail responses.
- [x] 2.2 Add an application-owned trace recorder for intent, workflow/state, context, MCP, event, guard, and transition boundaries; accept LLM/RAG/MCP provider metadata only through a sanitized integration envelope.
- [x] 2.3 Add Runtime Inspector HTTP controllers and authenticated routes for instance list/detail guarded by `instance:read` and debug trace guarded by `debug:read`.
- [x] 2.4 Extend the role-permission matrix with `instance:*` for owner/admin and `debug:read` for operator; preserve least privilege for viewer.
- [x] 2.5 Add split OpenAPI paths and schemas for runtime instance discovery, detail, timeline, and debug trace responses.

## 3. Backend validation

- [x] 3.1 Add unit tests for inspector composition: timeline ordering, missing context, deterministic reason codes, and tenant isolation.
- [x] 3.2 Add trace recorder/repository tests for append-only ordering, redaction before persistence, and partial/unrecorded stages.
- [x] 3.3 Add authorization/controller tests proving `instance:read` and `debug:read` are enforced independently.
- [x] 3.4 Add integration-boundary tests proving inspector queries never invoke or authenticate to LLM, RAG, MCP provider, or OTLP systems.

## 4. Frontend contracts and data access

- [x] 4.1 Add shared Zod schemas and response types for runtime instance list/detail, timeline, reason codes, and debug trace entries.
- [x] 4.2 Add API route constants, query keys, and React Query hooks for runtime instance list/detail and debug trace queries with truthful forbidden/error handling.
- [x] 4.3 Add focused schema and hook tests for query parameters, tenant-scoped responses, and forbidden Debug View behavior.

## 5. Runtime Inspector Admin Console

- [x] 5.1 Add the thin `/admin/runtime-instances` route and route content for searchable, filterable runtime instance discovery.
- [x] 5.2 Add the runtime instance detail route with workflow/version summary, current state, sanitized available/missing context, and chronological timeline.
- [x] 5.3 Add Debug View components that render per-turn stage status, reason codes, correlation ids, and clearly labelled external-provider metadata.
- [x] 5.4 Add loading, empty, forbidden, and error states that distinguish unavailable trace data from an unsuccessful external-provider operation.
- [x] 5.5 Add component tests covering timeline order, redaction markers, debug permission denial, and the absence of direct browser calls to external providers.

## 6. Quality gate and documentation

- [x] 6.1 Run `go build ./...`, `go vet ./...`, and `go test ./...` in `apps/api`; verify goose/sqlc regeneration is clean.
- [x] 6.2 Run `bun run lint`, `bun run typecheck`, `bun run test`, and `bun run build` in `apps/web`.
- [x] 6.3 Run `bun run check`, review the change against all Runtime Inspector and Debug Trace scenarios, and update operator/API documentation with the external-provider and redaction boundaries.
