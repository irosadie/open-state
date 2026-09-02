## Context

The provider mock already serves a scenario-defined MCP catalog and has a
stateful store for doctor appointment operations. The current doctor fixture
mixes useful catalog data with an unavailable-appointment outcome, so it is
not sufficient for exercising a complete consultation flow or isolated
failure handling. See `proposal.md` for the motivation and
`specs/mcp/doctor-provider-mock-scenarios/spec.md` for the observable
contract.

The design must preserve the existing provider mock boundary at
`/mcp`, keep the State MCP as the gatekeeper in end-to-end tests, and avoid
introducing real healthcare, payment, or patient data. Existing padel and
food scenarios must continue to work unchanged.

## Goals / Non-Goals

**Goals:**

- Make every important doctor-flow outcome selectable and reproducible from a
  local fixture.
- Exercise both read-only discovery and stateful write operations using the
  same MCP tool schemas a real provider would expose.
- Keep scenarios independent so one test can target one business or transport
  failure without depending on test order.
- Make the scenario matrix easy to run through direct provider MCP and through
  the configured State MCP gateway.
- Keep the default local behavior backward compatible for existing padel,
  food, and doctor smoke tests.

**Non-Goals:**

- No production doctor-provider adapter or external service credentials.
- No changes to State MCP authorization, workflow semantics, or capability
  routing.
- No attempt to model clinical advice, medical records, payment settlement, or
  notification delivery outside the boundaries needed for orchestration tests.
- No concurrent multi-tenant simulation inside one provider process.

## Decisions

### 1. Use explicit fixture files for scenario selection

Add a small, named fixture set under the provider mock's existing fixtures
directory and select one fixture per process through the existing scenario
configuration mechanism. Keep `doctor.json` as the compatibility baseline and
add a clearly named happy-path fixture plus focused outcome fixtures, for
example:

| Fixture family | Primary purpose |
| --- | --- |
| `doctor-happy.json` | Full available consultation, booking, payment, and notification flow |
| `doctor-no-results.json` | Empty catalog/search/recommendation outcomes |
| `doctor-unavailable.json` | No slot plus deterministic alternatives |
| `doctor-queue-full.json` | Full queue with next available option |
| `doctor-conflict.json` | Existing reservation/booking and lifecycle conflicts |
| `doctor-payment-failed.json` | Payment rejection and verification failure |
| `doctor-notification-failed.json` | Notification delivery failure |
| `doctor-provider-error.json` | Structured provider errors |
| `doctor-timeout.json` | Delayed operations exceeding client timeout |
| `doctor-invalid-output.json` | Intentionally malformed result for validation tests |

The exact filenames can be adjusted during implementation, but each named
scenario must have one obvious command-line/environment selection. The
scenario files may extend the compatibility `doctor.json` and override only
the data or tool outcome they are testing; the loader resolves the inheritance
before exposing the final validated scenario. A process per scenario gives
clean state isolation and keeps the MCP catalog stable for the duration of a
test.

**Alternatives considered:**

- One large fixture with a request parameter controlling outcomes: rejected
  because it makes the tool contract ambiguous to an LLM and allows a caller
  to accidentally switch behavior mid-flow.
- Runtime admin endpoints to mutate mock behavior: rejected because they add
  another control plane and make curl tests order-dependent.
- Separate mock servers for every outcome: rejected because it duplicates the
  server lifecycle and obscures that all outcomes implement the same provider
  contract.

### 2. Keep the MCP tool catalog stable across normal doctor scenarios

Normal doctor fixtures should expose the same logical tool names and input
schemas: lookup, catalog/list, search, recommendation, schedule check, queue
check, reserve, confirm, cancel, direct booking, payment create, payment
verify, and notification. Scenario differences should primarily affect data,
initial state, and declared deterministic outcomes. This lets the State MCP
binding and an LLM discover one contract while tests switch business outcomes.

Failure-focused fixtures may use the existing tool-level error and delay
configuration. The invalid-output fixture is intentionally the exception: it
must remain discoverable but return a documented payload that violates the
expected domain shape, allowing the gateway/client validation path to be
tested.

**Alternatives considered:**

- Hide tools that do not succeed in a scenario: rejected because discovery
  would no longer represent the provider contract and would let tests pass
  only because a capability disappeared.
