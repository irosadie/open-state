## Why

When an LLM sees both State MCP and a provider MCP, it can invoke the provider directly and skip OpenState's gate. A production state-control guarantee requires one execution path that checks state policy before any external provider call is sent.

## What Changes

- Add an OpenState MCP gateway execution mode that exposes state-authorized provider actions through the State MCP endpoint.
- Resolve every invocation from the current tenant, project, workflow instance, state requirement, and project MCP binding; never accept a provider URL, alias, or arbitrary tool name from the LLM as authority.
- Enforce required capability, scope, input schema, idempotency, timeout, result normalization, evidence persistence, and transition eligibility around each forwarded call.
- Make the gateway tool surface expose only state-authorized provider actions, while direct-two-MCP mode remains explicitly advisory and non-enforcing.
- Document the secure deployment profile in which the LLM connects only to OpenState and provider endpoints/credentials remain internal.

## Capabilities

### New Capabilities

- `mcp/enforced-gateway`: State-authorized forwarding of registered provider MCP tools with evidence and auditability.

### Modified Capabilities

- `mcp/client-adapter`: The real MCP client resolves a project connection and invokes only a gateway-authorized discovered tool.
- `mcp/orchestrator-tools`: Provider execution becomes a state-gated gateway operation; direct result reporting remains a compatibility path with no bypass guarantee.
- `mcp/server-runtime`: The State MCP advertises the secure gateway mode and does not expose raw provider credentials or endpoints.

## Impact

- Gateway application service, MCP transport clients, capability execution port, runtime authorization, retry/idempotency/evidence path, tests, documentation, and deployment configuration.
- Depends on Phases 3–5.
- Changes the recommended production deployment from two directly visible MCP endpoints to one State MCP endpoint plus internal provider connections.

## Non-goals

- Requiring third-party MCP servers to implement OpenState-specific protocols.
- Replacing the existing direct two-MCP development/advisory mode.
- Full OAuth refresh-token lifecycle, network egress controls, and operational observability (Phase 7).
