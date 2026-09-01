## 1. Gateway mode and authorization model

- [x] 1.1 Verify Phase 3–5 connection, catalog, binding, evidence, and capability security-chain contracts.
- [x] 1.2 Add an explicit gateway/advisory runtime configuration with compatibility-safe default and deployment documentation.
- [x] 1.3 Define resolved gateway invocation input/output models that exclude raw provider URL, alias, credential, and arbitrary tool selection.
- [x] 1.4 Implement application-layer authorization that resolves tenant/project/instance/current state/required capability/verified binding before outbound work.

## 2. Provider execution path

- [x] 2.1 Extend the capability provider port and MCP client adapter to accept only an internally resolved project connection/tool target.
- [x] 2.2 Implement provider input-schema validation, output normalization/validation, safe error classification, and timeout propagation.
- [x] 2.3 Implement idempotency handling for gateway requests and integrate successful normalized results with Phase 2 execution evidence/context persistence.
- [x] 2.4 Ensure provider failures, invalid outputs, disabled tools, and unavailable bindings cannot satisfy a state requirement or transition workflow state.

## 3. State MCP surface

- [x] 3.1 Evolve the authorized capability invocation tool into the state-gated gateway operation without adding provider passthrough tools.
- [x] 3.2 Update State MCP instructions, tool descriptions, and projections to distinguish secure gateway from advisory direct-two-MCP mode.
- [x] 3.3 Ensure secure mode emits no provider URLs, credential references, unrestricted catalogs, or raw provider errors.
- [x] 3.4 Preserve direct result reporting only as documented advisory compatibility behavior.

## 4. Deployment and verification

- [x] 4.1 Add gateway integration fixtures using the provider mock for read and write capability execution.
- [x] 4.2 Add tests for authorized success, unauthorized bypass attempt, missing binding, disabled tool, timeout, invalid output, duplicate idempotency key, and evidence-gated transition.
- [x] 4.3 Add secure-mode MCP smoke tests proving the LLM needs only State MCP while the provider remains reachable internally.
- [x] 4.4 Document local, advisory, and secure gateway deployment profiles with migration/rollback instructions.
- [x] 4.5 Run Go tests/vet, provider mock tests, MCP curl smoke tests, web tests/typecheck/lint, and OpenSpec validation.
