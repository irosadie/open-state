## Why

Epic **#4 (MCP & Integrasi)** already registers 13 MCP tools (mcp-server-runtime +
mcp-orchestrator-tools), but several are still **stubs or partially wired**: 
`get_active_workflow` always returns "no active workflow", `invoke_capability` runs
with a nil resolver/validator (so authorization + schema validation are not enforced),
and `resolve_intent` only lists a dummy registry. `replay_workflow` (PRD 52) is
declared in the epic checklist but has no handler. This slice wires those tools to the
real persistence + capability + intent infrastructure so the 3rd-party LLM/RAG client
gets truthful, enforceable results — completing the MCP contract (PRD 170).

## What Changes

- **`get_active_workflow`** — resolve the active workflow instance for a conversation
  from the instance repository (PRD 10, 142) instead of returning a stub.
- **`invoke_capability`** — wire the capability resolver (via `ICapabilityRepository`)
  and the schema validator (`JSONSchemaValidator`) into the invoker so execution is
  authorized (PRD 59-62) and payloads are validated (PRD 62) before any provider call.
- **`resolve_intent`** — resolve an intent from the real intent registry + workflow
  definition lookup (PRD 38, 171) instead of a dummy list.
- **`replay_workflow`** — new use case + MCP tool that replays the recorded event
  history of an instance in sequence order to reproduce its resulting state (PRD 52).

## Capabilities

### New Capabilities

- `mcp/tool-wiring`: completes the wiring of active-workflow, capability-invocation,
  intent-resolution, and workflow-replay so the MCP tools return real, enforceable
  results.

### Modified Capabilities

- `mcp/server-runtime`: `get_active_workflow`, `resolve_intent`, `invoke_capability`
  become fully wired (no longer stubs).
- `mcp/orchestrator-tools`: adds the `replay_workflow` tool + use case.
- `mcp/capability-execution`: capability resolver + schema validator wired into the
  invoker at the composition root.

## Impact

- **`apps/api/internal/application/services/`** — extend `OrchestratorService` with
  `ReplayWorkflow` and `ResolveActiveWorkflow`; wire capability resolver/validator.
- **`apps/api/internal/interfaces/mcp/`** — replace the stub handlers for
  `get_active_workflow` and `resolve_intent`; add the `replay_workflow` tool.
- **`apps/api/internal/infrastructure/capability/`** — ensure the capability resolver
  backed by `ICapabilityRepository` is available.
- **`apps/api/cmd/mcp-server/main.go`** — construct the capability resolver +
  schema validator and pass them into the invoker; pass the intent resolver.
- **No** web, worker, or shared-package changes.

## Non-Goals

- Wiring the full `engine.Engine` (EngineRepositories adapter) — large and separate;
  `propose_event` transition validation remains in a follow-up.
- A concrete RAG backend (already covered by `mcp/rag-provider` port).
- Frontend/UI.

## Dependencies

- `mcp-server-runtime`, `mcp-orchestrator-tools`, `mcp-capability-execution`.
- `persistence-*` (workflow/instance/event/capability repositories).
- Epic #4.

## Notes

- PRD 170 fixes the boundary: tools are invoked by a 3rd-party LLM client; the
  platform never calls an LLM internally. All work here stays on the tool side.
