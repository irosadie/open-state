## 1. Context model

> Skill: `api-feature`

- [ ] 1.1 Define `ContextScope` (Tenant/Conversation/Workflow/State/Turn/Memory) and
      `ContextEntry` (scope, key, value, sensitive bool)
- [ ] 1.2 Define `Context` resolver input (tenant id, conversation id, workflow
      instance, current state) and output (available map, missing list)
- [ ] 1.3 Unit test model & scopes

## 2. Resolver

> Skill: `api-feature`

- [ ] 2.1 Implement `Resolve(ctx, state)` returning `available` + `missing` based on
      `requiredContext` of the state (PRD §36)
- [ ] 2.2 Implement hierarchy: tenant → conversation → workflow → state → turn
      (later scopes override earlier on conflict) (PRD §23)
- [ ] 2.3 Implement memory vs workflow-data split (PRD §24): persistent memory kept
      separate from transient workflow data
- [ ] 2.4 Add `sensitive` flag propagation for future PII redaction (PRD §90)
- [ ] 2.5 Unit tests: missing detection, scope precedence, memory/workflow split

## 3. Quality gate

> Skills: `api-code-review`

- [ ] 3.1 `go build ./...`, `go vet ./...`, `go test ./...` pass
- [ ] 3.2 `api-code-review` passes on context package
- [ ] 3.3 No infra/HTTP/LLM dependency in context package
