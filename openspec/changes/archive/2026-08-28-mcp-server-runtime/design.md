## Context

Epic #4 (MCP & Integrasi) requires the runtime to be reachable by LLM systems via MCP
and to execute real MCP providers in production. The capability execution layer
(`mcp-capability-execution`) defines the `CapabilityProvider` port and invoker. This
slice provides the standalone MCP server and the real MCP client adapter + secrets.
See proposal.md for motivation.

## Goals / Non-Goals

**Goals:**
- Run a standalone MCP server (Streamable HTTP) exposing intent/context/capability tools.
- Provide a real MCP `CapabilityProvider` adapter for production execution.
- Resolve `credential_reference` securely (PRD §61) without leaking secrets (PRD §91).
- Keep the core engine MCP-agnostic (PRD §172).

**Non-Goals:**
- Capability execution internals (done in `mcp-capability-execution`).
- Admin HTTP endpoints (separate proposal `capability-admin-api`).
- LLM provider abstraction, RAG, audit persistence.

## Decisions

### D1. MCP server module
New Go module `apps/mcp-server` using an MCP Go SDK (`mark3labs/mcp-go`) that supports
**Streamable HTTP** transport. Tool handlers are thin adapters that delegate to the
existing application services (intent resolver, context resolver, capability invoker).
The server is wired into the API binary's HTTP mux or run as its own listener on the
existing Go binary (modular monolith, PRD §175).

### D2. Tool set
- `resolve_intent` → `engine.IntentResolver` (PRD §40.1).
- `get_active_workflow` → active workflow + current state + allowed events (PRD §1684).
- `get_context` → `engine-context-resolver` compiled context (PRD §22).
- `invoke_capability` → `capability.Invoker` via the `CapabilityProvider` port (PRD §153).
Tool authorization reuses the binding-aware resolver; only allowed capabilities are
advertised (PRD §106, §3309).

### D3. MCP client adapter
`apps/api/internal/infrastructure/capability/mcp_provider.go` implements
`CapabilityProvider` using the same MCP SDK as a client over Streamable HTTP. It maps a
`Capability` (provider_id) to an MCP endpoint + tool, invokes, and normalizes into
`InvocationResult` / `CapabilityError`. This is the primary production integration
(PRD §2201); the mock provider remains the default in sandbox mode (PRD §2064).

### D4. Secrets resolution
A `credential_resolver` port resolves `credential_reference` from env / secret manager /
Vault (PRD §61). Adapter holds resolved credentials in memory only, redacts them from
logs (PRD §91), and never writes secrets to workflow definitions or audit logs.

### D5. Transport choice: Streamable HTTP
Chosen over SSE for lower connection overhead and native HTTP scaling behind ingress
(PRD §177). The MCP SDK supports it; if a provider only speaks stdio, a small
stdio-bridge adapter may be added later without changing the port.

## Risks / Trade-offs

- [External MCP server outage blocks capability] → Timeout + retry policy from the
  invoker (PRD §160, §88); graceful degradation (PRD §179).
- [Secret store unavailability] → Fail closed for that provider; surface classified
  `capability.unavailable`.
- [MCP SDK version churn] → Keep SDK usage confined to the adapter; domain never imports
  it (PRD §2559).
- [Tool surface too large] → Start with the four tools; add more per PRD §40.1 as needed.

## Migration Plan

- Land after `mcp-capability-execution`.
- Add `apps/mcp-server` module and register its routes/transport.
- Flip provider binding from mock to real per tenant as credentials are provisioned.
- Rollback: stop advertising the MCP server routes; revert to mock provider default.

## Open Questions

None — LLM provider abstraction is intentionally out of scope here and does not change
the MCP server design.
