## Context

See `proposal.md` for motivation and the delta specs for the behavior contract.
The State Builder holds the complete canvas definition in a Zustand store and already
uses the engine's event/guard semantics indirectly at runtime. The existing engine
has an in-memory replay repository set for replaying a persisted instance, but it
cannot execute an arbitrary current draft or expose per-candidate guard decisions.
The capability registry has a deterministic mock provider, while state capabilities
are declarative names rather than executable transition actions.

## Goals / Non-Goals

**Goals:**

- Execute the exact current client snapshot in an isolated, deterministic engine
  sandbox and return an actionable trace.
- Reuse production event selection and guard semantics so simulation cannot drift from
  runtime behavior.
- Make the result inspectable in the State Builder without saving the draft.

**Non-Goals:**

- Add a database table, background worker, retained simulation session, or migration.
- Execute a provider (including the existing mock provider) as part of a state
  capability; this model has no capability-action payload to execute.
- Make a prior successful simulation a publish prerequisite. PRD 54 validation remains
  the publishing gate; an untested branch is not necessarily an invalid workflow.

## Decisions

### D1 — Use a snapshot endpoint, not a persisted workflow endpoint

Add `POST /api/workflows/simulate` with `{ definition, initialContext?, events? }`.
The endpoint is authenticated and tenant-scoped via `X-Tenant-ID`, but accepts the
canvas definition directly and does not require a workflow ID or version. It assigns
only an ephemeral project identity within the sandbox, so an unsaved/imported draft
can be tested without querying or creating the tenant's default project.

The response is `{ data: SimulationResult }`, with an entry snapshot, ordered trace
steps, final state/context, and a terminal/rejected outcome. The request is read-only
by design, so it has no idempotency record and no audit write.

- **Alternative considered:** `POST /api/workflows/:id/simulate` loading the draft from
  storage. Rejected because auto-save is asynchronous and simulation must faithfully
  exercise the current unsaved canvas.
- **Alternative considered:** browser-only simulation. Rejected because it would
  duplicate the Go guard/priority engine and eventually drift from production runtime.

### D2 — Add a first-class engine sandbox with trace-aware selection

Add a domain-level simulation entry point that receives an already decoded
`engine.WorkflowDefinition`, initial context, and ordered simulation events. It builds
fresh in-memory repositories (reusing/refactoring the replay repository helpers),
starts an instance, seeds its context, then processes the supplied events in order.
No repository passed to the production engine is reused.

Factor candidate guard evaluation into a shared helper returning both the selected
transition and per-candidate outcomes. `ProcessEvent` continues to use that helper,
and sandbox execution converts it into response-safe trace data. A trace exposes only
structured facts: transition id, event, priority, guard pass/fail or validation error,
and state/context snapshots. It does not expose internal reasoning.

The sandbox stops at the first rejected event and returns the completed prefix plus a
structured rejection. Event IDs are generated from their sequence and idempotency keys
are omitted, keeping repeat runs deterministic. The result intentionally contains no
random session or runtime-instance ID.

- **Alternative considered:** call `Engine.ProcessEvent` and infer trace output after
  every call. Rejected because failed candidates and their guard results would be lost.
- **Alternative considered:** copy the event-selection algorithm into a simulation
  service. Rejected because two implementations of priority/guard rules would drift.

### D3 — Capabilities are planned mock requests, never invocations

When the sandbox enters the entry or target state, it copies that state's declared
capability names into the trace as `mock: true` request plans. It never resolves a
binding, runs `CapabilityService.TestInvoke`, or calls a provider. This exactly models
the capability-request visibility required by PRD 53 without claiming a provider
result or creating a risk of accidental side effects.

- **Alternative considered:** call the existing `MockProvider` for each name. Rejected
  because state capability declarations do not supply an action/payload, and such a
  call would misleadingly present a provider result as workflow behavior.

### D4 — Keep transport mapping in the application/interface layers

Add simulation DTOs alongside the workflow Builder DTOs, a `SimulationService` that
decodes/validates the snapshot and delegates to the domain sandbox, and a controller
method plus route registration. The service maps UI-only layout fields out of the
definition, applies the ephemeral sandbox project ID, and maps domain trace values to
the HTTP contract. The controller remains responsible only for header/body parsing and
the standard `{ data }` response envelope.

Add the endpoint and schemas to split OpenAPI documentation and regenerate the checked
OpenAPI artifact. Tests cover DTO validation, tenant header handling, no persistence,
priority/guard trace, rejection, and mock capability planning.

### D5 — Model simulation as transient State Builder UI state

Add shared Zod request/response schemas and response types, an API route constant, and
a React Query mutation. The State Builder store owns only transient simulation form,
result, request status, selected trace step, and stale flag; none is written by the
draft bridge or included in export/import. The canvas passes a cloned, materialized
snapshot to the mutation.

The toolbar opens a panel with JSON editors for initial context and event payloads.
Client parsing prevents malformed JSON requests. Existing validation errors are shown
before run; the API remains the defensive authority for executable input. Selecting a
successful trace step sets transient canvas highlighting/focus for its source node,
selected edge, and target node. Any tracked workflow mutation clears focus and marks
the result stale.

- **Alternative considered:** store the result with the draft for later review.
  Rejected because it adds retention, schema, privacy, and staleness concerns outside
  this slice.

### D6 — Bound the sandbox and keep its data private to the response

The API will reject empty/non-executable definitions and malformed script entries and
will apply a conservative configured maximum number of event steps to bound CPU and
response size. Context and event payloads are returned only in the authenticated
response and are not logged or audited by simulation code. General request limits and
authentication middleware continue to protect the route.

## Risks / Trade-offs

- **Client and server validation can differ** → The UI treats client validation as
  guidance; the backend still rejects non-executable input and engine simulation is
  authoritative for runtime outcomes.
- **A capability plan is not a capability result** → The trace UI explicitly says
  mock/sandbox request and does not show a completed provider invocation.
- **Long scripts can produce large context snapshots** → Bound event count and use a
  finite request-size limit; operators can run focused scenarios.
- **No retained result makes collaboration harder** → Intentional for Phase 2; runtime
  inspector/replay and shareable test scenarios remain later work.

## Migration Plan

1. Add trace-aware sandbox primitives and deterministic domain tests without changing
   production event outcomes.
2. Add DTO/service/controller/route/OpenAPI contract and API tests; deploy as an
   additive endpoint with no migration.
3. Add shared frontend contract, mutation, panel, and canvas focus behavior; cover
   form parsing, stale result, and trace rendering in frontend tests.
4. Run API and web quality gates plus OpenSpec validation. Rollback is safe by removing
   the additive route and UI control; no persisted simulation data requires cleanup.
