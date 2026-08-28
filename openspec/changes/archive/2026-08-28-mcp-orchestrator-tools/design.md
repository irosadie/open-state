## Context

Epic #4 needs the MCP server to expose the complete orchestrator contract to 3rd-party
LLM/RAG clients (PRD 170). `mcp-server-runtime` already ships the server + several
tools; `mcp-capability-execution` ships the capability resolver/invoker. This slice adds
the remaining orchestrator tools, the LLM context compiler (PII-redacted, PRD 22/90),
and the `RAGProvider` (Retrieve) port (PRD 171). See proposal.md for motivation.

## Goals / Non-Goals

**Goals:**
- Expose lifecycle + history + event-proposal MCP tools.
- Compile minimal, PII-redacted per-turn context (available/missing, memory/workflow).
- Define the `RAGProvider` (Retrieve) port.

**Non-Goals:**
- Built-in `LLMProvider` (platform-initiated LLM) — out of scope (PRD 170).
- Concrete RAG backend / retrieval index — only the port.
- MCP transport rework (in `mcp-server-runtime`).

## Decisions

### D1. Tools are thin orchestrator adapters
Each new MCP tool handler stays thin: it parses the tool input, resolves the tenant from
the authenticated context, calls an application use case, and formats the result. No
business logic in handlers (mirrors the existing MCP tool pattern).

### D2. Compose existing repositories + engine, no new persistence
Lifecycle/history/replay tools reuse `IWorkflowRepository`, `IInstanceRepository`, and
`IEventRepository` (from `persistence-*`) and the engine (from `engine-core`). No new
tables. `replay_workflow` replays recorded events in sequence order (PRD 52) using the
event history.

### D3. Context compiler is a separate application service
`ContextCompiler` in `internal/application/` composes: current state + workflow data
(via repositories), persistent memory (via `IContextRepository` memory methods), and
RAG retrievals (via `RAGProvider`). It returns a normalized `CompiledContext` DTO with
`available` / `missing` / `memory` / `workflow` sections. PII redaction runs last.

### D4. PII redaction via a port
A `Redactor` port (`Redact(ctx, string) (string, error)`) is injected into the
compiler. Default implementation masks configured PII patterns; operators can supply
their own (PRD 90, 169).

### D5. RAGProvider is a domain port
`internal/domain/rag` defines:
```go
type RAGProvider interface {
    Retrieve(ctx context.Context, query string) (*Retrieval, error)
}
type Retrieval struct {
    Text     string
    Metadata map[string]any
}
```
The engine/compiler depends only on this port. A no-op / stub is wired at the
composition root until a concrete backend lands (PRD 171).

### D6. Tool list (scope of this change)
- Read: `get_current_state`, `get_allowed_capabilities`, `get_workflow_instances`,
  `get_history`.
- Plan/act: `propose_event`, `start_workflow`, `suspend_workflow`, `resume_workflow`,
  `cancel_workflow`.
- Replay: `replay_workflow`.

## Schema Outline

No new DB tables. All tools operate on existing entities (workflow, workflow instance,
state instance, event).

## Risks / Trade-offs

- [Tool surface grows] → Handlers are thin and delegate to use cases; tool registry is
  additive. Documented under `mcp/server-runtime`.
- [replay_workflow complexity] → Keep it deterministic event replay over the immutable
  event history (PRD 52); no outbox/interop changes.
- [PII redaction false positives] → Default redactor is conservative and configurable;
  audit via `backend/persistence/audit-adapter`.

## Migration Plan

1. Branch `feature/epic4-mcp-orchestrator-tools`.
2. Add `internal/domain/rag` (RAGProvider + Retrieval).
3. Add `internal/application/context-compiler` (compiler + redactor port).
4. Add orchestrator use cases (lifecycle, propose-event, history/replay).
5. Register MCP tool handlers on the existing MCP server.
6. Wire `RAGProvider` (stub) + `Redactor` (default) at the composition root.
7. `go build ./...`, `go vet ./...`, `go test ./...`.
8. Smoke: invoke each tool against a seeded instance; verify context compiler output +
  PII redaction; verify replay reproduces state.
9. PR → review → merge.

**Rollback**: tool registration + new application/domain files are additive; no data
migration. Remove the handlers to disable the tools.

## Open Questions

None — the tool contract and PRD refs are fixed by MAIN_PRD and `mcp-server-runtime`.
