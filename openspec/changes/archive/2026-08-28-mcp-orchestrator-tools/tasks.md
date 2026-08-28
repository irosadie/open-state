## 1. Domain: RAGProvider port (Skill: api-feature)

- [x] 1.1 Read `.agents/settings.json`, `.agents/guides/api-entity.md`, `.agents/guides/api-repository.md`
- [x] 1.2 Create `internal/domain/rag/provider.go` — `RAGProvider` interface with `Retrieve(ctx, query)` + `Retrieval` result (PRD 171)
- [x] 1.3 Verify: no `any`/`interface{}`; domain port only (PRD 169)

## 2. Application: context compiler (Skill: api-feature)

- [x] 2.1 Read `.agents/guides/api-service.md`, `.agents/guides/api-dto.md`
- [x] 2.2 Create `internal/application/services/context_compiler.go` — `ContextCompiler` composing current state + workflow data + persistent memory + RAG retrievals (PRD 22)
- [x] 2.3 DTO: `CompiledContext` with `available`, `missing`, `memory`, `workflow` sections (PRD 24, 43.2)
- [x] 2.4 Create `internal/domain/rag/redactor.go` — `Redactor` port `Redact(ctx, text) (string, error)` (PRD 90, 169)
- [x] 2.5 Compiler applies the injected redactor last; no PII in output (PRD 90)

## 3. Application: orchestrator use cases (Skill: api-feature)

- [x] 3.1 Create use cases: `start_workflow`, `suspend_workflow`, `resume_workflow`, `cancel_workflow` (PRD 25, 42-43)
- [x] 3.2 Create `propose_event` use case: validate event against current state, execute transition (PRD 38)
- [x] 3.3 Create `get_workflow_instances`, `get_history`, `replay_workflow` use cases (PRD 52, 142)
- [x] 3.4 All use cases tenant-scoped (PRD 4, 96); map repo DomainError → application errors

## 4. MCP tool handlers (Skill: api-feature)

- [x] 4.1 Register `get_current_state`, `get_allowed_capabilities` tool handlers
- [x] 4.2 Register `propose_event`, `start_workflow`, `suspend_workflow`, `resume_workflow`, `cancel_workflow` handlers
- [x] 4.3 Register `get_workflow_instances`, `get_history`, `replay_workflow` handlers
- [x] 4.4 Handlers are thin; resolve tenant from auth context; no business logic (PRD 74)

## 5. Composition root wiring (Skill: api-feature)

- [x] 5.1 Wire `RAGProvider` stub + default `Redactor` at the composition root
- [x] 5.2 Construct `ContextCompiler` and pass it to the MCP server tools that need context
- [x] 5.3 Ensure existing MCP server startup/tools still work (no regression)

## 6. Quality gate (Skills: api-code-review, api-bugfix)

- [x] 6.1 `go build ./...`, `go vet ./...`, `go test ./...` pass in `apps/api`
- [x] 6.2 Unit tests: context compiler (available/missing/memory split), redactor (PII masked), propose_event (valid + invalid event)
- [x] 6.3 Unit tests: lifecycle use cases + history/replay determinism
- [x] 6.4 Smoke: invoke each tool against a seeded instance; verify context output + PII redaction; verify replay reproduces state
- [x] 6.5 `api-code-review`: tenant-scoped everywhere, thin handlers, domain ports only, no LLM-initiated calls (PRD 170), all files end with newline
