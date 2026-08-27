# engine-core Specification

## Purpose

Define the deterministic runtime engine core: entities, guard evaluation,
state machine execution, intent resolution, and repository ports — domain-pure
with no HTTP/DB/LLM dependency.

## ADDED Requirements

### Requirement: Domain entities

The engine SHALL define domain entities for workflow definitions and runtime
instances.

#### Scenario: Workflow definition

- GIVEN a workflow is loaded
- THEN it has nodes (with kind START/STATE/DECISION/EVENT/END), transitions,
  guards, policies, and triggers
- AND a status (DRAFT/VALIDATING/VALID/PUBLISHED/ARCHIVED)

#### Scenario: Runtime instance

- GIVEN an instance is created
- THEN it references a workflow version
- AND tracks a current state id and a lifecycle status

### Requirement: Guard evaluation

The engine SHALL evaluate guards deterministically from data (no arbitrary code).

#### Scenario: Operator evaluation

- GIVEN a guard condition (field, operator, value) and a context map
- THEN the engine evaluates `==`, `!=`, `>`, `>=`, `<`, `<=`, `IN`, `EXISTS`
- AND returns true/false deterministically

#### Scenario: AND/OR grouping

- GIVEN guard groups with AND/OR logic
- THEN the engine combines conditions accordingly
- AND all conditions in an AND group must pass

### Requirement: State machine execution

The engine SHALL process events through a deterministic pipeline.

#### Scenario: Process event

- GIVEN an instance in a state and an incoming event
- THEN the engine validates the event is allowed from that state
- AND evaluates guards of candidate transitions
- AND applies the highest-priority passing transition
- AND updates the instance current state + status

#### Scenario: Guard failure

- GIVEN an event whose candidate transition guard fails
- THEN the engine does NOT transition
- AND returns a structured rejection (guard failed)

### Requirement: Lifecycle

The engine SHALL enforce workflow and state lifecycles.

#### Scenario: Workflow lifecycle

- GIVEN a workflow starts
- THEN status is CREATED → RUNNING → (WAITING) → terminal
- AND terminal states are COMPLETED/CANCELLED/FAILED/EXPIRED

#### Scenario: State lifecycle

- GIVEN a state is entered
- THEN status is ENTERING → ACTIVE → (WAITING) → EXITING → COMPLETED
- AND timeout may transition to another state

### Requirement: Intent resolution

The engine SHALL resolve conversation intent to a workflow and initial state.

#### Scenario: Resolve intent

- GIVEN an intent id
- THEN the engine returns the mapped workflow, entry event, and initial state

### Requirement: Repository ports

The engine SHALL depend only on repository interfaces, not a concrete DB.

#### Scenario: DB-agnostic

- GIVEN the engine is constructed with repository ports
- THEN it works with any implementation (Postgres, in-memory fake)
- AND is unit-testable without a database

### Requirement: Deterministic tests

The engine SHALL be fully unit-testable without an LLM.

#### Scenario: No LLM dependency

- GIVEN the engine core is imported
- THEN it has no LLM/HTTP/DB dependency
- AND unit tests cover guard eval & state machine (>80%)
