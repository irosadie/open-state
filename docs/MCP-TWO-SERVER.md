# Two-MCP host contract

OpenState uses two MCP connections:

```text
LLM host ── State MCP (control plane, :8030)
LLM host ── Provider MCP (data plane, :8031 in development)
```

State MCP is the source of truth for intent, workflow, current state, required
context, allowed transitions, and capability requirements. The provider MCP is
the source of third-party data and side effects. State MCP does not proxy the
provider connection.

The host registers a stable alias for every provider connection. A development
configuration may look like this:

```json
{
  "openstate": {
    "url": "http://localhost:8030/mcp",
    "authorization": "Bearer osk_..."
  },
  "padel-provider-mock": {
    "url": "http://localhost:8031/mcp"
  }
}
```

Workflow definitions store only the provider alias and concrete tool name. They
never store a provider URL or secret. The host should verify the alias identity
and exact tool through the provider's `initialize` and `tools/list` responses.

The required runtime order is:

1. Call State MCP `list_intents`, then `resolve_intent`.
2. Start/find the workflow and call `get_current_state`.
3. For every `requiredCapabilities` entry, call the exact `tool` on the already
   connected `providerServer` alias.
4. Report the normalized result with State MCP `report_capability_result`.
5. Call State MCP `propose_event` only after the report is accepted.

A provider response by itself does not authorize a transition. State MCP keeps
an explicit evidence record scoped to tenant, project, instance, current state,
logical capability, and idempotency key. Missing, failed, stale, undeclared, or
cross-tenant evidence is rejected.

For a local end-to-end run, set `STATE_MCP_TOKEN`, `INSTANCE_ID`, `EVENT_TYPE`,
and optionally `PROVIDER_ARGS_JSON`, then run:

```bash
bash ./scripts/mcp-two-server-smoke.sh
```

Use `bun run mcp:check` to catch a stale process or a wrong server listening on
8030/8031.
