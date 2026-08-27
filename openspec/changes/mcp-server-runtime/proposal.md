## Why

Epic **#4 (MCP & Integrasi)** requires exposing the Orchestrator's runtime to external
AI/LLM systems over the **Model Context Protocol**. The PRD specifies an MCP server that
lets the LLM resolve intents and read workflow state (PRD §40.1), read compiled context
(PRD §22), and invoke authorized capabilities (PRD §153). MCP must remain **standalone**
(PRD §20) and the core engine MCP-agnostic (PRD §172, §2559). A real MCP **client**
implementation of `CapabilityProvider` is the primary production integration (PRD §2201),
distinct from the mock provider defined in `mcp-capability-execution`.

This second slice — **mcp-server-runtime** — builds the actual MCP server binary and the
MCP client adapter, plus the secrets resolution for `credential_reference` (PRD §61). It
stands on the execution layer (`mcp-capability-execution`) and the context/engine layers.

## What Changes

- **NEW Go app/module** `apps/mcp-server/` (standalone binary) exposing an MCP server
  over **Streamable HTTP** for the LLM/RAG integration (per architecture notes).
- **MCP tools** exposed to the LLM:
  - `resolve_intent` / `get_active_workflow` — return classified intent + workflow +
    current state to the LLM (PRD §40.1, §1684).
  - `get_context` — return compiled runtime context (tenant, active workflow, current
    state, purpose, instructions, available/missing context, allowed events/transitions,
    available capabilities, relevant memory) (PRD §22).
  - capability invocation tool — lets the LLM request an authorized capability; the
    server runs the security chain and returns the normalized result (PRD §153).
- **MCP client adapter** `apps/api/internal/infrastructure/capability/mcp_provider.go` —
  a real `CapabilityProvider` implementation using the MCP SDK over Streamable HTTP
  (replaces mock in production; mock remains default in sandbox per PRD §2064).
- **Secrets resolution** — resolves `credential_reference` to the actual credential from
  secure infrastructure (env / secret manager) (PRD §61); never stored in workflow defs
  or logged (PRD §91).
- **Tool authorization** — each tool returns only capabilities allowed by
  tenant/project/workflow/state/policy; never the full global registry (PRD §106, §3309).
- **Idempotency & outbox integration** — MCP-invoked side effects support idempotency
  (PRD §64); emitted events flow via the event/outbox mechanism.
- **Docker readiness** — Dockerfile for the MCP server (PRD §176/§177 deployment).

## Capabilities

### New Capabilities

- `mcp/server-runtime`: a standalone MCP server exposing intent resolution, active
  workflow, compiled context, and authorized capability invocation to the LLM.
- `mcp/client-adapter`: a real MCP `CapabilityProvider` implementation (Streamable HTTP)
  with credential resolution for production execution.

### Modified Capabilities

- None (new capabilities introduced by this epic).

## Impact

- **`apps/mcp-server/`** — new Go module/binary (MCP server, tool handlers, auth filter).
- **`apps/api/internal/infrastructure/capability/`** — add `mcp_provider.go` (client
  adapter) + `secrets.go` (credential resolution).
- **`apps/api/internal/domain/capability/`** — reused port from `mcp-capability-execution`;
  no change.
- **`packages/go-shared/`** — reused for `DomainError`; no change.
- **`apps/worker/`** — no change in this proposal (event/outbox publisher exists
  elsewhere); MCP-emitted events may reuse it.
- **`docker/`** or root `Dockerfile` context — add MCP server image definition.
- Quality gate: `go build ./...`, `go vet ./...`, `go test ./...`; MCP server boots and
  exposes declared tools.

## Non-Goals

- Capability **execution** internals (resolver, security, retry, mock) — already in
  `mcp-capability-execution`.
- HTTP **admin** endpoints for capability management — separate proposal
  `capability-admin-api`.
- LLM multi-provider abstraction & structured-output classification — separate concern
  (kept out to keep this slice focused on MCP).
- RAG integration — out of scope (PRD §19, RAG standalone).
- Persistence of MCP audit events — covered by `persistence-audit-adapter`.

## Dependencies

- `mcp-capability-execution` (`CapabilityProvider` port, resolver, invoker, mock).
- `engine-context-resolver` (compiled context for `get_context`).
- `engine-domain-core` (intent resolver / active workflow state).
- Epic #4.
