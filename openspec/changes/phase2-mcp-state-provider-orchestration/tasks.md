## 1. Contract and domain model

- [x] 1.1 Define the response-safe provider requirement and execution evidence types shared by State MCP, capability execution, and orchestration responses.
- [x] 1.2 Add an explicit provider tool mapping to the capability model, keeping the logical capability name separate from the concrete MCP tool name.
- [x] 1.3 Add validation rules for required provider metadata, including provider server id, tool name, safe purpose, and required flag.
- [x] 1.4 Add unit tests for provider requirement projection, missing mapping, and secret-safe serialization.

## 2. Persistence and admin contract

- [x] 2.1 Add a backward-compatible migration for the provider tool mapping and capability execution evidence storage.
- [x] 2.2 Update sqlc queries, generated models, repository conversions, and domain entities for the new capability metadata/evidence fields.
- [x] 2.3 Update capability create/update/read DTOs, validation, and admin endpoints to accept and return provider server/tool metadata without secrets.
- [x] 2.4 Update capability registry and binding UI forms to display and edit the provider MCP server id and concrete tool name.
- [x] 2.5 Add repository/service tests for tenant-scoped provider mappings and execution evidence lifecycle.

## 3. Provider alias and host contract

- [x] 3.1 Document the host-side provider alias configuration that may map `provider-mock` to `http://localhost:8031/mcp` in development while keeping endpoints outside workflow definitions.
- [x] 3.2 Define the trusted provider alias/tool mapping used by workflow requirements and reject arbitrary provider endpoints or aliases supplied in user payloads.
- [x] 3.3 Add host/provider contract checks that match the declared provider alias and tool against the provider MCP `tools/list` response.
- [x] 3.4 Add tests for alias isolation, missing provider connections, missing tools, and provider discovery failures.

## 4. State MCP instructions and requirement projection

- [x] 4.1 Add OpenState server instructions to the MCP initialize response describing the mandatory State MCP gatekeeper protocol.
- [x] 4.2 Add provider requirement projection to `resolve_intent`, including state-controller metadata and entry-state requirements.
- [x] 4.3 Add provider requirements and pending/satisfied/failed status to `get_current_state`.
- [x] 4.4 Add provider server/tool metadata to `get_allowed_capabilities` and keep tenant/project values derived from authenticated API-key context.
- [x] 4.5 Add MCP contract tests asserting initialize instructions, structured requirement fields, and no credential leakage.

## 5. Execution evidence and state gate

- [x] 5.1 Add a State MCP capability-result reporting operation with safe provider/tool metadata and a correlation identifier.
- [x] 5.2 Persist successful capability evidence against the tenant, project, workflow instance, current state, logical capability, and idempotency key.
- [x] 5.3 Make duplicate side-effect invocations reuse the original normalized result and execution evidence.
- [x] 5.4 Validate required capability evidence before applying a proposed event transition, returning a deterministic requirement-not-satisfied error when missing or failed.
- [x] 5.5 Ensure rejected transitions do not mutate workflow instance state, context, or event history.
- [x] 5.6 Add service/engine tests for satisfied requirements, missing evidence, failed provider results, stale-state evidence, and cross-tenant isolation.

## 6. Provider mock and local two-MCP runtime

- [x] 6.1 Verify the provider mock on port 8031 reports its provider identity and exposes only the active fixture tools through `tools/list`.
- [x] 6.2 Add provider mock contract coverage for English padel, doctor, and food tool descriptions, schemas, reads, and writes.
- [x] 6.3 Add startup health checks that distinguish State MCP on 8030 from Provider MCP Mock on 8031 and fail on an unexpected server identity.
- [x] 6.4 Update local development orchestration so both MCP processes start with the documented ports and stale/wrong listeners are reported clearly.

## 7. End-to-end verification and documentation

- [x] 7.1 Add a curl smoke flow that initializes both MCP sessions, lists State MCP and provider tools, resolves an intent, reads state requirements, and invokes the declared provider tool.
- [x] 7.2 Extend the curl flow to submit the provider result through State MCP and verify the allowed transition succeeds only after evidence is accepted.
- [x] 7.3 Add negative curl/integration coverage for an unconnected provider, an undeclared provider tool, and a transition attempted without required evidence.
- [x] 7.4 Document the LLM host configuration for two MCP connections: State MCP `8030` and Provider MCP Mock `8031`.
- [x] 7.5 Document the production rule that provider endpoints and credentials are configured by the host/runtime, while workflow definitions store only logical provider/tool metadata.

## 8. Quality gate

- [x] 8.1 Run backend unit, integration, build, and vet checks for State MCP, capability execution, persistence, and provider adapter changes.
- [x] 8.2 Run provider mock tests, TypeScript tests, lint, and typecheck for admin capability mapping changes.
- [x] 8.3 Run OpenSpec strict validation and inspect the final diff for scope, secret-safety, and generated-code consistency.
