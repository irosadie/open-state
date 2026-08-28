# capability/test-ui Specification

## Purpose
Define the admin UI for testing a capability's execution in sandbox/mock mode and
viewing the normalized result, consuming the capability test endpoint.

## Requirements

### Requirement: Test capability invocation

The admin UI SHALL let an admin test-invoke a capability in sandbox/mock mode.

#### Scenario: Submit test payload

- **WHEN** an admin enters an invocation payload for a capability and submits
- **THEN** the UI calls the test endpoint
- **AND** shows the normalized result and duration
- **AND** clearly indicates the result came from the mock/sandbox provider

#### Scenario: Show failure

- **WHEN** the test invocation fails
- **THEN** the UI shows the classified error kind and code
- **AND** never shows a raw provider error

### Requirement: Sandbox indication

The admin UI SHALL make it explicit that test execution is sandboxed and non-invasive.

#### Scenario: Sandbox badge

- **WHEN** a capability is being tested
- **THEN** the UI displays a sandbox/mock badge
- **AND** explains that no live external call is made
