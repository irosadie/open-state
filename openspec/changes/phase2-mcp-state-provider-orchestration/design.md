## Context

See `proposal.md` for the motivation and scope. The current State MCP already
authenticates the tenant and default project, discovers intents, resolves an intent
to a workflow, and exposes current-state/orchestration tools. Workflow nodes already
carry logical capability names, but the MCP projection omits those capabilities,
the MCP initialize response has no server instructions, and the engine does not
validate required capability execution before applying a transition.

The capability domain already separates logical capabilities from provider
implementations through `CapabilityProvider`. The registry currently has a provider
identifier but no explicit concrete MCP tool field. The provider mock is a separate
Streamable HTTP server and must be treated as a data-plane server that the LLM host
connects independently.

## Goals / Non-Goals

**Goals:**

- Make the State MCP self-describing as the mandatory state controller.
- Return a deterministic, machine-readable provider requirement from intent/state
  responses so an LLM knows the exact provider server and tool to use.
- Keep provider endpoint configuration outside workflow definitions and secrets.
- Make capability execution produce evidence that the engine can use to gate a
  transition.
- Support a two-MCP setup where the LLM host connects to State MCP and one or more
  provider MCP servers, with local development commonly using ports 8030 and 8031.
- Preserve the engine's MCP-agnostic design and the existing capability security
  chain.

**Non-Goals:**

- Dynamic discovery or connection to arbitrary provider URLs supplied by an LLM.
- A production credential vault or provider management marketplace.
- Replacing the existing capability registry, binding model, or workflow builder.
- Making third-party providers depend on OpenState-specific business state.

## Decisions

### D1 — Use a declarative provider requirement contract

Introduce one response-safe projection for provider requirements and reuse it from
`resolve_intent`, `get_current_state`, `get_allowed_capabilities`, and capability
execution results. The projection contains:

```json
{
  "capability": "padel.availability.read",
  "providerServer": "provider-mock",
  "tool": "padel.cek_available",
  "purpose": "Check available padel courts",
  "required": true,
  "beforeTransitions": ["court.available", "court.unavailable"],
  "inputMapping": {
    "date": "context.booking.date",
    "venue_id": "context.booking.venueId"
  }
}
```

The workflow stores the logical capability reference. The capability registry owns
the provider server and concrete tool mapping. This avoids duplicating provider
details in every state and lets several states reuse one capability safely.

**Alternative considered:** put a free-form provider instruction in each state's
`instructions`. Rejected because it is difficult to validate, filter, display, and
enforce, and it encourages the LLM to guess tool names.

### D2 — Use provider server aliases; keep endpoints in the MCP host

Treat the capability's provider ID as a stable server name/alias such as
`padel-provider`, and add an explicit provider tool mapping such as
`padel.check_available`. The MCP client/LLM host resolves that alias to an already
configured MCP connection. In development the host may map it to the provider mock
at `http://localhost:8031/mcp`; production can map the same alias to another endpoint
without changing workflow definitions.

The State MCP returns only the logical provider alias and tool as runtime requirements.
The LLM host must pre-register both MCP servers; State MCP neither forwards the MCP
connection nor establishes a new client connection on the LLM's behalf. Provider
aliases must come from trusted tenant/project configuration, not arbitrary user text.

**Alternative considered:** store the raw URL in workflow JSON or return it as an LLM
connection instruction. Rejected because it couples definitions to deployment, permits
arbitrary endpoint selection, and violates the host-managed two-MCP boundary.

### D3 — Direct provider execution with State-controlled transition evidence

The LLM connects directly to the provider MCP selected by the host and invokes the
exact tool declared by State MCP. The LLM then submits the provider result and its
provider alias/tool metadata to State MCP through a capability-result reporting tool.
State MCP validates the declared requirement, output schema, tenant/project scope,
correlation, and transition eligibility before accepting the evidence.

