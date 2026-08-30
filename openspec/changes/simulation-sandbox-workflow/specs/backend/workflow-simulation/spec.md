## Purpose

Provide a deterministic, side-effect-free way for an operator to execute a
workflow draft and inspect its runtime decisions before the draft is published.

## ADDED Requirements

### Requirement: Simulate the supplied draft snapshot in isolation
The system SHALL simulate the supplied workflow definition, initial context, and
ordered event script without requiring the workflow to have been saved or published.
The simulation SHALL apply the same event eligibility, guard evaluation, priority,
and context-merge rules used by the deterministic workflow engine.

#### Scenario: Valid event script traverses the draft
- **WHEN** an operator supplies a structurally executable draft and an ordered script
  of events whose guards pass
- **THEN** the response shows each resulting state in script order and the final
  state and context match the deterministic engine outcome for that draft

#### Scenario: Same script is repeatable
- **WHEN** the same draft, initial context, and event script are simulated repeatedly
- **THEN** the selected transitions, guard outcomes, final state, and final context
  are identical on every run

### Requirement: Return an inspectable runtime trace
The system SHALL return a trace beginning at the entry state and containing, for
each processed event, the input event, state before processing, matching candidate
transitions with guard outcomes and priorities, the selected transition when one
exists, state after processing when a transition succeeds, and the resulting context.
Each entered state SHALL include its declared capability requests.

#### Scenario: Guard selects the highest-priority passing transition
- **WHEN** multiple transitions match an event and more than one guard passes
- **THEN** the trace identifies all candidates and selects the passing transition
  with the lowest numeric priority

#### Scenario: Event cannot progress the workflow
- **WHEN** an event is not allowed from the current state or no matching transition
  passes its guards
- **THEN** the trace records the structured rejection reason and the simulation
  stops at that event without fabricating a later state transition

### Requirement: Sandbox execution has no external or persistent effects
The system SHALL execute simulation using ephemeral state only. It SHALL NOT create
or change workflow instances, events, workflow definitions, workflow versions,
contexts, or audit records. It SHALL NOT call an LLM, MCP server, webhook, live
provider, or arbitrary tenant-supplied code.

Capability requests in a trace SHALL be marked as mock/sandbox requests; they
describe what the workflow would request and do not represent an external execution.

#### Scenario: Capability-bearing state is simulated
- **WHEN** a simulated transition enters a state declaring one or more capabilities
- **THEN** the trace lists those capability requests as mock/sandbox and no external
  provider invocation occurs

#### Scenario: Simulation does not leave runtime data behind
- **WHEN** a simulation completes or is rejected
- **THEN** no persisted runtime instance, event, context, or audit record is created

### Requirement: Reject non-executable simulation input safely
The system SHALL reject a simulation request whose definition lacks the data needed
to begin execution or whose event script is malformed. The rejection SHALL be a
structured validation error and SHALL NOT partially persist or execute any data.

#### Scenario: Draft has no executable entry state
- **WHEN** an operator submits a definition with no entry state
- **THEN** the system returns a validation error before processing the event script
