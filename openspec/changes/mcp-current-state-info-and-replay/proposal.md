## Why

The `get_current_state` and `replay_workflow` MCP tools partially implement their specs.
`get_current_state` returns the state + allowed transitions but omits the state's
`purpose` (description) and `instructions` that a 3rd-party LLM needs to interact
correctly (spec `mcp/orchestrator-tools`, PRD 12, 14). `replay_workflow` only merges
event payloads instead of re-driving the events through the engine to reproduce state
(PRD 52). This change completes both to fully satisfy the epic #4 spec.

## What Changes

- **MODIFIED** — `engine.Engine` gains `CurrentStateInfo(ctx, tenantID, instanceID)`
  returning the current node's `purpose` (description), `instructions`, and
  `requiredContext`, derived from the pinned workflow definition.
- **MODIFIED** — `engine.Engine` gains `Replay(ctx, tenantID, instanceID, events)` that
  re-drives the recorded events through a fresh in-memory engine instance to reproduce
  the resulting context and state, without persisting (PRD 52).
- **MODIFIED** — `OrchestratorPort`/`OrchestratorService` expose `CurrentStateInfo` and an
  engine-backed `ReplayWorkflow`.
- **MODIFIED** — the `get_current_state` handler includes `purpose`, `instructions`, and
  `requiredContext`; the `replay_workflow` handler returns engine-reproduced state.
- **No** new MCP tools or tool-name changes.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `mcp/orchestrator-tools`: `get_current_state` SHALL return the state's purpose,
  instructions, and required context; `replay_workflow` SHALL reproduce state by
  re-driving events through the engine (PRD 52).

## Impact

- **`apps/api/internal/domain/engine/state_machine.go`** — `CurrentStateInfo`, `Replay`.
- **`apps/api/internal/domain/engine/`** — in-memory replay helper + tests.
- **`apps/api/internal/application/services/orchestrator_service.go`** — wiring.
- **`apps/api/internal/interfaces/mcp/`** — `OrchestratorPort`, handlers, mocks.

## Non-Goals

- A concrete RAG backend (3rd-party, PRD 170).
- Changing `propose_event` semantics.

## Dependencies

- `mcp-engine-runtime-wiring`, `mcp-get-current-state-transitions` (already merged).
- `mcp/orchestrator-tools` spec.

## Notes

- Implements remaining spec gaps for epic #4; after this, the epic's "keluar saat
  selesai" is fully met.
