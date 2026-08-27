# capability/test-invocation Specification

## Purpose

Define an HTTP endpoint to test or simulate a capability's execution in sandbox/mock
mode, so operators can validate schema and provider wiring without live side effects.

## ADDED Requirements

### Requirement: Test capability invocation

The platform SHALL expose an HTTP endpoint to test-invoke a capability in sandbox/mock
mode.

#### Scenario: Successful mock invocation

- **WHEN** an operator submits an invocation payload for a capability
- **THEN** the platform runs the security chain and invokes it through the mock provider
- **AND** returns the normalized result flagged as coming from the mock/sandbox provider

#### Scenario: Schema validation failure

- **WHEN** the payload fails the capability's input schema validation
- **THEN** the platform rejects the request with a validation error
- **AND** does NOT invoke any provider

#### Scenario: Unauthorized or unknown capability

- **WHEN** the capability is not authorized for the requested context or does not exist
- **THEN** the platform returns an authorization or not-found error

### Requirement: Sandbox guarantee

The test endpoint SHALL never trigger live external side effects.

#### Scenario: No live call

- **WHEN** the test endpoint runs
- **THEN** it executes through the mock/sandbox provider
- **AND** never calls a real external MCP provider
