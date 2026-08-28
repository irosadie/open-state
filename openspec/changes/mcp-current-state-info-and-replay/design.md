## Context

Two MCP tools partially implement their specs: `get_current_state` omits the current
node's purpose/instructions, and `replay_workflow` merges payloads instead of re-driving
events through the engine. The runtime engine is wired into the MCP path, so the engine
can derive current-state info and replay events deterministically. See proposal.md for
motivation and the spec for required behavior.

## Goals / Non-Goals

**Goals:**
- Return current-state `purpose`/`instructions`/`requiredContext` via the engine.
- Replay recorded events through a fresh in-memory engine instance to reproduce state
  without persisting (PRD 52).

**Non-Goals:**
- A concrete RAG backend.
- Changing `propose_event` semantics.

## Decisions

### D1 — Engine `CurrentStateInfo`
Add `Engine.CurrentStateInfo(ctx, tenantID, instanceID)` that loads the instance +
definition and returns the current node's `Description` (as purpose), `Instructions`,
and `RequiredContext`. Reuses the existing `Instances.Get` (adapter resolves
project/slug/state key) + `loadDefinition` + `nodeByID`.

### D2 — Engine `Replay` with an in-memory engine
Replay must reproduce state without persisting to PostgreSQL. Add
`Engine.Replay(ctx, tenantID, instanceID, events)` that:
1. Reads the original instance (project/slug/conversation) via `Instances.Get`.
2. Loads the definition via `loadDefinition`.
3. Builds a **fresh in-memory** engine (using in-memory repos for Instance + Event +
   Workflow + Project) so replay side effects are isolated.
4. Starts a fresh instance (`StartWorkflow`) and re-drives each recorded event via
   `ProcessEvent`.
5. Returns the reproduced current state + merged context.

This keeps replay deterministic and non-destructive (PRD 52, 170). The in-memory repos
are test-double-style structs moved into a non-test file so production replay can use
them.

### D3 — Orchestrator + handler wiring
`OrchestratorService` exposes `CurrentStateInfo` (engine-backed) and updates
`ReplayWorkflow` to call `engine.Replay` when wired (falling back to the merge approach
when the engine is nil). The `get_current_state` handler appends `purpose`,
`instructions`, `requiredContext`; the `replay_workflow` handler returns the
engine-reproduced state.

## Risks / Trade-offs

- **Replay side effects** → mitigated by using a fresh in-memory engine; nothing is
  written to PostgreSQL.
- **Engine nil callers** → `CurrentStateInfo` returns empty fields and `ReplayWorkflow`
  falls back to merge, so non-engine callers keep working.

## Migration Plan

1. Add engine in-memory replay repos (non-test) + `Engine.Replay` + `CurrentStateInfo`.
2. Add engine tests for both.
3. Extend `OrchestratorPort`/`OrchestratorService`.
4. Update handlers + mocks.
5. `go build ./...`, `go vet ./...`, `go test ./...`.
