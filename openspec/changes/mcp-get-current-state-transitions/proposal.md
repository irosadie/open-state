## Why

The `get_current_state` MCP tool (spec `mcp/orchestrator-tools`) is required to return
"the state id, purpose, instructions, and the **allowed events/transitions**". The tool
currently returns the instance + state key/status, but omits the allowed
events/transitions, so a 3rd-party LLM does not know which events it may propose next.
This change completes that gap by deriving the allowed transitions from the instance's
pinned workflow definition via the (now wired) runtime engine.

## What Changes

- **MODIFIED** — `engine.Engine` gains an `AllowedTransitions(ctx, tenantID, instanceID)`
  method that loads the instance, resolves its pinned workflow definition, and returns
  the transitions whose source state equals the instance's current state.
- **MODIFIED** — `OrchestratorPort` + `OrchestratorService` gain
  `GetAllowedTransitions` (engine-backed; empty list when no engine).
- **MODIFIED** — the `get_current_state` MCP handler includes `allowedTransitions`
  (`event`, `targetStateId`, `priority`) in its response.
- **No** new MCP tools; no change to the tool name/schema beyond the added field.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `mcp/orchestrator-tools`: the `get_current_state` tool SHALL return the allowed
  events/transitions derived from the workflow definition for the current state
  (implementing the existing requirement, PRD 12, 14, 33-34).

## Impact

- **`apps/api/internal/domain/engine/state_machine.go`** — `AllowedTransitions` method.
- **`apps/api/internal/domain/engine/deterministic_test.go`** — test for `AllowedTransitions`.
- **`apps/api/internal/application/services/orchestrator_service.go`** — `GetAllowedTransitions`.
- **`apps/api/internal/interfaces/mcp/server.go`** — `OrchestratorPort` interface.
- **`apps/api/internal/interfaces/mcp/tools.go`** — `handleGetCurrentState` response.
- **`apps/api/internal/interfaces/mcp/server_test.go`**, `e2e_test.go` — mock updates.

## Non-Goals

- Deriving `purpose`/`instructions` for the current state (could be a follow-up; the spec
  lists it, but the primary gap is allowed events/transitions).
- Engine-backed `replay_workflow` re-execution (separate refinement).

## Dependencies

- `mcp-engine-runtime-wiring` (engine wired into the runtime) — already merged.
- `engine-core`, `mcp/orchestrator-tools` spec.

## Notes

- Implements an existing spec requirement; no new external behavior is introduced beyond
  the `allowedTransitions` field on the existing tool response.
