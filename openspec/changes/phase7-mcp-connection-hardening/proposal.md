## Why

Registered and forwarded external MCP connections hold sensitive credentials and make outbound network calls. Production use requires a complete security and operations envelope beyond the basic registry and gateway behavior.

## What Changes

- Add secure bearer-secret lifecycle, OAuth authorization/refresh lifecycle, and credential rotation/revocation for project MCP connections.
- Enforce outbound endpoint protections: HTTPS policy, private-network/SSRF controls, DNS revalidation, redirect policy, allowlists, and bounded STDIO execution.
- Add connection health, structured redacted audit records, metrics, traces, rate limits, circuit breaking, and operator-visible diagnostics.
- Define credential access boundaries so plaintext secrets never enter API responses, State MCP tools, workflow definitions, logs, or browser state.
- Provide deployment guidance for secure gateway and direct/advisory modes, including key rotation and incident response.

## Capabilities

### New Capabilities

- `ops/mcp-connection-security`: Credential, egress, and execution-boundary requirements for external MCP connections.
- `ops/mcp-gateway-operations`: Health, observability, rate-control, and failure-handling requirements for the MCP gateway.

### Modified Capabilities

- `ops/secrets-management`: External MCP bearer and OAuth credentials become managed secret references with lifecycle controls.

## Impact

- Secret storage/integration, OAuth callbacks and token refresh, network policy, worker/runtime behavior, audit/metrics/tracing, admin diagnostics, deployment docs, and security tests.
- Depends on Phases 3–6.

## Non-goals

- Providing a generic identity provider or OAuth authorization server for third-party providers.
- Supporting unreviewed arbitrary shell commands for hosted STDIO execution.
- Guaranteeing availability of an external provider beyond classifying, retrying, and surfacing its failure.
