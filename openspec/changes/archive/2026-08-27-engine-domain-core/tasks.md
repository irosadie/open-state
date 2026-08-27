## 1. Domain entities (engine/domain-model)

> Skill: `api-feature`

- [x] 1.1 Define `WorkflowNodeKind` & `WorkflowNode` (id, kind, name, description,
      requiredContext, capabilities, instructions, policy, isTerminal) in
      `apps/api/internal/domain/engine/model.go`
- [x] 1.2 Define `TransitionDefinition` (sourceStateId, event, targetStateId,
      guards, priority) and `GuardGroup` / `GuardCondition`
- [x] 1.3 Define `WorkflowDefinition` (slug, name, schemaVersion, status, nodes,
      transitions, policy, triggers) and `WorkflowStatus` enum
- [x] 1.4 Define `WorkflowInstance` (id, workflowId, versionId, status, currentStateId,
      context) + `WorkflowInstanceStatus` enum (CREATED/RUNNING/WAITING/COMPLETED/
      CANCELLED/FAILED/EXPIRED)
- [x] 1.5 Define `StateInstance` (id, workflowInstanceId, stateId, status, enteredAt,
      expiresAt) + `StateInstanceStatus` enum
- [x] 1.6 Define `Event` (id, type, source, payload, timestamp, idempotencyKey) +
      `EventSource` enum
- [x] 1.7 Define `Policy`, `Capability` value objects
- [x] 1.8 Unit test entities & enums

## 2. Guard evaluator (engine/guard-eval)

> Skill: `api-feature`

- [x] 2.1 Implement `Evaluate(guards, context) (bool, error)` — supports AND/OR groups
- [x] 2.2 Implement operators: `==`, `!=`, `>`, `>=`, `<`, `<=`, `IN`, `EXISTS`,
      `NOT_EXISTS` (PRD §35)
- [x] 2.3 Context lookup helper (dot-path, e.g. `payment.status`)
- [x] 2.4 No arbitrary code execution (data-driven only)
- [x] 2.5 Unit tests for every operator + AND/OR logic

## 3. State machine executor (engine/state-machine)

> Skill: `api-feature`

- [x] 3.1 Implement `NewEngine(ports)` constructor
- [x] 3.2 Implement `StartWorkflow(def, intentEvent)` → creates instance, enters START
- [x] 3.3 Implement `ProcessEvent(instanceId, event)` pipeline:
      load → validate event allowed → eval guards → pick transition (priority,
      lower first) → apply → emit result (PRD §152, §33, §34)
- [x] 3.4 Apply state lifecycle (ENTERING→ACTIVE→WAITING→EXITING→COMPLETED)
      with timeout detection
- [x] 3.5 Apply workflow lifecycle transitions & terminal states (PRD §10)
- [x] 3.6 Update instance `currentStateId` & version snapshot
- [x] 3.7 Unit tests: happy path, guard-fail, priority, terminal, timeout

## 4. Intent resolution (engine/intent-resolver)

> Skill: `api-feature`

- [x] 4.1 Define `IntentDefinition` + `IntentRegistry` (id, name, description,
      workflowSlug, entryEvent, examples, priority) (PRD §40.1)
- [x] 4.2 Implement `Resolve(intentId)` → workflow + entryEvent + initial state
- [x] 4.3 Seed intent registry (BOOKING_PADEL, ORDER_MAKANAN, ORDER_DOKTER)
- [x] 4.4 Unit test: resolve each intent → correct workflow + initial state

## 5. Repository ports (engine/repository-ports)

> Skill: `api-feature` (define interfaces only; Postgres impl is Epic #3)

- [x] 5.1 Define `WorkflowRepository` interface (save/get definition, list)
- [x] 5.2 Define `InstanceRepository` interface (create/get/update instance,
      optimistic version)
- [x] 5.3 Define `EventRepository` interface (append/get events, idempotency)
- [x] 5.4 Unit test with in-memory fake implementing ports (proves engine is DB-agnostic)

## 6. Quality gate

> Skills: `api-code-review`, `meta-skill-hygiene`

- [x] 6.1 `go build ./...`, `go vet ./...`, `go test ./...` pass
- [x] 6.2 `api-code-review` passes on new engine package
- [x] 6.3 No circular imports; domain has no infra/HTTP/LLM dependency
- [x] 6.4 Coverage for guard evaluator & state machine > 80%
