# quality/load-tests Specification

## Purpose

Define a baseline load test that measures state-transition throughput so operators
get an initial performance signal for the engine (PRD §123 load test).

## ADDED Requirements

### Requirement: Baseline throughput measurement

The platform SHALL provide a load test that measures state-transition throughput.

#### Scenario: Measure transitions per second

- **WHEN** the load test runs a fixed number of events through the engine
- **THEN** it reports the achieved transitions/second
- **AND** the result is printed/exported as a baseline metric

### Requirement: In-memory baseline

The load test SHALL include a deterministic in-memory (no external dependency) run.

#### Scenario: Pure engine run

- **GIVEN** a workflow and context in memory
- **WHEN** the load test processes a batch of events
- **THEN** it measures the pure engine throughput without DB/network contention

### Requirement: Persistence-backed run

The load test SHALL include an optional persistence-backed run to capture the
end-to-end cost.

#### Scenario: Postgres-backed run

- **WHEN** the load test runs against the Postgres-backed repository
- **THEN** it measures throughput including persistence
- **AND** reports the number alongside the in-memory baseline

### Requirement: Non-flaky bound

The load test SHALL avoid tight timing assertions that are flaky on shared machines.

#### Scenario: Loose lower bound

- **WHEN** the load test asserts a minimum throughput
- **THEN** the bound is deliberately loose (guards against gross regressions)
- **AND** the measured value is reported rather than a tight pass/fail threshold
