## 1. Domain entities (engine/domain-model)

> Skill: `api-feature`

- [ ] 1.1 Define `WorkflowNodeKind` & `WorkflowNode` (id, kind, name, description,
      requiredContext, capabilities, instructions, policy, isTerminal) in
      `apps/api/internal/domain/engine/model.go`
- [ ] 1.2 Define `TransitionDefinition` (sourceStateId, event, targetStateId,
      guards, priority) and `GuardGroup` / `GuardCondition`
- [ ] 1.3 Define `WorkflowDefinition` (slug, name, schemaVersion, status, nodes,
      transitions, policy, triggers) and `WorkflowStatus` enum
- [ ] 1.4 Define `WorkflowInstance` (id, workflowId, versionId, status, currentStateId,
      context) + `WorkflowInstanceStatus` enum (CREATED/RUNNING/WAITING/COMPLETED/
      CANCELLED/FAILED/EXPIRED)
- [ ] 1.5 Define `StateInstance` (id, workflowInstanceId, stateId, status, enteredAt,
      expiresAt) + `StateInstanceStatus` enum
- [ ] 1.6 Define `Event` (id, type, source, payload, timestamp, idempotencyKey) +
      `EventSource` enum
- [ ] 1.7 Define `Policy`, `Capability` value objects
- [ ] 1.8 Unit test entities & enums

## 2. Guard evaluator (engine/guard-eval)

> Skill: `api-feature`

- [ ] 2.1 Implement `Evaluate(guards, context) (bool, error)` — supports AND/OR groups
- [ ] 2.2 Implement operators: `==`, `!=`, `>`, `>=`, `<`, `<=`, `IN`, `EXISTS`,
      `NOT_EXISTS` (PRD §35)
- [ ] 2.3 Context lookup helper (dot-path, e.g. `payment.status`)
- [ ] 2.4 No arbitrary code execution (data-driven only)
- [ ] 2.5 Unit tests for every operator + AND/OR logic

## 3. State machine executor (engine/state-machine)

> Skill: `api-feature`

- [ ] 3.1 Implement `NewEngine(ports)` constructor
- [ ] 3.2 Implement `StartWorkflow(def, intentEvent)` → creates instance, enters START
- [ ] 3.3 Implement `ProcessEvent(instanceId, event)` pipeline:
      load → validate event allowed → eval guards → pick transition (priority,
      lower first) → apply → emit result (PRD §152, §33, §34)
- [ ] 3.4 Apply state lifecycle (ENTERING→ACTIVE→WAITING→EXITING→COMPLETED)
      with timeout detection
- [ ] 3.5 Apply workflow lifecycle transitions & terminal states (PRD §10)
- [ ] 3.6 Update instance `currentStateId` & version snapshot
- [ ] 3.7 Unit tests: happy path, guard-fail, priority, terminal, timeout

## 4. Intent resolution (engine/intent-resolver)

> Skill: `api-feature`

- [ ] 4.1 Define `IntentDefinition` + `IntentRegistry` (id, name, description,
      workflowSlug, entryEvent, examples, priority) (PRD §40.1)
- [ ] 4.2 Implement `Resolve(intentId)` → workflow + entryEvent + initial state
- [ ] 4.3 Seed intent registry (BOOKING_PADEL, ORDER_MAKANAN, ORDER_DOKTER)
- [ ] 4.4 Unit test: resolve each intent → correct workflow + initial state

## 5. Repository ports (engine/repository-ports)

> Skill: `api-feature` (define interfaces only; Postgres impl is Epic #3)

- [ ] 5.1 Define `WorkflowRepository` interface (save/get definition, list)
- [ ] 5.2 Define `InstanceRepository` interface (create/get/update instance,
      optimistic version)
- [ ] 5.3 Define `EventRepository` interface (append/get events, idempotency)
- [ ] 5.4 Unit test with in-memory fake implementing ports (proves engine is DB-agnostic)

## 6. Quality gate

> Skills: `api-code-review`, `meta-skill-hygiene`

- [ ] 6.1 `go build ./...`, `go vet ./...`, `go test ./...` pass
- [ ] 6.2 `api-code-review` passes on new engine package
- [ ] 6.3 No circular imports; domain has no infra/HTTP/LLM dependency
- [ ] 6.4 Coverage for guard evaluator & state machine > 80%
