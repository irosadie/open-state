## Context

See `proposal.md` for motivation. The API fixture catalog currently contains static response payloads for all three provider domains, while `apps/mcp-provider-mock` can load a single scenario whose tools are registered dynamically. The mock already supports static tool results as well as padel-specific availability and booking behavior.

## Goals / Non-Goals

**Goals:**

- Make the provider mock the runtime source for the migrated padel, food-order, and doctor fixture responses.
- Keep scenario selection explicit and tool discovery isolated by domain.
- Verify the public Streamable HTTP MCP boundary without an SDK client.

**Non-Goals:**

- Make all fixture responses stateful or reproduce a production provider API.
- Change State MCP, capability authorization, or the API's unrelated generic fixtures.

## Decisions

### Use one scenario JSON file per provider domain

Each domain receives an independently selectable scenario containing its declared tools, input schemas, and static payloads. This makes discovery representative of a third-party provider rather than exposing a combined development super-server.

Alternative considered: one fixture containing every tool. Rejected because clients would not validate which tools belong to a provider connection.

### Preserve payloads and expose them through generic static tool results

Existing JSON data is carried into scenario result payloads unchanged where possible. The mock's generic static operation returns the configured payload after JSON-schema validation; existing padel dynamic booking remains in its dedicated scenario.

Alternative considered: write separate domain handlers. Rejected because static catalogs do not need domain-specific mutation logic.

### Use curl for protocol smoke tests

A shell test starts the mock on a temporary local port and issues JSON-RPC `initialize`, `tools/list`, and `tools/call` requests using `curl`. It asserts protocol-visible results with standard command-line JSON parsing and always stops the spawned process.

Alternative considered: SDK-only test coverage. Rejected because the request explicitly needs an HTTP-level verification path independent from an MCP SDK.

## Risks / Trade-offs

- [Static fixture input schemas could be too permissive] → Declare required input fields only when the original capability contract requires them; use an empty object schema for fixture responses that have no meaningful request input.
- [Curl session handling differs across MCP transports] → Run the mock in stateless JSON-response mode and send the required MCP protocol version headers in the smoke helper.
- [Removing API fixture keys can break existing tests] → Search and migrate impacted API tests to the mock invocation path before removing only provider-domain entries.

## Migration Plan

1. Add the three provider scenarios and generic static-result behavior needed for their tool catalogs.
2. Migrate API tests that consume the provider-domain fixture keys.
3. Remove only migrated entries from the API fixture catalog.
4. Add and run curl smoke coverage plus existing provider mock and API checks.

Rollback consists of restoring the API fixture entries and selecting the prior padel-only scenario; no persistent data migration is involved.
