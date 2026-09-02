## Purpose

This capability provides a deterministic doctor-provider MCP test double that
can exercise the complete consultation and appointment lifecycle, including
successful outcomes, business rejections, provider failures, and stateful
side effects through the same MCP contract used by a real integration.

## ADDED Requirements

### Requirement: Doctor scenarios are explicitly selectable and discoverable
The provider mock MUST allow a caller to select a named doctor scenario before
the server starts, and MUST expose a stable provider identity and tool catalog
for the selected scenario through MCP initialization and tool discovery. A
scenario selection that cannot be loaded MUST fail at startup with an
actionable error identifying the invalid scenario configuration.

#### Scenario: Start the happy-path doctor scenario
- **WHEN** the provider mock is started with the happy-path doctor scenario
- **THEN** MCP initialization succeeds with a doctor-provider identity and
  `tools/list` exposes the doctor tools required by the consultation flow

#### Scenario: Select an unknown doctor scenario
- **WHEN** the provider mock is started with a scenario path or name that does
  not exist or cannot be parsed
- **THEN** the process fails before accepting MCP requests and reports which
  scenario could not be loaded

#### Scenario: Discover tools without invoking business operations
- **WHEN** an MCP client calls `initialize` followed by `tools/list`
- **THEN** the response contains each tool name, input schema, and description
  needed for an LLM or gateway to select the doctor capability without first
  making a business call

### Requirement: The doctor catalog supports consultation discovery
The provider mock MUST expose deterministic read operations for doctor lookup,
doctor listing, search, and recommendation. Each operation MUST return the
configured scenario data and MUST preserve the declared response shape for
both populated and empty results.

#### Scenario: Look up a known doctor
- **WHEN** the client requests a doctor using an identifier present in the
  scenario data
- **THEN** the provider returns that doctor's profile, specialty, status, and
  supported consultation information

#### Scenario: List doctors by specialty
- **WHEN** the client requests doctors for a specialty configured in the
  scenario
- **THEN** the provider returns a deterministic ordered list containing only
  matching doctors

#### Scenario: Search returns no doctors
- **WHEN** the client searches for a name, specialty, or symptom with no
  configured match
- **THEN** the provider returns a successful empty result with an explicit
  reason or next-step metadata, and MUST NOT fabricate a doctor

#### Scenario: Recommendation is available for a consultation need
- **WHEN** the client requests recommendations for a need represented in the
  scenario data
- **THEN** the provider returns deterministic recommendation candidates with
  enough information for the State flow to ask the user to choose one

### Requirement: The appointment availability flow covers positive and negative outcomes
The provider mock MUST support deterministic schedule and queue checks for a
selected doctor or specialty. The responses MUST distinguish an available
appointment from an unavailable slot, a full queue, and a later alternative,
so the State MCP can keep the LLM behind the availability gate.

#### Scenario: Requested appointment slot is available
- **WHEN** the client checks a configured available date and time
- **THEN** the provider returns availability details including doctor,
  schedule identifier, date, time, and a bookable status

#### Scenario: Requested slot is unavailable with alternatives
- **WHEN** the client checks a configured unavailable slot
- **THEN** the provider returns an explicit unavailable status and deterministic
  alternative slots without presenting the requested slot as bookable

#### Scenario: Queue is full with a later queue option
- **WHEN** the client checks a configured full queue
- **THEN** the provider returns a full-queue status, the current queue details,
  and any configured next available queue or time

#### Scenario: Availability data is missing
- **WHEN** the client checks a doctor, date, or schedule that is absent from
  the scenario data
- **THEN** the provider returns a deterministic not-found or unavailable
  business result and does not create a reservation

### Requirement: The appointment lifecycle is stateful and deterministic
The provider mock MUST support reservation, confirmation, direct booking, and
cancellation with persisted in-process state. Operations MUST validate
references and lifecycle transitions, return stable identifiers, and expose
business conflicts instead of silently overwriting an existing appointment.

#### Scenario: Reserve and confirm an available appointment
- **WHEN** the client reserves an available schedule and then confirms the
  returned reservation
- **THEN** the provider returns a reservation identifier followed by a
  confirmed booking, and subsequent availability no longer treats the same
  schedule as freely bookable

#### Scenario: Reserve an already reserved schedule
- **WHEN** a second client attempts to reserve a schedule already reserved or
  booked in the current mock process
- **THEN** the provider returns a deterministic conflict result and preserves
  the original reservation or booking

#### Scenario: Confirm an unknown reservation
- **WHEN** the client tries to confirm a reservation identifier that does not
  exist or is not confirmable
- **THEN** the provider returns a validation or not-found error and creates no
  booking

