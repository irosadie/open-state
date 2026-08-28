## Context

Epic #4's "keluar saat selesai" (LLM can drive a full conversation flow over MCP) is
currently unmet because the domain engine is not wired into the production path. The MCP
tools and `OrchestratorService` are wired to the persistence layer, but
`OrchestratorService.ProposeEvent` only appends the event; it does not run the engine's
`event → guard → transition`. The engine (`engine.NewEngine`) is only exercised in tests
and the E2E mock.

The engine consumes its own ports (`engine.EngineRepositories`) with engine-domain types
(`engine.WorkflowDefinition`, `engine.WorkflowInstance`, `engine.Event`), while the
persistence layer exposes `entities.*` types through `I*Repository` interfaces. Wiring
the engine into production therefore requires an **adapter** at the boundary.

See proposal.md for motivation and the specs for required behavior.

## Goals / Non-Goals

**Goals:**
- Wire the engine into `OrchestratorService` so `propose_event`, `get_current_state`, and
  `replay_workflow` execute real state-machine transitions against PostgreSQL.
- Provide a persistence→engine adapter satisfying `engine.EngineRepositories`.
- Keep the MCP tool surface and HTTP contract unchanged.

**Non-Goals:**
- New MCP tools / schema changes.
- A concrete RAG backend (RAG/LLM remain 3rd-party, PRD 170).
- New engine behavior (retry/suspend beyond what the engine already implements).
- Rate limiting / observability (epic #6).

## Decisions

### D1 — Adapter lives in `infrastructure`, produces engine ports

Create `apps/api/internal/infrastructure/engineadapter/` with an `Adapter` that
implements the four engine ports by composing the existing pgx repositories. This keeps
the engine domain-pure (ADR-001) and mirrors how the `PostgresAdapter` composes the six
persistence interfaces. The adapter is the only place that maps `entities.*` ↔
`engine.*`.

- `ProjectRepository`: `Get`/`Save` → `IProjectRepository.FindByID`/`Create` (or a
  no-op `Save` since projects are seeded; only `Get` is used by the engine today).
- `WorkflowRepository`: `GetBySlug` → resolve the project UUID, `FindCurrentVersion`,
  unmarshal `WorkflowVersion.Definition` JSON into `engine.WorkflowDefinition`.
  `Save` → not used by the engine runtime; return nil or persist via the workflow repo
  if needed.
- `InstanceRepository`: `Create`/`Get`/`UpdateWithVersion` → `IInstanceRepository`
  equivalents, converting `engine.WorkflowInstance` ↔ `entities.WorkflowInstance`.
- `EventRepository`: `Append`/`IsProcessed`/`MarkProcessed` → `IEventRepository`
  equivalents, converting `engine.Event` ↔ `entities.Event`.

- **Alternative considered:** extending `OrchestratorService` to call the engine's
  domain functions directly. Rejected because the engine's public surface is the
  `Engine` struct over its own ports; the adapter is the correct seam and keeps the
  engine replaceable/testable.

### D2 — OrchestratorService holds an optional engine

Add an `engine *engine.Engine` field to `OrchestratorService` (via a new constructor or a
setter). When present, `ProposeEvent` runs `engine.ProcessEvent`; `GetCurrentState`
returns engine-computed state + allowed transitions; `ReplayWorkflow` replays through the
engine. When absent (nil), the service degrades to the current append/merge behavior,
so existing unit tests and non-engine callers keep working.

- **Alternative:** a hard engine dependency in `OrchestratorService`. Rejected to avoid
  breaking existing tests and to allow a phased rollout.

### D3 — Allowed transitions from current state

`GetCurrentState` derives allowed events/transitions from the loaded workflow definition
via the engine's helpers (transitions whose `SourceStateID` equals the current state),
matching `get_current_state` spec (purpose, instructions, allowed events/transitions).

### D4 — Composition root wires the adapter

In `apps/api/cmd/mcp-server/main.go` (and `apps/api/cmd/server/main.go`), construct the
adapter over `adapter.Projects()/Workflows()/Instances()/Events()`, build
`engine.NewEngine(adapterRepos)`, and inject it into the orchestrator. This makes the
production MCP path engine-backed.

## Risks / Trade-offs

- **Adapter type-conversion bugs** → mitigated by unit tests on the adapter using the
  seeded/persisted definitions (the E2E mock already exercises the engine path; we add
  an engine-backed service test).
- **Engine error semantics vs HTTP** → the engine returns `domain.DomainError`
  (Conflict/Validation); the MCP handler maps these to structured tool errors, which is
  already the behavior for orchestrator tools.
- **Behavior change to propose_event** → previously append-only; now engine-transitioning.
  This is the intended fix (spec `orchestrator-tools` already requires it), but it changes
  observable behavior for existing callers. Mitigated by clear rejection semantics and the
  existing E2E test updated to the engine-backed path.

## Migration Plan

1. Add `infrastructure/engineadapter` with the four port implementations + conversions.
2. Add unit tests for the adapter (mapping + workflow-definition unmarshal).
3. Add the engine field to `OrchestratorService`; implement engine-backed
   `ProposeEvent`/`GetCurrentState`/`ReplayWorkflow`; update service tests.
4. Wire the composition root (`mcp-server/main.go`, `server/main.go`).
5. Update the MCP E2E test to assert engine-backed persistence.
6. `go build ./...`, `go vet ./...`, `go test ./...` in `apps/api`.

## Open Questions

- Whether `start_workflow` (MCP) should begin via `engine.StartWorkflow` vs the existing
  `IInstanceRepository.Create`. Default: use the engine for consistency once wired;
  revisit if it conflicts with lifecycle semantics.
