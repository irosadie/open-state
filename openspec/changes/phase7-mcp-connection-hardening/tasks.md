## 1. Secret and OAuth lifecycle

- [ ] 1.1 Review existing secrets-management and provider credential paths and define a secret-store port with development and production adapters.
- [ ] 1.2 Move MCP bearer credentials to protected secret references with replace, rotate, revoke, and safe status operations.
- [ ] 1.3 Implement OAuth connection authorization lifecycle with protected state/PKCE, callback validation, secret storage, disconnect, and action-required status.
- [ ] 1.4 Implement server-side OAuth refresh/expiry handling without exposing token artifacts.
- [ ] 1.5 Update Admin UI to start/disconnect OAuth and render only safe lifecycle status.

## 2. Outbound and STDIO security

- [ ] 2.1 Define configurable outbound MCP egress policy for schemes, ports, host/network allowlists, redirects, and local development profiles.
- [ ] 2.2 Implement DNS/IP validation and connect-time revalidation that blocks prohibited private/link-local/loopback destinations outside explicit development policy.
- [ ] 2.3 Apply outbound policy consistently to connection test, discovery, OAuth metadata, token exchange, and gateway invocation.
- [ ] 2.4 Replace free-form STDIO execution with reviewed runner profiles, executable/argument/environment allowlists, isolation, and resource limits.

## 3. Gateway resilience and operations

- [ ] 3.1 Add per-connection timeout, concurrency, rate-limit, idempotency-aware retry, and circuit-breaker policies around provider execution.
- [ ] 3.2 Add safe health state transitions, manual diagnostic/test/reset controls, and action-required status for authorized operators.
- [ ] 3.3 Emit redacted structured audit events, metrics, and traces with tenant/project/connection/tool/correlation identifiers.
- [ ] 3.4 Add operator diagnostics views and deployment/incident-response documentation for credential rotation, provider outage, and blocked egress.

## 4. Security verification

- [ ] 4.1 Add secret-leakage tests covering HTTP responses, State MCP tools, logs/audits, workflow definitions, and frontend state.
- [ ] 4.2 Add OAuth lifecycle tests for callback validation, expiry, refresh failure, disconnect, and token redaction.
- [ ] 4.3 Add egress tests for forbidden URLs, redirect bypass, DNS rebinding, local development allowance, and blocked private networks.
- [ ] 4.4 Add STDIO profile, circuit-breaker, rate-limit, telemetry, and recovery tests.
- [ ] 4.5 Run security-focused integration tests plus Go tests/vet, web tests/typecheck/lint, MCP smoke tests, and OpenSpec validation.
