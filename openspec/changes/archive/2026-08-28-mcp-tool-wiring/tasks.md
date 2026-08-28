## 1. Orchestrator service extensions (Skill: api-feature)

- [x] 1.1 Read `.agents/settings.json`, `.agents/guides/api-service.md`
- [x] 1.2 Add `GetActiveWorkflow(ctx, tenantID, conversationID)` — resolve active instance via `IInstanceRepository` (PRD 10, 142)
- [x] 1.3 Add `ReplayWorkflow(ctx, tenantID, instanceID)` — replay event history in sequence order to reproduce state (PRD 52)
- [x] 1.4 Both tenant-scoped (PRD 4, 96); map repo DomainError → application errors

## 2. Capability invoker wiring (Skill: api-feature)

- [x] 2.1 Construct a capability resolver backed by `ICapabilityRepository` (via adapter)
- [x] 2.2 Pass the resolver + `JSONSchemaValidator` into `capability.NewCapabilityInvoker` at `cmd/mcp-server`
- [x] 2.3 Verify `invoke_capability` enforces authorization (PRD 59-62) and payload validation (PRD 62)

## 3. Intent resolver (Skill: api-feature)

- [x] 3.1 Provide a real `IntentResolver` resolving an intent to its workflow definition + entry state via `IWorkflowRepository` (PRD 38, 171)
- [x] 3.2 Wire it into the MCP server `Dependencies.IntentResolver`
- [x] 3.3 `resolve_intent` returns workflow slug/version/entry state

## 4. MCP handlers (Skill: api-feature)

- [x] 4.1 Replace the stub `get_active_workflow` handler to call `GetActiveWorkflow`
- [x] 4.2 Replace the stub `resolve_intent` to use the real resolver
- [x] 4.3 Add the `replay_workflow` tool handler delegating to `ReplayWorkflow`
- [x] 4.4 Handlers stay thin (parse → call service → format); no business logic (PRD 74)

## 5. Quality gate (Skills: api-code-review, api-bugfix)

- [x] 5.1 `go build ./...`, `go vet ./...`, `go test ./...` pass in `apps/api`
- [x] 5.2 Unit tests: `GetActiveWorkflow` (active + none), `ReplayWorkflow` (deterministic order)
- [x] 5.3 Unit tests: capability invocation authorization + schema validation
- [x] 5.4 Smoke: get_active_workflow on seeded instance; unauthorized/validation capability; resolve_intent; replay reproduces state
- [x] 5.5 `api-code-review`: tenant-scoped everywhere, thin handlers, no regressions, all files end with newline
