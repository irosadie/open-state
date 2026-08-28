# quality/deterministic-tests Specification

## Purpose

Define a deterministic runtime test suite that exercises the state engine without an
LLM, proving that `event → guard → transition` is deterministic (PRD 126).

## ADDED Requirements

### Requirement: LLM-free deterministic suite

The platform SHALL test the engine deterministically without an LLM.

#### Scenario: Direct engine drive

- **GIVEN** a workflow definition and a context loaded in memory
- **WHEN** an event is processed via the engine
- **THEN** the test asserts the resulting state and status deterministically
- **AND** no LLM is invoked (PRD 126)

### Requirement: Guard operator coverage

The deterministic suite SHALL cover every guard operator the engine supports.

#### Scenario: All operators exercised

- **GIVEN** guards using `==`, `!=`, `>`, `>=`, `<`, `<=`, `IN`, `EXISTS`
- **THEN** each operator has a passing and a failing case
- **AND** each asserts the expected transition or rejection

#### Scenario: AND/OR grouping

- **GIVEN** guard groups with AND/OR logic
- **THEN** the suite asserts combined-condition evaluation
- **AND** a failing condition in an AND group blocks the transition

### Requirement: Priority-ordering determinism

The suite SHALL assert that transitions resolve by priority when multiple candidates
pass.

#### Scenario: Highest-priority transition wins

- **GIVEN** an event with multiple passing candidate transitions of differing priority
- **WHEN** the engine processes the event
- **THEN** it applies the highest-priority transition (PRD 33-34)

### Requirement: Rejection determinism

The suite SHALL assert that invalid events are rejected without side effects.

#### Scenario: Disallowed event

- **GIVEN** an event not allowed from the current state
- **WHEN** the engine processes it
- **THEN** the state does not change
- **AND** a structured rejection (guard failed / disallowed) is returned
