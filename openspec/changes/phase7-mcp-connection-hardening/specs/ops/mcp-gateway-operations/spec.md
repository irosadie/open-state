## Purpose

Make MCP gateway execution observable, bounded, and diagnosable without leaking
provider data or credentials into operational telemetry.

## ADDED Requirements

### Requirement: Connection health and diagnostics

The platform SHALL expose safe per-connection health, last successful verification,
last failure classification, and action-required state to authorized project
operators. It SHALL retain provider error details only in redacted operational records.

#### Scenario: Provider is unavailable

- **WHEN** a gateway call cannot reach a provider
- **THEN** the platform records the classified unavailable state
- **AND** presents safe diagnostics to authorized operators.

### Requirement: Bounded execution resilience

The platform SHALL enforce configured timeouts, concurrency/rate limits, retries only
for idempotent-safe operations, and circuit breaking for unhealthy connections.

#### Scenario: Open circuit

- **WHEN** a provider reaches the configured failure threshold
- **THEN** subsequent calls are rejected or deferred according to circuit policy
- **AND** the provider is not contacted until recovery conditions are met.

### Requirement: Redacted audit and telemetry

The platform SHALL emit structured audit events, metrics, and traces for connection
and gateway actions with tenant/project correlation. Telemetry SHALL exclude credential
values, authorization headers, and unredacted sensitive provider payloads.

#### Scenario: Gateway invocation telemetry

- **WHEN** the gateway invokes a provider tool
- **THEN** telemetry records the connection alias, tool identity, outcome class, duration, and correlation identifier
- **AND** excludes secret values and unredacted payload content.
