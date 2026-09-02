## 1. Secret and OAuth lifecycle

- [x] 1.1 Review existing secrets-management and provider credential paths and define a secret-store port with development and production adapters.
- [x] 1.2 Move MCP bearer credentials to protected secret references with replace, rotate, revoke, and safe status operations.
- [x] 1.3 Implement OAuth connection authorization lifecycle with protected state/PKCE, callback validation, secret storage, disconnect, and action-required status.
- [x] 1.4 Implement server-side OAuth refresh/expiry handling without exposing token artifacts.
- [x] 1.5 Update Admin UI to start/disconnect OAuth and render only safe lifecycle status.

## 2. Outbound and STDIO security

- [x] 2.1 Define configurable outbound MCP egress policy for schemes, ports, host/network allowlists, redirects, and local development profiles.
- [x] 2.2 Implement DNS/IP validation and connect-time revalidation that blocks prohibited private/link-local/loopback destinations outside explicit development policy.
- [x] 2.3 Apply outbound policy consistently to connection test, discovery, OAuth metadata, token exchange, and gateway invocation.
- [x] 2.4 Replace free-form STDIO execution with reviewed runner profiles, executable/argument/environment allowlists, isolation, and resource limits.

## 3. Gateway resilience and operations

- [x] 3.1 Add per-connection timeout, concurrency, rate-limit, idempotency-aware retry, and circuit-breaker policies around provider execution.
- [x] 3.2 Add safe health state transitions, manual diagnostic/test/reset controls, and action-required status for authorized operators.
- [x] 3.3 Emit redacted structured audit events, metrics, and traces with tenant/project/connection/tool/correlation identifiers.
- [x] 3.4 Add operator diagnostics views and deployment/incident-response documentation for credential rotation, provider outage, and blocked egress.

## 4. Security verification

- [x] 4.1 Add secret-leakage tests covering HTTP responses, State MCP tools, logs/audits, workflow definitions, and frontend state.
- [x] 4.2 Add OAuth lifecycle tests for callback validation, expiry, refresh failure, disconnect, and token redaction.
- [x] 4.3 Add egress tests for forbidden URLs, redirect bypass, DNS rebinding, local development allowance, and blocked private networks.
- [x] 4.4 Add STDIO profile, circuit-breaker, rate-limit, telemetry, and recovery tests.
- [x] 4.5 Run security-focused integration tests plus Go tests/vet, web tests/typecheck/lint, MCP smoke tests, and OpenSpec validation.
