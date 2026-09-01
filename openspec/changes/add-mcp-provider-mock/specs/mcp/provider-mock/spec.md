## Purpose

Provide a deterministic external MCP provider simulator so OpenState can test
state-authorized data operations over the same MCP transport used by real providers.

## ADDED Requirements

### Requirement: Streamable MCP provider interface
The system SHALL provide a development and test-only MCP provider endpoint that
supports Streamable HTTP initialization, tool discovery, and tool invocation, as
well as unauthenticated live and ready health endpoints.

#### Scenario: Discover configured provider tools
- **WHEN** an MCP client initializes a session and lists tools
- **THEN** the provider returns only the tools declared by its active scenario
- **AND** each returned tool includes its configured name, description, and input schema

#### Scenario: Check provider readiness
- **WHEN** a caller requests the provider ready endpoint after a valid scenario has loaded
- **THEN** the endpoint returns a successful ready result

### Requirement: Scenario-driven provider behavior
The provider mock SHALL load a JSON scenario before becoming ready and SHALL use
that scenario as the source of truth for its visible tool contracts, deterministic
responses, and configured failure behaviors.

#### Scenario: Reject an invalid scenario
- **WHEN** the configured scenario is unreadable or fails validation
- **THEN** the provider does not report ready
- **AND** it reports a startup error that identifies the scenario configuration problem without exposing secrets

#### Scenario: Return a configured successful result
- **WHEN** a client invokes a configured tool with valid input for a successful scenario
- **THEN** the provider returns the scenario's structured result
- **AND** the result is deterministic for the same scenario state and input

#### Scenario: Reject invalid tool input
- **WHEN** a client invokes a configured tool with input that does not satisfy the tool schema
- **THEN** the provider returns an MCP tool error
- **AND** it does not apply a write operation or mutate scenario state

#### Scenario: Simulate a configured provider failure
- **WHEN** a client invokes a tool configured with an error or delayed outcome
- **THEN** the provider returns the configured tool error or delays its response for the configured duration
- **AND** it does not perform an unconfigured side effect

### Requirement: Padel read and write reference scenario
The provider mock SHALL ship a padel scenario that exposes a read-only availability
tool and a booking write tool, allowing capability mappings to be exercised against
representative third-party data operations.

#### Scenario: Check available padel slots
- **WHEN** a client invokes `padel.cek_available` with a valid venue and date
- **THEN** the provider returns the configured available slots for that venue and date

#### Scenario: Create a padel booking
- **WHEN** a client invokes `padel.create_booking` with a valid available slot
- **THEN** the provider records the booking in its in-memory scenario state
- **AND** returns a deterministic booking reference

#### Scenario: Prevent duplicate booking of a slot
- **WHEN** a client invokes `padel.create_booking` for a slot already booked in the active scenario state
- **THEN** the provider returns a business tool error
- **AND** it does not create another booking

### Requirement: Isolated and non-production test data
The provider mock SHALL keep all scenario data in memory and SHALL reset it when
the process starts, without connecting to OpenState production databases, storing
credentials, or contacting external provider systems.

#### Scenario: Reset mutable scenario data
- **WHEN** the provider mock restarts with the same scenario
- **THEN** the scenario returns to the fixture's initial availability and booking state

#### Scenario: Run without production dependencies
- **WHEN** the provider mock starts for local development or automated tests
- **THEN** it requires only its scenario configuration and local network port
- **AND** it does not require a database connection or real provider credential
