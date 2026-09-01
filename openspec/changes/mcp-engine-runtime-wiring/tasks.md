## 1. Persistence → engine adapter (Skill: api-feature)

- [x] 1.1 Read `.agents/guides/api-repository.md`, `api-db-repository.md`; review how `PostgresAdapter` composes the six persistence repos.
- [x] 1.2 Create `apps/api/internal/infrastructure/engineadapter/adapter.go` implementing `engine.EngineRepositories` ports (Project, Workflow, Instance, Event) over the pgx repositories.
- [x] 1.3 Implement `WorkflowRepository.GetBySlug`: resolve project UUID, `FindCurrentVersion`, unmarshal `WorkflowVersion.Definition` JSON into `engine.WorkflowDefinition`.
- [x] 1.4 Implement `InstanceRepository` Create/Get/UpdateWithVersion with `engine.WorkflowInstance` ↔ `entities.WorkflowInstance` conversion (including optimistic version).
- [x] 1.5 Implement `EventRepository` Append/IsProcessed/MarkProcessed with `engine.Event` ↔ `entities.Event` conversion.
- [x] 1.6 Implement `ProjectRepository` Get (and a no-op/appropriate Save).
- [x] 1.7 Add `engineadapter/adapter_test.go` unit tests for the conversions (definition unmarshal, instance mapping).

## 2. Wire engine into OrchestratorService (Skill: api-feature)

- [x] 2.1 Add an optional `engine *engine.Engine` field to `OrchestratorService` (new constructor or setter).
- [x] 2.2 Implement engine-backed `ProposeEvent`: run `engine.ProcessEvent`, persist transition; fall back to append-only when engine is nil.
- [x] 2.3 Implement engine-backed `GetCurrentState`: return current state + allowed events/transitions from the loaded definition.
- [x] 2.4 Implement engine-backed `ReplayWorkflow`: replay recorded events through the engine to reproduce state.
- [x] 2.5 Update `orchestrator_service_test.go` to cover engine-backed propose/transition, guard rejection, and replay (in-memory adapter).
- [x] 2.6 Ensure existing non-engine callers/tests still pass when the engine is nil.
- [x] 2.7 Initialize engine-backed `start_workflow` instances at the workflow entry node and persist the `RUNNING` status before MCP reads current state.

## 3. Composition root wiring (Skill: api-feature)

- [x] 3.1 Construct the engine adapter + `engine.NewEngine` in `apps/api/cmd/mcp-server/main.go` and inject into the orchestrator.
- [x] 3.2 Construct the engine adapter + engine in `apps/api/cmd/server/main.go` and inject into the orchestrator.
- [x] 3.3 Update the MCP E2E test (`internal/interfaces/mcp/e2e_test.go`) to assert engine-backed persistence via the repository.

## 4. Quality gate (Skills: api-code-review, api-bugfix)

- [x] 4.1 `go build ./...`, `go vet ./...`, `go test ./...` pass in `apps/api` (including without a DB; Postgres load test skips).
- [x] 4.2 E2E test passes (mock LLM/MCP → engine → persisted transition).
- [x] 4.3 `api-code-review`: adapter conversions correct, no behavior regression for nil-engine path, files end with newline.
- [x] 4.4 Run `go test ./internal/domain/engine/` (deterministic + golden) to confirm no engine regression.
