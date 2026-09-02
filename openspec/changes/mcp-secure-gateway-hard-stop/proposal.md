## Why

The secure State MCP gateway correctly blocks capabilities that are not declared by the active state, but the LLM can still interpret a failed provider call as permission to search broader scopes and try a substitute tool. This makes the doctor flow appear to bypass the state gate and makes provider configuration failures difficult to diagnose.

## What Changes

- Make secure `invoke_capability` failures explicit, machine-readable hard stops.
- Instruct secure MCP clients to stop after any failed gateway invocation and never select a fallback capability, scope, provider, or tool.
- Clarify that capability and context discovery tools are diagnostic/read-only and cannot authorize a provider call or state transition.
- Carry idempotency keys through secure event proposals so an LLM retry cannot append the same transition twice.
- Add regression coverage for provider mapping failures, provider failures, hard-stop responses, and repeated event proposals.

## Capabilities

### New Capabilities

- `mcp/secure-gateway-guardrails`: Safe failure semantics and client-facing hard-stop guidance for secure State MCP execution.

### Modified Capabilities

- `openspec/specs/mcp/orchestrator-tools`: The secure gateway execution and event proposal contracts gain explicit failure and retry semantics.

## Impact

- `apps/api/internal/interfaces/mcp/server.go` and `tools.go` secure MCP tool descriptions and response projections.
- `apps/api/internal/application/services/mcp_gateway_service.go` and tests.
- Orchestrator event proposal plumbing and tests for durable idempotency.
- OpenSpec MCP contracts and the provider-mock compatibility harness; no provider endpoint or credential is exposed.

## Non-goals

- Automatically choosing or registering an external provider.
- Allowing tenant/workflow capabilities to bypass the active state.
- Changing the provider mock fixture catalog or exposing provider routing to the LLM. Static canned mock tools may accept caller context when their fixture declares no inputs.
