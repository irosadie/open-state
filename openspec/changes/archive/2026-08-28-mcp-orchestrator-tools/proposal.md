## Why

Epic **#4 (MCP & Integrasi)** requires the MCP server to expose the full orchestrator
contract so a 3rd-party LLM/RAG client (the primary integration surface, PRD 170) can
drive a conversation from intent to state transition — without the platform calling an
LLM internally. `mcp-server-runtime` already covers the server startup, intent
resolution, active-workflow lookup, context retrieval, capability invocation, and tool
filtering. This slice completes the remaining orchestrator tools, the **LLM context
compiler** (minimal per-turn context with PII redaction, PRD 22, 90), and the
**RAGProvider** abstraction (PRD 171) so the engine can request relevant knowledge.

## What Changes

- **NEW** MCP orchestrator tools (in `mcp-server-runtime` / `mcp-capability-execution`
  surface):
  - `get_current_state` — current state, purpose, instructions, allowed
    events/transitions (PRD 12, 14, 33-34).
  - `get_allowed_capabilities` — capabilities authorized for a state/context (PRD 59-62).
  - `propose_event` — LLM suggests an event; engine validates & transitions (PRD 38).
  - `start_workflow`, `suspend_workflow`, `resume_workflow`, `cancel_workflow` (PRD 25, 42-43).
  - `get_workflow_instances`, `get_history`, `replay_workflow` (PRD 52).
- **NEW** LLM context compiler (`context/compiler.go`): compiles the minimal per-turn
  context (available + missing context, memory/workflow split) and redacts PII before
  returning to the client (PRD 22, 90).
- **NEW** `RAGProvider` abstraction (`domain/rag`): a portable `Retrieve(ctx, query)`
  interface so the State Engine requests relevant knowledge without depending on a
  concrete RAG backend (PRD 169, 171).

## Capabilities

### New Capabilities

- `mcp/orchestrator-tools`: the remaining lifecycle + history + event-proposal MCP
  tools.
- `mcp/context-compiler`: minimal, PII-redacted per-turn context assembly for LLM/RAG
  clients.
- `mcp/rag-provider`: the `RAGProvider` (Retrieve) port for external knowledge lookup.

### Modified Capabilities

- `mcp/server-runtime`: add the orchestrator tool handlers (start/suspend/resume/
  cancel, instances/history/replay, propose_event, get_current_state,
  get_allowed_capabilities).
- `mcp/capability-execution`: reuse the resolver/invoker for `get_allowed_capabilities`
  and `propose_event`.

## Impact

- **`apps/api/internal/domain/`** — add `rag/` (RAGProvider port + Retrieve result) and
  `context/compiler.go` (LLM context compiler + PII redactor port).
- **`apps/api/internal/application/`** — add `context-compiler` service + orchestrator
  use cases (start/suspend/resume/cancel, propose event, history/replay) that compose
  the existing workflow/instance/event repositories and the engine.
- **`apps/api/internal/interfaces/mcp/`** — register the new MCP tool handlers on the
  existing server.
- **No** web, worker, shared-package changes.

## Non-Goals

- Built-in `LLMProvider` (platform-initiated LLM calls) — explicitly out of scope
  (PRD 170); the platform stays LLM-agnostic via MCP.
- A concrete RAG backend / retrieval index — only the `Retrieve` port is delivered here.
- MCP transport/startup rework — already in `mcp-server-runtime`.
- Frontend/UI — separate epic.

## Dependencies

- `mcp-server-runtime` (MCP server + tool registration surface).
- `mcp-capability-execution` (resolver/invoker for capability-authorization tools).
- `persistence-*` slices (workflow/instance/event repositories for lifecycle + history).
- Epic #4.

## Notes

- PRD 170 fixes the boundary: the platform exposes MCP tools; the LLM is a 3rd-party
  client. All tools here are read/plan/transition operations the client invokes — never
  a platform-initiated LLM call.
- PRD 52 (history/replay) and PRD 25/42-43 (lifecycle) drive the orchestrator tools.
