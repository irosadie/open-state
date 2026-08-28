## Context

Epic #4 registers 13 MCP tools, but `get_active_workflow`, `resolve_intent`, and
`invoke_capability` are still stubs/partial, and `replay_workflow` is missing. This
slice wires them to the real persistence + capability + intent infrastructure. See
proposal.md for motivation.

## Goals / Non-Goals

**Goals:**
- `get_active_workflow` resolves the active instance from the repository.
- `invoke_capability` is authorized + validated (resolver + schema validator wired).
- `resolve_intent` resolves from the real intent registry + workflow definition.
- Add `replay_workflow` (PRD 52).

**Non-Goals:**
- Wiring `engine.Engine` (EngineRepositories adapter) — large, separate follow-up.
- Concrete RAG backend.
- Frontend.

## Decisions

### D1. Active workflow lookup via instance repository
`OrchestratorService.GetActiveWorkflow(ctx, tenantID, conversationID)` lists instances
for the tenant and returns the active (non-terminal) one whose `CorrelationKey` matches
the conversation. The MCP handler formats id/workflow/status/current-state.

### D2. Capability invoker gets resolver + validator
At the composition root (`cmd/mcp-server`), construct a capability resolver backed by
`ICapabilityRepository` (via the adapter) and pass `capinfra.JSONSchemaValidator{}` into
`capability.NewCapabilityInvoker(...)`. This makes `invoke_capability` enforce
authorization (PRD 59-62) and payload validation (PRD 62).

### D3. Intent resolution via a real resolver
Provide an `IntentResolver` that resolves an intent to its workflow definition +
entry state using the workflow repository (`IWorkflowRepository`) and the intent
registry domain model. The MCP `resolve_intent` handler returns the resolved workflow
slug/version/entry state.

### D4. Replay via event history
`OrchestratorService.ReplayWorkflow(ctx, tenantID, instanceID)` loads the event history
in sequence order and re-applies payloads deterministically to reproduce the resulting
context/state snapshot. Without the engine wired, replay merges event payloads in order
and returns the reproduced context + last event (documented as a state projection).

### D5. Keep handlers thin
Handlers only parse args → call the service → format JSON. All logic lives in the
application layer.

## Schema Outline

No new DB tables. All tools operate on existing entities (workflow instance, state
instance, event, capability).

## Risks / Trade-offs

- [Replay is a projection, not full engine] → Without the engine wired, replay merges
  event payloads deterministically; full guard/transition re-execution is a follow-up
  when the engine is wired.
- [Active-workflow match by correlation] → Reasonable first cut; conversation→instance
  mapping is already the correlation_key convention (PRD 6).

## Migration Plan

1. Branch `feature/epic4-mcp-tool-wiring`.
2. Extend `OrchestratorService` (`GetActiveWorkflow`, `ReplayWorkflow`).
3. Wire capability resolver + schema validator into the invoker at `cmd/mcp-server`.
4. Provide a real `IntentResolver`.
5. Replace stub handlers (`get_active_workflow`, `resolve_intent`); add `replay_workflow`.
6. `go build ./...`, `go vet ./...`, `go test ./...`.
7. Smoke: get_active_workflow on a seeded instance; unauthorized/validation capability;
   resolve_intent; replay reproduces state.
8. PR → review → merge.

**Rollback**: additive; removing the wiring restores stubs. No data migration.

## Open Questions

None — behavior is fixed by the existing MCP specs and PRD refs.