- Add a generic `doctor.run_scenario` tool: rejected because it teaches the
  client an artificial abstraction instead of validating the real business
  tool surface.

### 3. Extend the existing state store with fixture-driven doctor records

Represent doctors, specialties, schedules, queues, lifecycle records,
payments, and notifications as synthetic records in scenario data. The store
will copy fixture data at startup, generate deterministic IDs for newly
created records, validate references and transitions, and return explicit
business errors for conflicts or unknown identifiers. The store remains
in-process; restarting the server is the reset operation.

For the happy path, the initial schedule and queue must be bookable. For
negative business scenarios, the fixture declares the unavailable/full/
conflicted/payment-failed state rather than relying on wall-clock time or
randomness. Any rebooking behavior after cancellation must be represented by
the fixture's declared policy and covered by a test.

**Alternatives considered:**

- Use an in-memory database or PostgreSQL: rejected because it adds startup
  dependencies to a test double whose value is deterministic local execution.
- Generate schedules from the current date: rejected because it makes curl
  examples and CI results expire or vary by timezone.

### 4. Separate business outcomes from transport/protocol failures

Business fixtures return valid MCP tool results containing stable status and
error codes. Provider-error, timeout, and malformed-output fixtures target
different client behavior:

- structured errors test propagation and safe messaging;
- delays test caller timeout and retry behavior;
- malformed output tests response validation.

This separation ensures that a failed smoke test identifies whether the
business flow, provider error propagation, timeout handling, or schema
validation is broken.

**Alternatives considered:**

- Simulate every failure by terminating the server: rejected because it cannot
  distinguish a provider rejection from process health and makes retry tests
  flaky.
- Encode timeouts as errors immediately: rejected because it does not test the
  actual client timeout boundary.

### 5. Test in layers with curl as the black-box contract

Keep unit tests for scenario parsing and store lifecycle rules, provider-level
tests for MCP registration and tool responses, and a shell smoke test for the
HTTP MCP boundary. The smoke test will initialize, list tools, and call the
representative read/write tools for at least one success and the major
failure classes. When State MCP credentials and a registered connection are
available, an optional section will invoke the same capability through
`8030`, proving the provider is reached behind the State gatekeeper rather
than called directly by the client.

The direct provider endpoint remains configurable, with the local default
`http://127.0.0.1:8031/mcp`. The State MCP endpoint remains configurable, with
the local default `http://127.0.0.1:8030/mcp`. No API key is committed to a
fixture, README, or test script; callers provide it through an environment
variable.

**Alternatives considered:**

- Browser-only verification: rejected because the provider contract is MCP
  and must be testable without the web UI.
- Direct store-only tests: rejected because they would not catch incorrect MCP
  schemas, tool registration, or serialized error results.

## Risks / Trade-offs

- **[Fixture matrix becomes repetitive]** → Share only test helpers and schema
  conventions in code; keep scenario differences explicit in JSON so a reader
  can understand each outcome without tracing hidden branching.
- **[A happy fixture drifts from State MCP bindings]** → Assert all bound
  doctor tools are present in `tools/list` and run the optional gateway smoke
  test when the local State MCP setup is available.
- **[Delay tests slow the normal suite]** → Keep delays short by default and
  run the intentionally long timeout case only in the curl/integration test
  or with an explicit scenario selection.
- **[Malformed-output tests are mistaken for product defects]** → Name the
  fixture and test explicitly as invalid-output, and assert that the failure
  is expected rather than treating any parse failure as a passing result.
- **[Changing doctor.json breaks existing local flows]** → Preserve its
  current contract and outcome; make the new happy path an explicit fixture.
- **[Synthetic data is accidentally reused as real patient data]** → Label all
  records as mock-only in fixture metadata and documentation, and keep values
  obviously non-production.

## Migration Plan

1. Add the scenario fixtures and extend parsing/store behavior without changing
   the existing default scenario selection.
2. Add unit/provider/curl coverage and run the provider package checks.
3. Update the provider README with the scenario matrix and direct/gateway
   commands.
4. Run the State MCP gateway smoke test against the local registered doctor
   connection when services are available.
5. If a fixture causes a regression, revert only the new fixture selection or
   select the compatibility `doctor.json`; no database migration or production
   rollback is required.
