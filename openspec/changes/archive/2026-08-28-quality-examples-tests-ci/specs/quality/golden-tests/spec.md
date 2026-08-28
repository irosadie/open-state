# quality/golden-tests Specification

## Purpose

Define golden conversation tests that replay user utterances and assert the resolved
intent / current state after each turn, providing AI-behavior regression tests for
each example workflow (PRD 125).

## ADDED Requirements

### Requirement: Golden conversation fixtures

The platform SHALL provide golden conversation test fixtures for each example
workflow.

#### Scenario: Fixture per workflow

- **GIVEN** the `padel-court-booking`, `order-food`, and `order-doctor` workflows
- **THEN** each has at least one golden conversation fixture
- **AND** each fixture lists a sequence of `User` utterances with the `Expected`
  state after each turn (PRD 125)

#### Scenario: Expected state per turn

- **WHEN** a fixture replays a user utterance
- **THEN** the expected resolved intent and current state for that turn are declared
- **AND** the harness asserts the actual state equals the expected state

### Requirement: Replay harness

The platform SHALL provide a harness that replays a golden conversation against
intent resolution and the engine.

#### Scenario: Replay a conversation

- **WHEN** the harness replays a fixture's utterances in order
- **THEN** each turn advances the workflow via the resolved intent
- **AND** the harness compares actual vs expected state per turn

#### Scenario: Assertion failure

- **WHEN** an actual state differs from the expected state on a turn
- **THEN** the harness fails with a diff of expected vs actual state for that turn

### Requirement: Deterministic classification stub

Golden tests SHALL not depend on a real LLM for intent classification (PRD 170).

#### Scenario: Stubbed intent resolution

- **WHEN** a golden test runs
- **THEN** intent resolution is stubbed/mocked to return the fixture's expected intent
- **AND** the test asserts the deterministic state-machine outcome, not LLM quality

### Requirement: Regression gate

Golden conversation tests SHALL run as part of the automated test suite.

#### Scenario: Run in CI

- **WHEN** CI runs the test suite
- **THEN** all golden conversation fixtures are executed
- **AND** any behavior regression fails the build (PRD 125)
