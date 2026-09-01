## 1. Persistence and access model

- [x] 1.1 Inspect existing project scope, RBAC, audit, capability, and secret-reference conventions before adding the registry.
- [x] 1.2 Add the MCP connection persistence migration with tenant/project scope, alias uniqueness, transport/authentication fields, safe lifecycle state, and additive rollback safety.
- [x] 1.3 Add sqlc queries and regenerate database code for scoped connection CRUD and safe lifecycle updates.
- [x] 1.4 Add domain entities, repository port, and pgx adapter with project isolation checks.
- [x] 1.5 Add project MCP connection read/manage permissions and role/default-policy migration without widening existing access.

## 2. Connection application services

- [x] 2.1 Implement create/update/list/get/delete/enable/disable connection commands with alias and transport validation.
- [x] 2.2 Implement write-only bearer/OAuth configuration handling that stores only a protected credential reference and returns safe status.
- [x] 2.3 Define a transport-neutral MCP handshake/test port and adapters for Streamable HTTP, SSE compatibility, and trusted STDIO profiles.
- [x] 2.4 Implement deliberate connection test with safe status/error classification and no `tools/list` or provider business-tool invocation.
- [x] 2.5 Emit redacted audit events for all registry mutations and verification actions.

## 3. HTTP contract and admin navigation

- [x] 3.1 Add project-scoped MCP connection REST routes, request validation, response DTOs, and authorization middleware.
- [x] 3.2 Document the registry HTTP API and safe authentication field semantics in split OpenAPI docs.
- [x] 3.3 Add the permission-aware MCP Connections destination to the Admin Console navigation.
- [x] 3.4 Add shared Zod schemas, response types, API constants, and React Query hooks for connection operations.

## 4. Project MCP Connections UI

- [x] 4.1 Build the project-scoped connection list with transport, auth status, enablement, and latest test state.
- [x] 4.2 Build the create/edit form with transport-dependent fields and write-only credential inputs.
- [x] 4.3 Add test, enable/disable, and delete controls with confirmation, loading, and error states.
- [x] 4.4 Verify the UI never renders secret values or connections outside the active project.

## 5. Verification

- [x] 5.1 Add backend unit tests for scope isolation, alias conflict, validation, secret redaction, and disabled connection handling.
- [x] 5.2 Add adapter/HTTP integration tests for successful and failed handshake without provider tool execution.
- [x] 5.3 Add frontend tests for permission visibility, form validation, write-only secrets, and lifecycle feedback.
- [x] 5.4 Run migration/sqlc checks, Go tests/vet, web tests/typecheck/lint, and OpenSpec validation.
