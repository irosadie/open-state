## 1. Contract and scenario foundation

- [x] 1.1 Inventory the existing doctor tool names, input schemas, response shapes, and State MCP bindings; record the compatibility rules that every normal doctor fixture must satisfy.
- [x] 1.2 Extend scenario data validation only where needed for doctor records, initial lifecycle state, configured business outcomes, and intentional malformed output; preserve padel/food fixture compatibility and duplicate-tool validation.
- [x] 1.3 Add a small fixture-loading test matrix proving every planned doctor scenario parses successfully and an unknown or invalid scenario fails with an actionable startup error.

## 2. Happy-path consultation fixture

- [x] 2.1 Add an explicit happy-path doctor fixture with synthetic doctors, specialties, schedules, queues, and consultation metadata for lookup, list, search, and recommendation.
- [x] 2.2 Configure the happy path with at least one available schedule and queue, then verify schedule and queue responses contain stable identifiers and bookable statuses.
- [x] 2.3 Verify the happy-path fixture supports the complete reserve → confirm → payment create → payment verify → notification sequence and documents the expected identifiers between steps.

## 3. Catalog and availability outcomes

- [x] 3.1 Add a no-results doctor fixture covering empty lookup/search/list/recommendation results without fabricating doctors or schedules.
- [x] 3.2 Add an unavailable-slot fixture with an explicit unavailable result and deterministic alternative appointment slots.
- [x] 3.3 Add a full-queue fixture with current queue details and a deterministic next available option.
- [x] 3.4 Add provider-level tests for populated, empty, unavailable, full-queue, and missing-schedule responses through the registered MCP tools.

## 4. Stateful appointment lifecycle

- [x] 4.1 Complete or refactor doctor store lifecycle rules for reservation, confirmation, direct booking, cancellation, unknown references, invalid transitions, and rebooking after cancellation.
- [x] 4.2 Add a conflict fixture with pre-existing reservation or booking state and verify a second reservation returns a stable conflict without overwriting the original record.
- [x] 4.3 Add unit tests that assert state changes after reserve, confirm, cancel, and repeated lifecycle calls, including unchanged unrelated records after failures.
- [x] 4.4 Add provider MCP tests that consume identifiers from one tool response as inputs to the next tool instead of relying on hard-coded generated IDs.

## 5. Payment and notification outcomes

- [x] 5.1 Verify successful doctor payment creation and verification are linked to a confirmed booking, return deterministic payment metadata, and reject duplicate payment creation safely.
- [x] 5.2 Add a payment-failed fixture covering create and verify rejection with stable machine-readable reasons while preserving the unpaid booking state.
- [x] 5.3 Verify successful notification delivery returns a deterministic notification identifier and add a notification-failure outcome that never reports delivery as successful.
- [x] 5.4 Add unit/provider coverage for payment and notification retries, unknown booking/payment references, duplicate operations, and lifecycle-safe error responses.

## 6. Provider and protocol failure fixtures

- [x] 6.1 Add a structured provider-error fixture and assert MCP tool failures preserve stable error codes and safe human-readable messages.
- [x] 6.2 Add a timeout fixture with a bounded configurable delay and verify a caller can time out while a later retry remains deterministic after the delay completes.
- [x] 6.3 Add an invalid-output fixture that remains discoverable but returns an explicitly malformed domain payload for response-validation tests.
- [x] 6.4 Add tests that distinguish business rejection, provider error, timeout, and malformed output so one failure class cannot accidentally satisfy another test.

## 7. Curl smoke-test coverage

- [x] 7.1 Extend the provider curl smoke test to select a scenario, initialize MCP, call `tools/list`, and assert the required doctor tools and schemas are present.
- [x] 7.2 Add a happy-path curl sequence that checks catalog, availability, reservation, confirmation, payment, verification, and notification over `http://127.0.0.1:8031/mcp`.
- [x] 7.3 Add focused curl runs for no-results, unavailable, queue-full, conflict, payment-failed, provider-error, timeout, and invalid-output scenarios; treat the intended failure as a passing assertion.
- [x] 7.4 Add an optional State MCP curl section using environment-provided API key, tenant, project, and provider connection settings to verify the same doctor call is gated through `http://127.0.0.1:8030/mcp`.
- [x] 7.5 Ensure curl tests clean up or clearly report the selected mock process and never embed a real API key or production endpoint.

## 8. Documentation and package verification

- [x] 8.1 Update `apps/mcp-provider-mock/README.md` with the doctor scenario matrix, startup commands, expected direct/gateway endpoints, lifecycle examples, and mock-only data warning.
- [x] 8.2 Add a convenient package command or documented loop for running the doctor scenario smoke matrix without changing the default padel/food workflow.
- [x] 8.3 Run provider lint, typecheck, unit tests, build, and curl smoke tests; fix regressions without changing unrelated State MCP behavior.
- [x] 8.4 Run strict OpenSpec validation and confirm every requirement scenario is covered by an implementation test or explicit curl assertion.
