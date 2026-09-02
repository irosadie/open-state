## Why

The doctor provider mock currently exposes one mostly static fixture, so it cannot exercise the complete State MCP gatekeeper flow or the failure paths a client needs to validate. We need deterministic, selectable doctor scenarios that make every meaningful provider outcome observable through the registered MCP connection at `8031`.

## What Changes

- Add selectable doctor-provider scenarios for the full consultation lifecycle.
- Cover happy paths for doctor lookup, catalog search, schedule selection, queue validation, reservation, confirmation, payment, and notification.
- Cover deterministic negative paths: no doctor, unavailable schedule, full queue, reservation conflict, payment rejection, provider timeout, malformed output, and cancellation.
- Keep provider responses stateful where the operation has side effects, while keeping scenario data isolated and resettable per process.
- Make scenario selection explicit through `MCP_PROVIDER_MOCK_SCENARIO`, without changing the OpenState gateway contract or bypassing project bindings.
- Add curl smoke tests that run each scenario through MCP `initialize`, `tools/list`, and relevant `tools/call` sequences.
- Document the scenario matrix, expected outcomes, and example commands for testing the doctor flow through OpenState `8030` to provider `8031`.

## Capabilities

### New Capabilities

- `mcp/doctor-provider-mock-scenarios`: Deterministic doctor MCP provider scenarios for complete success, business failure, resilience, and stateful lifecycle testing.

### Modified Capabilities

- None.

## Impact

- `apps/mcp-provider-mock/fixtures`: new doctor scenario JSON files and shared test data.
- `apps/mcp-provider-mock/src`: scenario loading/selection and any narrowly scoped stateful behavior needed for doctor operations.
- `apps/mcp-provider-mock/test`: unit, integration, and curl-level coverage for doctor outcomes.
- `apps/mcp-provider-mock/README.md`: scenario catalog and reproducible test instructions.
- Local development only: the doctor scenario will be selected for the provider mock on port `8031` when running the OpenState order-doctor verification.

## Non-goals

- This does not add production doctor integrations or real medical data.
- This does not change State MCP authorization, workflow guards, provider registration, or the secure gateway routing model.
- This does not expose provider URLs, credentials, or scenario internals to the LLM.
- This does not make the mock a general-purpose healthcare system; it only provides deterministic contracts needed by the example workflow.
