## 1. Engine (Skill: api-feature)

- [x] 1.1 Add `Engine.CurrentStateInfo(ctx, tenantID, instanceID)` returning the current node's purpose/instructions/requiredContext from the pinned definition.
- [x] 1.2 Add in-memory replay repos (non-test) + `Engine.Replay(ctx, tenantID, instanceID, events)` re-driving events through a fresh in-memory engine to reproduce state/context (non-persisting).
- [x] 1.3 Add tests for `CurrentStateInfo` and `Replay`.

## 2. Orchestrator + MCP (Skill: api-feature)

- [x] 2.1 Add `CurrentStateInfo` to `OrchestratorPort` and `OrchestratorService` (engine-backed; empty when nil).
- [x] 2.2 Update `ReplayWorkflow` to call `engine.Replay` when wired (fallback to merge when nil).
- [x] 2.3 Update `handleGetCurrentState` to append `purpose`, `instructions`, `requiredContext`.
- [x] 2.4 Update `handleReplayWorkflow` to return engine-reproduced state.
- [x] 2.5 Update mocks (`server_test.go`, `e2e_test.go`) to implement `CurrentStateInfo`.

## 3. Quality gate (Skills: api-code-review, api-bugfix)

- [x] 3.1 `go build ./...`, `go vet ./...`, `go test ./...` pass in `apps/api`.
- [x] 3.2 MCP E2E + engine tests pass (get_current_state info + replay).
- [x] 3.3 Files end with a newline; no regression to existing MCP tools.
