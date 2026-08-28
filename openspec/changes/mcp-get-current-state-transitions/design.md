## Context

The `get_current_state` MCP tool returns instance + state, but not the allowed
events/transitions the spec requires. The runtime engine is now wired into the MCP path
(`mcp-engine-runtime-wiring`), so the engine can derive allowed transitions from the
instance's pinned workflow definition. See proposal.md for motivation and the spec for
required behavior.

## Goals / Non-Goals

**Goals:**
- Derive and return allowed events/transitions from the current state via the engine.
- Add `GetAllowedTransitions` to the `OrchestratorPort` and handler.

**Non-Goals:**
- Deriving `purpose`/`instructions` (follow-up).
- Engine-backed `replay_workflow` re-execution.

## Decisions

### D1 — Engine exposes `AllowedTransitions`
Add `Engine.AllowedTransitions(ctx, tenantID, instanceID)` that reuses the existing
`Instances.Get` (which the adapter uses to resolve project/slug/state key) +
`loadDefinition` + a filter on transitions whose `SourceStateID` equals the current
state. This keeps the engine as the single source of transition truth (PRD 34).

### D2 — OrchestratorPort + service delegate to the engine
Add `GetAllowedTransitions` to `OrchestratorPort` and `OrchestratorService`; when the
engine is nil it returns an empty list (degraded), matching the optional-engine pattern
from `mcp-engine-runtime-wiring`.

### D3 — Handler appends `allowedTransitions`
`handleGetCurrentState` maps each engine `TransitionDefinition` to `{event,
targetStateId, priority}` and includes it in the response under `allowedTransitions`.

## Risks / Trade-offs

- **All transitions vs only passing-guard transitions** → The tool returns all
  transitions from the current state (candidates), not only those whose guards currently
  pass; guard evaluation happens at `propose_event` time. This matches the spec intent
  ("allowed events/transitions") and avoids over-evaluating context in a read tool.
- **Engine nil in some callers** → `GetAllowedTransitions` returns an empty list rather
  than erroring, so the tool still works for non-engine callers.

## Migration Plan

1. Add `Engine.AllowedTransitions` + test.
2. Add `GetAllowedTransitions` to `OrchestratorPort` + `OrchestratorService`.
3. Update `handleGetCurrentState`; update mocks (`server_test.go`, `e2e_test.go`).
4. `go build ./...`, `go vet ./...`, `go test ./...`.
