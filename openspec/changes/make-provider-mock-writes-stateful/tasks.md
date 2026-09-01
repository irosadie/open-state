## 1. Stateful scenario contracts

- [x] 1.1 Define scenario operations, input schemas, and initial data for padel booking/payment writes.
- [x] 1.2 Define scenario operations, input schemas, and initial data for food cart/order/payment writes.
- [x] 1.3 Define scenario operations, input schemas, and initial data for doctor appointment/payment writes.

## 2. In-memory provider lifecycles

- [x] 2.1 Extend the padel store for static booking and payment lifecycle reads/writes.
- [x] 2.2 Implement food cart, order, and payment state with deterministic IDs and error handling.
- [x] 2.3 Implement doctor reservation, booking, cancellation, and payment state with schedule availability updates.
- [x] 2.4 Route MCP tool execution through the correct domain store and ensure failures do not partially mutate state.

## 3. Verification

- [x] 3.1 Add SDK tests for successful and rejected write lifecycles in every domain.
- [x] 3.2 Extend curl smoke tests to perform write-then-read flows for padel, food-order, and doctor.
- [x] 3.3 Verify process-reset behavior and English payloads for every domain.

## 4. Final validation

- [x] 4.1 Run provider mock lint, typecheck, tests, build, and curl smoke tests.
- [x] 4.2 Validate the OpenSpec change strictly and update task completion status.
