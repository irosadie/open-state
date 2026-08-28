## Why

Epic **#4 (MCP & Integrasi)** defines "keluar saat selesai" as: *LLM bisa terhubung &
menjalankan alur percakapan penuh lewat MCP*. Today the MCP server registers all 14
tools and wires them to the persistence layer (via `mcp-tool-wiring`), but the **runtime
state engine is not wired into the production path**: `propose_event` only appends an
event to history (`OrchestratorService.ProposeEvent`), it does **not** run the engine's
`event → guard → transition` evaluation. Likewise `get_current_state`, `replay_workflow`,
and `get_active_workflow` read/merge persisted rows without executing the engine. The
engine (`engine.NewEngine`) is exercised only in tests and the E2E mock, never in the
production MCP/service path. This means an external LLM cannot actually drive a real
state transition through the platform — the core promise of epic #4 is unmet.

This change wires the domain engine into the application/runtime layer so the MCP
orchestrator tools execute real state-machine transitions against the persistence
adapter (ADR-001), completing the "full conversation flow" for epic #4.

## What Changes

- **NEW** — a persistence→engine adapter that satisfies the engine's `EngineRepositories`
  ports (`ProjectRepository`, `WorkflowRepository`, `InstanceRepository`,
  `EventRepository`) on top of the existing pgx repositories (`IProjectRepository`,
  `IWorkflowRepository`, `IInstanceRepository`, `IEventRepository`). It converts between
  the `entities.*` and `engine.*` domain models (e.g. unmarshalling
  `WorkflowVersion.Definition` JSON into `engine.WorkflowDefinition`, and mapping
  workflow/state instances).
- **MODIFIED** — `OrchestratorService` to hold an optional engine dependency and execute
  the engine for:
  - `ProposeEvent` → run `engine.ProcessEvent` (validate event, evaluate guards, apply
    transition, persist the new state), not just append the event.
  - `GetCurrentState` → return the engine-computed current state + allowed transitions.
  - `ReplayWorkflow` → replay through the engine to reproduce the resulting state.
  - `GetActiveWorkflow` → keep resolving the active instance (already engine-consistent).
- **MODIFIED** — composition root (`apps/api/cmd/mcp-server/main.go` and
  `apps/api/cmd/server/main.go`) to construct the engine + adapter and inject it into the
  orchestrator.
- **No** change to the MCP tool surface/contract (the 14 tools and their handlers stay);
  only the underlying behavior becomes engine-backed.

## Capabilities

### New Capabilities

- `mcp/engine-runtime`: wires the domain state engine into the application/MCP runtime
  path so orchestrator tools (`propose_event`, `get_current_state`, `replay_workflow`)
  execute deterministic state-machine transitions against the persistence adapter.

### Modified Capabilities

- `mcp/orchestrator-tools`: `propose_event` SHALL run the engine's
  `event → guard → transition` evaluation (guarded, prioritized) and persist the
  resulting state, not merely append the event. `get_current_state` and
  `replay_workflow` SHALL reflect engine-computed state.
- `mcp/server-runtime`: the active-workflow and proposal tools SHALL operate on
  engine-backed state transitions.

## Impact

- **`apps/api/internal/infrastructure/`** — new engine adapter (persistence → engine ports).
- **`apps/api/internal/application/services/orchestrator_service.go`** — engine wiring for
  propose/current-state/replay.
- **`apps/api/cmd/mcp-server/main.go`**, **`apps/api/cmd/server/main.go`** — composition root.
- **`apps/api/internal/application/services/orchestrator_service_test.go`** — tests updated
  to cover engine-backed propose/transition/replay (with an in-memory adapter).
- **`docs/openapi`** — only if the HTTP contract changes (likely none; MCP surface unchanged).

## Non-Goals

- Building a concrete RAG backend — RAG/LLM remain 3rd-party (PRD 170); the `RAGProvider`
  port already exists.
- New MCP tools or changing the tool names/schema.
- Retry/suspend engine semantics beyond what the engine already implements.
- Rate limiting, observability (covered by epic #6).

## Dependencies

- `engine-core` (state machine + guard evaluation), `persistence-*` repos, `mcp-tool-wiring`,
  `mcp-orchestrator-tools` specs.
- Epic #2 (engine), #3 (persistence), #4 (MCP).

## Notes

- The engine is domain-pure and already unit-tested (deterministic + golden + E2E). This
  change adds the production wiring and adapter, not new engine behavior.
