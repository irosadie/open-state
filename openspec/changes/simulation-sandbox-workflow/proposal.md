## Why

Epic #5 requires operators to exercise a workflow before publishing it (PRD 53),
but the State Builder currently offers only static browser validation and a direct
publish action. Operators cannot see how ordered events, guard conditions, context,
or allowed capabilities would behave on the exact draft snapshot they intend to
publish.

## What Changes

- **NEW** — an authenticated Builder simulation API that accepts the current workflow
  snapshot, initial context, and an ordered event script; it executes the deterministic
  engine entirely in memory and returns a structured trace.
- **NEW** — sandbox trace semantics for entered states, evaluated transition guards,
  selected/rejected transitions, resulting context, and capability requests. Every
  capability request is marked mock; the simulation never invokes a live MCP/provider
  or persists an instance, event, context, or audit payload.
- **NEW** — a State Builder simulation panel for authoring initial context and events,
  running the draft snapshot, viewing the trace, and highlighting the corresponding
  state/transition in React Flow.
- **MODIFIED** — the Builder API exposes the simulation operation under the same
  authenticated, tenant-scoped surface as other workflow-authoring operations.

## Capabilities

### New Capabilities

- `backend/workflow-simulation`: deterministic, non-persisting execution of a draft
  workflow with mock-only capability planning and a structured simulation trace.
- `web/state-builder-simulation`: State Builder controls and trace viewer for running
  and inspecting a simulation of the current canvas snapshot.

### Modified Capabilities

- `backend/builder-api`: authenticated Builder API gains a tenant-scoped workflow
  simulation endpoint and its request/response contract.

## Impact

- **Backend:** `apps/api/internal/domain/engine/` gains sandbox/trace support built on
  the deterministic transition and guard rules; application DTO/service, workflow
  controller, route wiring, OpenAPI, and tests are extended.
- **Frontend:** shared workflow simulation schemas/types, API route/query/mutation
  support, State Builder store state, toolbar, simulation panel, and focused tests are
  added.
- **Safety:** no schema migration or persistent simulation resource is introduced;
  simulation does not call real capabilities or LLM/MCP systems and does not become a
  required publish gate.

## Non-goals

- Persisting, sharing, replaying, scheduling, or auditing simulation sessions/results.
- Running live MCP providers, LLM inference, webhooks, or arbitrary tenant code.
- Inferring events from free-form conversation text; the operator supplies deterministic
  event names and payloads.
- Replacing the existing publish validation gate, runtime inspector, golden
  conversation tests, version comparison, or capability test-invocation endpoint.