A direct provider call without an accepted State MCP evidence record is not sufficient
to authorize a transition. Generic MCP does not guarantee a cryptographic receipt for
a call made by an arbitrary LLM host, so the default contract treats the configured
LLM/MCP host as the trusted execution boundary and validates the reported server/tool,
schema, and correlation. A future attested receipt extension can add stronger proof
without changing the state requirement contract.

**Alternative considered:** make State MCP proxy every provider call. Rejected for
this phase because the product contract requires the LLM to connect to two MCP
servers directly; State MCP remains the state control plane, not the connection proxy.

### D4 — Gate transitions using persisted capability evidence

When a capability succeeds for a workflow instance and state, persist a safe
execution record alongside the instance context. The record is keyed by instance,
state, logical capability, and action/idempotency key, and contains provider/tool,
status, normalized result reference, and correlation ID. `propose_event` checks the
current state requirements and the candidate transition before calling the engine's
transition application. Missing or failed evidence returns a deterministic conflict
or validation result and leaves state, context, and event history unchanged.

Evidence must be written atomically with the capability result and transition reads
must use the same tenant/project/instance scope. Existing capability result data can
continue to be persisted into context for downstream guards; the new evidence record
is the explicit enforcement marker and should not expose provider credentials.

### D5 — Add server instructions, but keep dynamic facts in tool results

Add static instructions to the State MCP initialize response for the global protocol:
OpenState controls state, provider requirements are mandatory when declared, and the
LLM must read State MCP before proposing a transition. Do not embed the live intent
catalog or provider list in those instructions. Dynamic intent/state/provider facts
come from structured tool results so they remain tenant/project and workflow scoped.

### D6 — Make provider mock verification part of the two-MCP contract

The provider mock remains independent and scenario-driven. Its own initialization and
`tools/list` response must expose only the active scenario's tools, including English
descriptions and schemas. Development startup and curl tests must verify:

```text
State MCP     → host-configured State MCP server
Provider mock → host-configured provider server alias
```

The local test configuration may use `8030` for State MCP and `8031` for the provider
mock, but the contract test must assert server identity/alias and tool discovery rather
than hard-coding those URLs into workflow state.

## Risks / Trade-offs

- **[Existing capability rows have no concrete tool mapping]** → Backfill the tool
  name from the logical name only where the mapping is unambiguous; require explicit
  configuration for ambiguous or missing mappings.
- **[Persisted evidence introduces a schema/data migration]** → Add a backward-
  compatible migration, preserve existing context records, and make pre-migration
  instances fail closed only for states that declare required capabilities.
- **[Provider calls add network latency]** → Reuse initialized MCP sessions where safe,
  enforce per-capability timeout, and expose classified unavailable/timeout errors.
- **[Direct provider calls may confuse LLM clients]** → State instructions and tool
  descriptions explicitly identify State evidence reporting as the transition-
  authorizing path; provider tools remain discoverable but are not sufficient alone.
- **[Stale local processes can bind the wrong port]** → Add health/readiness identity
  checks and two-MCP curl smoke tests that assert server names and expected tool sets.
- **[Provider output may not match state guards]** → Validate the normalized output
  schema before recording successful evidence and return a classified validation
  result when it does not.

## Migration Plan

1. Add the capability provider-tool mapping and execution-evidence model with a
   backward-compatible database migration and regenerated sqlc/types.
2. Add provider requirement projections, result reporting, and server instructions to State MCP without
   changing existing authentication semantics; default project remains derived from
   the API key.
3. Document the host-side provider alias configuration and map the local
   `provider-mock` alias to the provider mock endpoint only in development.
4. Update seeded padel, doctor, and food workflows/capabilities with explicit
   provider server/tool metadata.
5. Add engine/service enforcement and tests for success, missing evidence, provider
   unavailable, failed provider result, and duplicate side effects.
6. Add provider mock contract tests and curl smoke tests for both MCP endpoints.
7. Roll back by disabling required capability gating for an affected deployment while
   retaining the additive metadata/evidence columns; do not delete existing context
   or audit records.

## Open Questions

- Whether a future production deployment should expose provider endpoint hints in
  State MCP responses at all, or keep endpoints exclusively in the LLM host config.
