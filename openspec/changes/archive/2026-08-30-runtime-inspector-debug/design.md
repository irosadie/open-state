## Context

See `proposal.md` for motivation. The runtime already persists workflow instances,
state instances, events, context, and audit correlation. `OrchestratorService`
already exposes internal queries for instance lists, current state, history,
context replay, and allowed transitions, while the MCP interface consumes several
of them. No HTTP read model or operator UI exposes that information today.

OpenTelemetry trace export is intentionally external infrastructure; it cannot be
the source for product-level Debug View. LLM, RAG, and hosted MCP systems are also
independently deployable and must remain outside this application's ownership.

## Goals / Non-Goals

**Goals:**
- Add tenant-isolated, permission-protected runtime inspection from application
  persistence.
- Record a compact, append-only trace projection for runtime stages observed by
  OpenState.
- Show clear stage provenance and safe, redacted provider metadata in the Admin
  Console.

**Non-Goals:**
- Query an OTLP collector or any LLM/RAG/MCP provider at read time.
- Persist raw conversational or provider payloads for troubleshooting.
- Add live streaming; the first release uses normal query refresh behavior.

## Decisions

### D1: Separate runtime read model from command/MCP surfaces

Add authenticated runtime routes under `/api/runtime/instances`: a paginated list,
instance detail, and a separate debug-trace query. The application service composes
existing instance, event, context, and audit repositories into response DTOs rather
than exposing database records or reusing MCP handlers. `instance:read` protects
list/detail; `debug:read` protects the trace endpoint.

Alternative: expose existing MCP tools to the browser. Rejected because MCP is an
integration surface, not an operator API, and it would blur user authorization and
browser contracts.

### D2: Append-only trace projection owned by OpenState

Add a tenant-scoped `runtime_trace_entries` persistence model with an ordered entry
per observed stage. Each row carries instance id, optional turn id, sequence, stage,
source (`OPENSTATE` or `EXTERNAL_PROVIDER`), status, correlation id, duration,
reason/error codes, provider alias/reference, and sanitized JSON attributes. A
repository port keeps the application/domain independent of PostgreSQL; the pgx
adapter, goose migration, and sqlc queries implement it.

Alternative: reconstruct every Debug View from logs, audit rows, and OTel spans.
Rejected because those sources are incomplete for product semantics, may be retained
externally, and cannot safely represent a deterministic per-turn sequence.

### D3: Instrument integration boundaries; never inspect providers directly

OpenState writes trace entries only when its own orchestration or integration
adapter observes a boundary. For LLM/RAG/MCP providers, adapters may supply a
sanitized result envelope with provider alias, opaque operation/reference id,
duration, status, and allowed summary fields. The recorder has no provider SDK or
credential and cannot issue follow-up requests.

Alternative: let Debug View fetch the full request/response from each provider.
Rejected because LLM, RAG, and MCP systems are third-party/standalone, and that
would increase coupling, credential exposure, latency, and privacy risk.

### D4: Redact before persistence, not only before rendering

All trace and context projection data passes through a shared allowlist/redactor
before it reaches `runtime_trace_entries` or response DTOs. Secrets, credentials,
raw prompts, raw model responses, RAG documents, and values marked sensitive are
omitted or replaced with a stable redacted marker. Reason codes, keys, provider
aliases, duration, and opaque correlation references remain inspectable.

Alternative: store full payloads and hide them in the UI. Rejected because a later
API bug, export, or elevated role could disclose the original data.

### D5: Route-oriented Admin Console composition

Add a thin `/admin/runtime-instances` page plus a detail route. Page-content
components use transaction hooks for list/detail/trace data; private components
render filters, state/context summary, timeline, and Debug View. A missing or
forbidden trace is represented distinctly from an empty trace.

Alternative: embed inspector in State Builder. Rejected because the inspector
operates on persisted instances, not only the draft currently open in a builder.

## Risks / Trade-offs

- [External provider metadata is absent] → render the stage as not recorded; do
  not infer success or failure.
- [Trace volume grows with conversation traffic] → paginate list/timeline queries,
  index tenant/instance/turn/sequence, and introduce retention configuration in the
  implementation without weakening append-only behavior during its retention window.
- [Sensitive data reaches an adapter] → redact before persistence and cover the
  recorder plus API serializers with negative tests.
- [Authorization drift] → centralize route permissions and add role-matrix tests for
  `instance:read` and `debug:read`.

## Migration Plan

1. Add the trace migration, repository port, sqlc queries, and redaction tests.
2. Deploy recorder support before exposing the UI; existing instances simply have no
   trace entries.
3. Add runtime query endpoints and authorization guards, then ship the Admin Console
   routes behind those contracts.
4. Roll back by disabling the new routes/UI and recorder; retained trace rows remain
   append-only and do not affect workflow execution.
