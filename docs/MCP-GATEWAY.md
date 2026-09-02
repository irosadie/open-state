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

Every failed secure `invoke_capability` response is an explicit hard stop with
`ok: false`, `invoked: false`, `hardStop: true`, and `nextAction: "STOP"`.
The client must not search another scope, choose another capability/provider,
call a diagnostic context tool as a fallback, or propose an event. Secure
`start_workflow` and `propose_event` calls require stable idempotency keys;
retries reuse the same key and return the original outcome.

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

## Connection hardening defaults

External MCP connections are registered per project. The API stores only opaque
credential references; bearer values and OAuth access/refresh tokens are handled
by the configured secret store. Use the Admin Console connection page for
credential rotation/revocation, OAuth connect/disconnect, safe handshake tests,
diagnostics, and health reset.

Production defaults are HTTPS-only, port 443, no private or loopback targets,
and no arbitrary STDIO commands. Configure an explicit development policy for
the local provider mock:

```bash
MCP_EGRESS_MODE=development
MCP_EGRESS_SCHEMES=http,https
MCP_EGRESS_PORTS=8031,443
MCP_EGRESS_ALLOW_LOCAL_DEV=true
```

For hosted STDIO, configure `MCP_STDIO_PROFILES_JSON` with deployment-reviewed
profiles. A project selects only the profile name and reviewed argument prefixes;
it cannot submit a shell command or environment overrides.

Example (the command and environment belong to deployment configuration):

```json
{
  "trusted-padel": {
    "command": "/opt/openstate/bin/padel-mcp",
    "args": ["--stdio"],
    "allowedArgPrefixes": ["--region=", "--venue="],
    "env": ["PATH=/usr/bin"],
    "maxArgs": 16,
    "maxRuntimeMs": 30000,
    "maxOutputBytes": 8388608
  }
}
```

Gateway execution applies per-connection timeout, concurrency, token-bucket
rate limiting, idempotency-aware retries, and a circuit breaker. Provider
failures are stored as classified health/audit outcomes with tenant, project,
connection, tool, and correlation identifiers; raw payloads, headers, tokens,
and provider response bodies are not captured.
