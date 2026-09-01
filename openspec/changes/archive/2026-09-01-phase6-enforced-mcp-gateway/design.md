## Context

Direct two-MCP mode from Phase 2 lets an LLM invoke a provider before State MCP can
evaluate it. By Phase 5, OpenState has an owned project connection, verified tool, and
capability binding. Those inputs make it possible to enforce provider execution without
requiring third-party providers to understand OpenState.

## Goals / Non-Goals

**Goals:**

- Make one State MCP endpoint sufficient for a production LLM integration.
- Ensure provider calls are authorized from current workflow state, not LLM routing.
- Preserve core-engine independence from MCP SDK/transport details.

**Non-Goals:**

- Removing direct/advisory mode, modifying third-party MCP servers, or completing
  OAuth/egress hardening before its dedicated Phase 7.

## Decisions

### D1 — Add gateway mode, retain advisory mode

Introduce an explicit runtime mode. In `gateway` mode, the LLM connects only to
OpenState; provider endpoint and credentials remain server-side. In `advisory` mode,
existing direct State+provider connections keep working but State MCP instructions and
documentation label the mode non-enforcing.

### D2 — Reuse state-authorized capability invocation surface

Evolve the existing authorized `invoke_capability` path into the gateway execution
surface rather than exposing arbitrary provider passthrough tools. Its input contains
instance/context/capability input only. The application service loads the current state,
checks requirement/security/binding/schema/idempotency, and sends the resolved target
to the `CapabilityProvider` port.

**Alternative considered:** one mirrored MCP tool per provider tool. Rejected because
tool catalogs would become global/unbounded and reintroduce selection bypass.

### D3 — Enforce before outbound call and persist evidence after normalization

The gateway transaction flow is:

```text
authenticate → resolve tenant/project/instance/current state
→ authorize logical capability → resolve verified binding
→ validate input/idempotency → invoke provider adapter
→ normalize/validate result → persist evidence/context/audit
→ allow eligible propose_event transition
```

The adapter receives an internal resolved connection/tool object, never raw client
parameters. Transition enforcement reuses Phase 2 execution evidence.

### D4 — Failure is classified and fail-closed

Unavailable, timeout, invalid-input, invalid-output, unauthorized, disabled-tool, and
idempotency-conflict outcomes return safe structured failures. No failure records
successful evidence or mutates workflow state. Retry ownership remains in the existing
capability policy path and is restricted further by Phase 7.

### D5 — Keep provider session lifecycle behind an adapter

An `MCPConnectionClient`/adapter owns HTTP/SSE/STDIO session setup, credential lookup,
and `tools/call`. The application layer depends on a provider execution port and can be
tested with fakes. Connection reuse/pooling is implementation-local and cannot weaken
project or state authorization.

## Risks / Trade-offs

- **[Gateway adds latency and becomes a critical path]** → bounded timeouts, reusable
  sessions where safe, idempotency, metrics, and circuit controls in Phase 7.
- **[Provider schema drift after publish]** → validate against last discovery and the
  normalized capability output; fail closed and surface binding health.
- **[Legacy clients expect direct result reporting]** → keep it as an advisory
  compatibility path, but never claim it prevents bypass.

## Migration Plan

1. Add a runtime gateway-mode configuration defaulting to advisory for compatibility.
2. Implement resolved-binding invocation in the application and infrastructure adapter.
3. Wire State MCP `invoke_capability`, evidence, and safe response projections.
4. Add gateway integration tests with provider mock for success, denied call, timeout,
   duplicate idempotency key, invalid output, and transition gate.
5. Deploy gateway mode with providers reachable only from OpenState; roll back by
   returning to advisory mode without deleting binding/evidence records.
