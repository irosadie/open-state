## 1. Fixture migration preparation

- [x] 1.1 Inventory every padel, food-order, and doctor entry in the API capability fixture catalog and map it to its provider scenario and MCP tool name.
- [x] 1.2 Extend the provider mock scenario contract and static-result execution path only as needed to represent the migrated payloads and input schemas.

## 2. Domain provider scenarios

- [x] 2.1 Create the padel provider scenario containing migrated search, availability, booking, payment, and notification fixture responses.
- [x] 2.2 Create the food-order provider scenario containing migrated menu, cart, order, payment, and tracking fixture responses.
- [x] 2.3 Create the doctor provider scenario containing migrated lookup, catalog, schedule, queue, recommendation, booking, payment, and notification fixture responses.
- [x] 2.4 Add provider-app tests proving scenario-specific tool discovery and representative static fixture results for every domain.

## 3. API fixture migration

- [x] 3.1 Find and update API tests or configuration that consume migrated provider fixture keys to invoke the provider mock instead.
- [x] 3.2 Remove only migrated padel, food-order, and doctor entries from the API capability fixture catalog while preserving unrelated generic fixture entries.

## 4. Curl MCP smoke testing and documentation

- [x] 4.1 Add a repeatable shell smoke-test command that starts the mock on an isolated port, waits for readiness, and always cleans up its process.
- [x] 4.2 Use curl JSON-RPC requests to validate `initialize`, `tools/list`, and representative `tools/call` flows for padel, food-order, and doctor scenarios.
- [x] 4.3 Document scenario selection and the curl smoke-test command in the provider mock README.

## 5. Validation

- [x] 5.1 Run provider mock lint, typecheck, unit tests, build, and curl smoke tests.
- [x] 5.2 Run affected API tests and strict OpenSpec validation, then update completed task status.
