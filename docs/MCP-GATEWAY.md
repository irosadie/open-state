# MCP gateway modes

OpenState supports two provider-routing modes while projects migrate from the
direct two-MCP setup to the enforced gateway.

## Advisory mode

Set `MCP_GATEWAY_MODE=advisory` (the default). The State MCP continues to expose
the compatibility `report_capability_result` flow. The LLM host may still have
its own direct connection to a provider MCP, so this mode is useful for local
rollout and comparison but is not a security boundary.

## Secure gateway mode

Set `MCP_GATEWAY_MODE=secure` on the State MCP process. The LLM needs only the
State MCP endpoint and API key. `invoke_capability` accepts the workflow
instance, logical capability, correlation idempotency fields, and capability
input. OpenState derives the current project and state, checks the capability
authorization, resolves the active project MCP binding, and calls the exact
discovered provider tool internally.

Provider endpoint, authentication, connection alias, tool catalog, and raw
provider errors remain server-side. A missing or unhealthy binding, disabled
connection/tool, invalid input/output, provider failure, or duplicate failed
idempotency key fails closed and cannot satisfy the state transition gate.

## Local provider mock

Run the provider mock on `http://127.0.0.1:8031/mcp`, register the connection in
the project, refresh its tool catalog, and bind the logical capability to the
discovered tool. Run State MCP on `http://127.0.0.1:8030/mcp` with
`MCP_GATEWAY_MODE=secure`. The provider mock is then reachable by the OpenState
process, not by the LLM client.

## Migration and rollback

1. Start in advisory mode and create/test project connections.
2. Refresh the catalog and create one binding per external MCP capability used
   by the workflow.
3. Verify the workflow publishes with healthy bindings.
4. Deploy a State MCP instance with `MCP_GATEWAY_MODE=secure` and run the
   capability read/write smoke tests.
5. If rollback is needed, change the State MCP mode back to `advisory` and
   restart it. Bindings and evidence are retained.

OAuth refresh, egress restrictions, connection pooling, and circuit controls
remain part of the connection-hardening phase.
