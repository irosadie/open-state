## Purpose

Wire the domain state engine into the application/MCP runtime path so orchestrator
tools (`propose_event`, `get_current_state`, `replay_workflow`) execute deterministic
state-machine transitions against the persistence adapter, completing the full
conversation flow for epic #4 (PRD 170).

## ADDED Requirements

### Requirement: Engine-backed propose_event

The system SHALL run the domain engine's `event → guard → transition` evaluation when
a client proposes an event, and persist the resulting state.

- The engine SHALL load the workflow definition for the instance's pinned version.
- The engine SHALL validate the event is allowed from the current state.
- The engine SHALL evaluate guards and pick the highest-priority passing transition
  (PRD 33-34).
- The engine SHALL persist the resulting workflow-instance state (via the instance
  repository) and append the event (via the event repository).
- If no transition passes or the event is disallowed, the engine SHALL return a
  structured rejection and SHALL NOT change the instance state.

#### Scenario: Propose a valid event transitions state

- **WHEN** a client proposes an event allowed from the current state with passing guards
- **THEN** the engine executes the highest-priority transition and the instance's
  current state changes
- **AND** the new state is persisted and readable via the repository

#### Scenario: Propose a disallowed event is rejected

- **WHEN** a client proposes an event not allowed from the current state
- **THEN** the engine returns a validation/conflict error and the instance state is
  unchanged

#### Scenario: Guard failure rejects the proposal

- **WHEN** an allowed event has no transition whose guards pass
- **THEN** the engine returns a guard-failed rejection and the instance state is
  unchanged

### Requirement: Engine-backed current state

The system SHALL return the engine-computed current state of a workflow instance,
including its allowed events/transitions.

#### Scenario: Read engine-computed state

- **WHEN** a client requests the current state of an instance
- **THEN** the system returns the current state id, the node's purpose/instructions, and
  the allowed events/transitions derived from the workflow definition
- **AND** reflects the engine's persisted current-state projection

### Requirement: Engine-backed replay

The system SHALL replay a workflow instance's recorded events through the engine to
reproduce its resulting state.

#### Scenario: Replay reproduces state

- **WHEN** a client invokes replay for an instance
- **THEN** the system replays the event history in deterministic sequence order through
  the engine
- **AND** returns the reproduced resulting context/state

### Requirement: Persistence adapter for the engine

The system SHALL provide an adapter that satisfies the engine's repository ports using
the existing PostgreSQL repositories (ADR-001).

- The adapter SHALL map `IWorkflowRepository`/`IProjectRepository`/`IInstanceRepository`/
  `IEventRepository` to the engine's `EngineRepositories` ports.
- The adapter SHALL convert the persisted workflow version definition (JSON) into the
  engine's `WorkflowDefinition` model.
- The adapter SHALL convert workflow/state instances between the persistence and engine
  domain models.

#### Scenario: Engine resolves a published workflow

- **WHEN** the engine loads a workflow definition for an instance's pinned version
- **THEN** the adapter unmarshals the persisted definition JSON into an
  `engine.WorkflowDefinition` with nodes and transitions intact

#### Scenario: Engine persists a transition

- **WHEN** the engine executes a transition
- **THEN** the adapter persists the updated workflow-instance state and the appended
  event through the PostgreSQL repositories
