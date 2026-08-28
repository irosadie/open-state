## 1. Engine (Skill: api-feature)

- [x] 1.1 Add `Engine.AllowedTransitions(ctx, tenantID, instanceID)` in `state_machine.go` deriving transitions from the current state of the pinned definition.
- [x] 1.2 Add `TestAllowedTransitions` in `deterministic_test.go` (current-state transitions returned; terminal/no-transition returns empty).

## 2. Orchestrator + MCP (Skill: api-feature)

- [x] 2.1 Add `GetAllowedTransitions` to `OrchestratorPort` (`server.go`) and implement in `OrchestratorService` (engine-backed; empty when nil).
- [x] 2.2 Update `handleGetCurrentState` in `tools.go` to append `allowedTransitions` (`event`, `targetStateId`, `priority`).
- [x] 2.3 Update mocks in `server_test.go` and `e2e_test.go` to implement `GetAllowedTransitions`.

## 3. Quality gate (Skills: api-code-review, api-bugfix)

- [x] 3.1 `go build ./...`, `go vet ./...`, `go test ./...` pass in `apps/api`.
- [x] 3.2 MCP E2E test passes (get_current_state includes allowed transitions).
- [x] 3.3 Files end with a newline; no regression to existing MCP tools.
