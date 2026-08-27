## 1. Provider port & value objects (Skill: api-feature)

- [x] 1.1 Read `.agents/settings.json` and `.agents/guides/api-entity.md` before any change
- [x] 1.2 Create `apps/api/internal/domain/capability/provider.go` — `CapabilityProvider` interface with `Invoke(ctx, Invocation) (InvocationResult, error)`
- [x] 1.3 Define `Invocation` value object (capabilityID, tenantID, workflowID, stateID, actionID, payload map, idempotencyKey, policy)
- [x] 1.4 Define `InvocationResult` value object (data, fromMock bool, duration, capabilityEvent *string)
- [x] 1.5 No `interface{}`/`any`; typed Go structs; package doc comment

## 2. Capability resolver (Skill: api-feature)

- [x] 2.1 Read `.agents/guides/api-repository.md` (for `ICapabilityRepository` contract)
- [x] 2.2 Create `apps/api/internal/domain/capability/resolver.go` — `CapabilityResolver` that resolves a logical capability to a provider via `provider_type` + `provider_id`
- [x] 2.3 Implement binding resolution: Global → Tenant → Workflow → State with most-restrictive-wins (DENY > ALLOW; state > workflow > tenant) (PRD §60)
- [x] 2.4 Return classified `CapabilityError` for unknown/denied capability (never invoke provider)
- [x] 2.5 Unit test: allow/deny precedence at each scope

## 3. Execution pipeline (Skill: api-feature)

- [x] 3.1 Create `apps/api/internal/domain/capability/invoker.go` — `CapabilityInvoker` orchestrating the security chain (PRD §62): authenticate → authorize tenant → authorize workflow → authorize state → validate input schema → rate limit → invoke
- [x] 3.2 Inject an input-schema validator (JSON Schema or equivalent) for `input_schema` validation; reject invalid payload with validation error before provider call
- [x] 3.3 Inject a rate limiter port; enforce before invoke
- [x] 3.4 Invoke provider via `CapabilityProvider`, normalize result, attach `fromMock` flag
- [x] 3.5 Unit test each security-chain short-circuit path (no provider call on failure)

## 4. Errors, failures & events (Skill: api-feature)

- [x] 4.1 Create `apps/api/internal/domain/capability/errors.go` — `CapabilityError` with `Kind` (PRD §87: TIMEOUT, UNAUTHORIZED, VALIDATION, EXTERNAL, BUSINESS, …) and `Code` (PRD §63: capability.timeout, capability.unauthorized, capability.validation_failed, capability.unavailable, capability.business_error)
- [x] 4.2 Map provider errors → `CapabilityError` kinds/codes; never leak raw provider errors (PRD §2951)
- [x] 4.3 Emit capability event string on success/failure so transition logic can react
- [x] 4.4 Unit test failure-classification mapping table

## 5. Retry & timeout (Skill: api-feature)

- [x] 5.1 Create `apps/api/internal/domain/capability/retry.go` — exponential backoff + jitter (PRD §88)
- [x] 5.2 Retry only retryable kinds (timeout, unavailable, transient); short-circuit on authorization/validation/business
- [x] 5.3 Enforce timeout via `context.WithTimeout` from policy (PRD §160); respect `max_retry` and `retryable` from state policy
- [x] 5.4 Unit test: retryable vs non-retryable, backoff bounds, budget exhaustion

## 6. Idempotency (Skill: api-feature)

- [x] 6.1 Create `apps/api/internal/domain/capability/idempotency.go` — idempotency key = workflow_instance_id + action_id (PRD §64)
- [x] 6.2 Define an idempotency store **port**; in-memory implementation for tests
- [x] 6.3 On duplicate key, return stored prior result instead of re-invoking provider
- [x] 6.4 Unit test: first-run invokes, duplicate-run returns cached result

## 7. Mock provider (Skill: api-feature)

- [x] 7.1 Create `apps/api/internal/infrastructure/capability/mock_provider.go` implementing `CapabilityProvider` (default sandbox/mock mode, PRD §2064)
- [x] 7.2 Flag results `fromMock = true`
- [x] 7.3 Wire mock provider as default when no real provider bound
- [x] 7.4 Unit test mock provider behavior

## 8. Quality gate (Skills: api-code-review, meta-skill-hygiene)

- [x] 8.1 `go build ./...`, `go vet ./...`, `go test ./...` pass in `apps/api`
- [x] 8.2 Coverage for resolver, invoker, retry, idempotency, failure mapping > 80%
- [x] 8.3 Domain package has no HTTP/DB/MCP-SDK dependency; no circular imports
- [x] 8.4 `api-code-review` passes on the new capability package
- [x] 8.5 All files end with a newline (EOF)
