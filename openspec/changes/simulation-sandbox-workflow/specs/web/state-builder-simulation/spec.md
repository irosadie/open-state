## Purpose

Let State Builder operators run the workflow currently on the canvas and understand
its state, transition, guard, context, and mock capability behavior before publishing.

## ADDED Requirements

### Requirement: Author and run a simulation from the current canvas
The State Builder SHALL provide a simulation control that opens a simulation panel
for the current workflow snapshot. The operator SHALL be able to supply an optional
initial JSON context and an ordered list of named events with optional JSON payloads,
then run the simulation without saving or publishing the draft first.

The UI SHALL prevent submission of malformed JSON and clearly surface structural
validation errors that make the current workflow non-executable.

#### Scenario: Run an unsaved draft
- **WHEN** an operator configures context and events for a changed but unsaved canvas
- **THEN** the State Builder runs the submitted canvas snapshot and displays its result
  without creating a saved workflow version

#### Scenario: Invalid JSON is corrected before submission
- **WHEN** an operator enters invalid JSON in the context or event payload editor
- **THEN** the UI identifies the invalid input and does not send a simulation request

### Requirement: Display simulation decisions and mock capability requests
The simulation panel SHALL show the entry state, every processed event, guard result,
selected or rejected transition, state reached, resulting context, and mock capability
requests returned by the simulation. It SHALL make a rejection reason visible without
exposing internal reasoning or provider details.

#### Scenario: Trace shows a successful transition
- **WHEN** an event advances the workflow
- **THEN** the corresponding trace step shows the selected transition, destination
  state, guard outcome, and resulting context

#### Scenario: Trace shows a guard rejection
- **WHEN** no candidate transition passes its guards
- **THEN** the panel shows the failed/rejected step and does not present a fabricated
  next state

### Requirement: Relate a trace step to the workflow graph
The State Builder SHALL let the operator select a trace step and visibly focus the
associated state and selected transition on the React Flow canvas. A simulation result
SHALL be marked stale when the workflow definition changes after the result was run.

#### Scenario: Inspect a trace step on the canvas
- **WHEN** an operator selects a successful event step in the trace
- **THEN** the canvas highlights or focuses its source state, selected transition,
  and destination state

#### Scenario: Canvas changes after a simulation
- **WHEN** an operator changes a node, transition, or workflow metadata after running
  a simulation
- **THEN** the previous result is marked stale and is not represented as applying to
  the changed workflow

### Requirement: Simulation controls do not imply a live execution
The simulation UI SHALL label capability requests as mock/sandbox and SHALL not offer
controls that invoke live MCP, LLM, webhook, or provider execution.

#### Scenario: Capability appears in a trace
- **WHEN** a trace includes a capability request
- **THEN** the UI labels it mock/sandbox rather than as a completed live invocation
