# MCP Provider Mock

`@openstate/mcp-provider-mock` is a development and test-only stand-in for a
third-party data MCP provider. It is separate from the OpenState MCP server:

```text
LLM / OpenState MCP (control plane) → capability authorization
LLM host → MCP Provider Mock (data plane)
```

The mock exposes a Streamable HTTP endpoint at `/mcp`, not a REST JSON API. Its
tools and behavior come from one JSON scenario loaded into memory at startup.
Doctor outcome fixtures extend the compatibility doctor catalog and override
only the data or tool result needed by that scenario.

## Run locally

```bash
bun run dev:provider-mock
```

Default endpoint: `http://127.0.0.1:8031/mcp`

Health endpoints:

- `http://127.0.0.1:8031/health/live`
- `http://127.0.0.1:8031/health/ready`

## Scenario catalog

The default scenario is `fixtures/padel.json`. Each JSON scenario is written in
English and exposes one provider domain. Doctor fixtures cover the complete
consultation lifecycle and are safe synthetic data only:

| Fixture | Outcome covered |
| --- | --- |
| `doctor.json` | Compatibility baseline with doctor catalog, unavailable slot, and stateful appointment tools |
| `doctor-happy.json` | Available schedule and queue through reservation, confirmation, payment, and notification |
| `doctor-no-results.json` | Empty lookup, list, search, and recommendation results |
| `doctor-unavailable.json` | Unavailable slot with deterministic alternatives |
| `doctor-queue-full.json` | Full queue with next available option |
| `doctor-conflict.json` | Pre-existing booking and reservation conflict |
| `doctor-payment-failed.json` | Payment creation and verification rejection |
| `doctor-notification-failed.json` | Notification delivery failure |
| `doctor-provider-error.json` | Structured provider error |
| `doctor-timeout.json` | Delayed schedule response for client timeout tests |
| `doctor-invalid-output.json` | Intentionally malformed doctor result |

The other domain fixtures remain available:

- `fixtures/padel.json` — padel court search, availability, booking, payment,
  notifications, and the dynamic demo tools `padel.cek_available` and
  `padel.create_booking`.
- `fixtures/food-order.json` — food menu, cart, order, payment, and delivery
  tracking.

Override the scenario without changing the fixture file:

```bash
MCP_PROVIDER_MOCK_SCENARIO=./fixtures/padel-error.json bun run start
```

Use `MCP_PROVIDER_MOCK_PORT` to choose another port. Scenario state lives only
in memory; restarting the process resets slots, carts, orders, appointments,
payments, and notifications.

To run a doctor outcome directly on the local provider endpoint:

```bash
MCP_PROVIDER_MOCK_SCENARIO=./fixtures/doctor-happy.json \
MCP_PROVIDER_MOCK_PORT=8031 bun run start
```

Replace `doctor-happy.json` with any fixture from the doctor matrix. The
provider mock is the data-plane MCP at `http://127.0.0.1:8031/mcp`.

## Verify MCP with curl

Run the protocol-level smoke test:

```bash
bun run test:curl
```

It starts each scenario on a temporary local port, then uses `curl` for MCP
`initialize`, `tools/list`, and `tools/call` requests before stopping the
process. The full matrix covers padel, food, the compatibility doctor fixture,
all doctor outcomes, and the doctor timeout retry.

Run only the doctor matrix with:

```bash
bun run test:curl:doctor
```

The timeout case intentionally prints a curl timeout before asserting that a
subsequent retry remains deterministic.

## Verify through the secure State MCP gateway

When the doctor provider is registered to a project connection, the LLM or MCP
host should call State MCP at `http://127.0.0.1:8030/mcp`. In secure mode,
State MCP resolves the project binding and invokes the provider mock at `8031`
internally; the client does not select or connect to the provider directly.

The curl matrix can optionally verify this path. Provide runtime-only values
through the environment; never commit an API key:

```bash
STATE_MCP_TOKEN="osk_..." \
STATE_MCP_TENANT="tenant-id" \
STATE_MCP_PROJECT="project-id" \
STATE_MCP_PROVIDER_ALIAS="doctor-provider-mock" \
STATE_MCP_INSTANCE="workflow-instance-id" \
STATE_MCP_CAPABILITY="doctor.lookup" \
STATE_MCP_PAYLOAD_JSON='{}' \
bun run test:curl:doctor
```

The optional assertion requires secure gateway discovery (`invoke_capability`)
and verifies the capability is invoked through State MCP. The tenant and
project identify the registered scope; the provider alias is an operator-side
label used to confirm which connection the test is exercising. Provider
endpoint and credentials stay inside State MCP configuration.

## Included padel tools

- `padel.cek_available` — returns available court slots for a venue and date.
- `padel.create_booking` — reserves an available slot and rejects duplicate
  reservations for the lifetime of the process.

Food and doctor write tools are stateful too: food cart items become orders and
payments, while doctor reservations become appointments that can be paid and
cancelled. All generated identifiers are deterministic for repeatable tests.

`fixtures/padel-error.json` returns a deterministic provider tool error.
`fixtures/padel-delay.json` delays availability responses to test caller
timeouts.

This app does not contain production credentials, databases, workflow rules, or
state-authorization logic. It is a local third-party data-plane test double.
OpenState remains responsible for declaring the logical capability, resolving
the project-scoped provider binding, enforcing the gate, and accepting
execution evidence before a state transition.