#### Scenario: Cancel a confirmed booking
- **WHEN** the client cancels a confirmed booking
- **THEN** the provider marks the booking cancelled, returns the cancellation
  result, and makes the schedule eligible according to the scenario's
  configured rebooking policy

#### Scenario: Cancel an unknown or already cancelled booking
- **WHEN** the client cancels a booking that does not exist or has already
  been cancelled
- **THEN** the provider returns a deterministic lifecycle error without
  changing unrelated bookings

### Requirement: Payment and notification outcomes can be exercised independently
The provider mock MUST expose deterministic payment creation, payment
verification, and notification operations for doctor bookings. Scenarios MUST
be able to represent successful payment, rejected payment, verification
failure, notification failure, and retry-safe repeated calls without exposing
real payment or patient data.

#### Scenario: Create and verify a successful payment
- **WHEN** the client creates payment for a valid confirmed booking and verifies
  the returned payment reference
- **THEN** the provider returns a successful payment status and verification
  result linked to the booking

#### Scenario: Payment is rejected
- **WHEN** the client creates or verifies payment for a booking configured to
  reject payment
- **THEN** the provider returns a deterministic payment failure with a
  machine-readable reason and leaves the booking in the configured unpaid
  state

#### Scenario: Notification is delivered
- **WHEN** the client sends a confirmation notification for a valid confirmed
  booking
- **THEN** the provider returns a delivered status with a deterministic
  notification identifier

#### Scenario: Notification delivery fails
- **WHEN** the client sends a notification in a scenario configured for
  delivery failure
- **THEN** the provider returns a retryable or terminal failure as configured,
  and MUST NOT report the notification as delivered

### Requirement: Provider failures and malformed responses are reproducible
The provider mock MUST include scenarios for transport delay or timeout,
structured provider errors, and invalid response payloads. These scenarios
MUST be selectable independently from business-outcome scenarios so gateway
timeouts, MCP error handling, and response validation can be tested without
external services.

#### Scenario: Doctor provider responds slowly
- **WHEN** the client invokes a tool configured with a delay longer than the
  caller's timeout
- **THEN** the client can observe a timeout or connection failure while the
  mock remains deterministic for a subsequent retry

#### Scenario: Provider returns a structured business error
- **WHEN** the client invokes a tool configured to reject the request
- **THEN** the MCP response contains a structured error or tool failure with a
  stable code and safe human-readable message

#### Scenario: Provider returns an invalid payload
- **WHEN** the client invokes a tool configured with a malformed result
- **THEN** the MCP response remains observable for contract-validation tests,
  and the mock documentation identifies that the payload is intentionally
  invalid

### Requirement: Scenario state is isolated, resettable, and safe for test data
Each mock process MUST start from the selected scenario's declared initial
state. State changes from one scenario process MUST NOT leak into another
process, and restarting the process MUST reset reservations, bookings,
payments, and notifications. Scenario fixtures MUST contain synthetic doctor
and patient-related data only.

#### Scenario: Restart resets appointment state
- **WHEN** a client creates a reservation or booking and the provider process
  is restarted with the same scenario
- **THEN** the new process returns the fixture's initial availability and does
  not recognize identifiers from the previous process

#### Scenario: Two selected scenarios do not share state
- **WHEN** two provider processes run with different doctor scenario fixtures
- **THEN** each process returns only its own configured doctors, schedules,
  outcomes, and identifiers

#### Scenario: Fixture contains no production patient data
- **WHEN** a scenario is inspected or exercised
- **THEN** all doctor, patient, contact, payment, and notification values are
  synthetic and suitable for local tests

### Requirement: Doctor scenarios are verifiable through the real MCP boundary
The provider mock MUST include a repeatable HTTP MCP smoke-test flow that
initializes the server, discovers tools, invokes representative read and
write operations, and verifies both success and failure scenarios. The test
flow MUST be able to run against the provider endpoint directly and document
the optional State MCP gateway path without requiring an external doctor
service.

#### Scenario: Smoke-test the happy path over MCP
- **WHEN** the smoke test initializes the doctor provider, lists tools, checks
  availability, reserves, confirms, pays, and sends a notification
- **THEN** every step returns the expected MCP result and lifecycle state

#### Scenario: Smoke-test a negative path over MCP
- **WHEN** the smoke test runs the unavailable, conflict, payment-failure, or
  provider-error scenario
- **THEN** it asserts the expected failure code or business status and exits
  successfully only when the failure is the intended fixture outcome

#### Scenario: Smoke-test through State MCP when configured
- **WHEN** a State MCP API key, tenant, project, and provider connection are
  configured for the doctor project
- **THEN** the smoke test can invoke the same doctor capability through State
  MCP and verify that the gatekeeper receives the provider result rather than
  bypassing State MCP
