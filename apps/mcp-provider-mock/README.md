# MCP Provider Mock

`@openstate/mcp-provider-mock` is a development and test-only stand-in for a
third-party data MCP provider. It is separate from the OpenState MCP server:

```text
LLM / OpenState MCP (control plane) → capability authorization
LLM host → MCP Provider Mock (data plane)
```

The mock exposes a Streamable HTTP endpoint at `/mcp`, not a REST JSON API. Its
tools and behavior come from one JSON scenario loaded into memory at startup.

## Run locally

```bash
bun run dev:provider-mock
```

Default endpoint: `http://127.0.0.1:8031/mcp`

Health endpoints:

- `http://127.0.0.1:8031/health/live`
- `http://127.0.0.1:8031/health/ready`

## Select a scenario

The default scenario is `fixtures/padel.json`. Each JSON scenario is written in
English and exposes only the tools for one provider domain:

- `fixtures/padel.json` — padel court search, availability, booking, payment,
  notifications, and the dynamic demo tools `padel.cek_available` and
  `padel.create_booking`.
- `fixtures/food-order.json` — food menu, cart, order, payment, and delivery
  tracking.
- `fixtures/doctor.json` — doctor search, schedule, queue, recommendation,
  appointment, payment, and notification tools.

Override the scenario without changing the fixture file:

```bash
MCP_PROVIDER_MOCK_SCENARIO=./fixtures/padel-error.json bun run start
```

Use `MCP_PROVIDER_MOCK_PORT` to choose another port. Scenario state lives only
in memory; restarting the process resets slots, carts, orders, appointments,
payments, and notifications.

## Verify MCP with curl

Run the protocol-level smoke test:

```bash
bun run test:curl
```

It starts each scenario on a temporary local port, then uses `curl` for MCP
`initialize`, `tools/list`, and write-then-read `tools/call` requests before
stopping the process.

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
state-authorization logic. The LLM/MCP host connects to this server separately
from State MCP and maps a trusted alias (for example `padel-provider-mock`) to
this endpoint. OpenState remains responsible for declaring the logical
capability and accepting execution evidence before a state transition.
