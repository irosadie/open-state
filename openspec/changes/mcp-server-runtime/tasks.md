## 1. MCP server module setup (Skill: api-feature)

- [ ] 1.1 Read `.agents/settings.json`, `go.work`, and `.agents/guides/api-service.md`
- [ ] 1.2 Create `apps/mcp-server/` Go module and add it to `go.work` (`use ./apps/mcp-server`)
- [ ] 1.3 Add MCP Go SDK dependency (Streamable HTTP transport support)
- [ ] 1.4 `go build ./...` passes from repo root (go.work resolves module)

## 2. MCP server entrypoint (Skill: api-feature)

- [ ] 2.1 Create `apps/mcp-server/cmd/server/main.go` — starts the MCP server over Streamable HTTP on a configurable port
- [ ] 2.2 Register tool handlers for `resolve_intent`, `get_active_workflow`, `get_context`, `invoke_capability`
- [ ] 2.3 Wire handlers to existing services: intent resolver, active-workflow lookup, context resolver, capability invoker (PRD §40.1, §22, §153)
- [ ] 2.4 Add `/health/live` and `/health/ready` (PRD §178)

## 3. Capability invocation tool + filtering (Skill: api-feature)

- [ ] 3.1 Implement `invoke_capability` handler delegating to `capability.Invoker` (PRD §153)
- [ ] 3.2 Filter advertised capabilities by binding (tenant/workflow/state/policy); never expose full registry (PRD §106, §3309)
- [ ] 3.3 Reject unauthorized capability requests with an authorization result (no invocation)
- [ ] 3.4 Return normalized result/event; never raw provider errors (PRD §2951)

## 4. Context & intent tools (Skill: api-feature)

- [ ] 4.1 Implement `resolve_intent` → mapped workflow + current state (PRD §40.1, §1684)
- [ ] 4.2 Implement `get_active_workflow` → active workflow, current state, allowed events
- [ ] 4.3 Implement `get_context` → compiled context (tenant, workflow, state, purpose, instructions, available/missing context, allowed events/transitions, available capabilities, memory) (PRD §22)

## 5. MCP client adapter (Skill: api-feature)

- [ ] 5.1 Create `apps/api/internal/infrastructure/capability/mcp_provider.go` implementing `CapabilityProvider` via MCP SDK client (Streamable HTTP)
- [ ] 5.2 Map `Capability.provider_id` → MCP endpoint + tool; normalize result into `InvocationResult`
- [ ] 5.3 Map MCP failures/timeouts → classified `CapabilityError` (PRD §87)
- [ ] 5.4 Do NOT import MCP SDK in `domain/capability`; SDK stays in infrastructure only (PRD §2559)
- [ ] 5.5 Unit test the adapter against a mock/embedded MCP server

## 6. Credential resolution (Skill: api-feature)

- [ ] 6.1 Create `apps/api/internal/infrastructure/capability/secrets.go` — resolves `credential_reference` from env / secret manager (PRD §61)
- [ ] 6.2 Define a `credential_resolver` port; in-memory + env implementations for tests
- [ ] 6.3 Never log credentials/tokens/authorization headers (PRD §91); fail closed when store unavailable
- [ ] 6.4 Unit test: resolution, redaction, fail-closed behavior

## 7. Docker readiness (Skill: ops-docker)

- [ ] 7.1 Read `.agents/guides/ARCHITECTURE.md` and existing Dockerfile conventions
- [ ] 7.2 Add a multi-stage Dockerfile for `apps/mcp-server` (Linux deploy, PRD §177)
- [ ] 7.3 Ensure build runs without secrets baked in; image is stateless
- [ ] 7.4 Smoke-test the image boots and exposes health endpoints

## 8. Quality gate (Skills: api-code-review, meta-skill-hygiene)

- [ ] 8.1 `go build ./...`, `go vet ./...`, `go test ./...` pass at repo root
- [ ] 8.2 MCP server boots and declares all four tools over Streamable HTTP
- [ ] 8.3 `api-code-review` passes on `apps/mcp-server` and the new adapter
- [ ] 8.4 Domain/engine have no MCP SDK dependency; no circular imports
- [ ] 8.5 All files end with a newline (EOF)
