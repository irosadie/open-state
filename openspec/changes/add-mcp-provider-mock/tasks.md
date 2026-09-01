## 1. Workspace and local runtime setup

- [x] 1.1 Create the `apps/mcp-provider-mock` Bun workspace with development, build, and test commands.
- [x] 1.2 Add an opt-in root/Turbo command for starting only the provider mock without adding it to the default development stack.
- [x] 1.3 Document the local endpoint, scenario selection, and the distinction between State MCP and provider mock.

## 2. Scenario contract and in-memory provider state

- [x] 2.1 Define and validate the JSON scenario contract for provider identity, tool metadata, input schemas, initial data, and outcomes.
- [x] 2.2 Implement startup failure and readiness behavior for missing or invalid scenarios.
- [x] 2.3 Implement isolated in-memory scenario state with deterministic reset on process start.
- [x] 2.4 Add the padel availability and booking scenario fixtures, including success, invalid-input, business-error, and delayed outcomes.

## 3. MCP provider behavior

- [x] 3.1 Implement Streamable HTTP MCP initialization, tool discovery, and live/ready health endpoints.
- [x] 3.2 Register only tools declared by the active scenario and enforce their declared input schemas.
- [x] 3.3 Implement `padel.cek_available` against the active scenario state.
- [x] 3.4 Implement `padel.create_booking` with deterministic booking references and duplicate-slot rejection.
- [x] 3.5 Implement configured MCP tool-error and delayed-response behavior without unconfigured state mutation.

## 4. Provider mock verification

- [x] 4.1 Add provider-app tests for scenario validation, readiness, and MCP tool discovery.
- [x] 4.2 Add provider-app tests for padel availability, booking mutation, duplicate rejection, input validation, and fixture reset.
- [x] 4.3 Add provider-app tests for configured tool errors and delayed responses.

## 5. OpenState adapter integration coverage

- [x] 5.1 Add an API integration-test harness that starts the provider mock on an isolated local port.
- [x] 5.2 Verify `MCPProvider` initializes and invokes mapped padel tools against the provider mock.
- [x] 5.3 Verify OpenState handling of a provider tool error and caller timeout against the mock.

## 6. Final validation

- [x] 6.1 Run provider-mock tests, API integration tests, and affected workspace lint/type checks.
- [x] 6.2 Validate the completed OpenSpec change and update task status with the executed checks.
