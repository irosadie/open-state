## Context

Phases 3–6 introduce project MCP connections, catalog discovery, validated bindings,
and gateway execution. They make provider calls possible; this phase makes them safe
and operable for production deployments and open-source self-hosting.

## Goals / Non-Goals

**Goals:**

- Keep credential, OAuth, network, STDIO, and telemetry boundaries explicit.
- Let an operator diagnose a failed provider without seeing secrets or raw sensitive
  provider content.
- Bound external dependency failures so they do not cascade through the state engine.

**Non-Goals:**

- Becoming an OAuth identity provider, promising external provider uptime, or allowing
  arbitrary shell execution in a hosted service.

## Decisions

### D1 — Pluggable secret store with protected references

Define a secret-store port for put/resolve/rotate/revoke semantics. The development
implementation can resolve environment-backed references; production deployments bind
the same port to their vault/KMS mechanism. Business records keep only opaque IDs and
safe lifecycle status. OAuth access/refresh artifacts use the same store.

### D2 — OAuth authorization is a connection lifecycle, not a form field

The operator begins OAuth authorization from the connection page. The platform creates
a short-lived, PKCE/state-protected authorization transaction, handles the callback,
stores resulting artifacts through the secret store, and exposes only `connected`,
`expired`, `disconnected`, or `action_required`. Refresh happens server-side before a
gateway call when safe; refresh failure changes status without revealing details.

### D3 — Egress policy is enforced on every resolution and request

Build an outbound policy component used by testing, discovery, OAuth metadata access,
and gateway calls. It validates scheme/port/host allowlists, resolves DNS, rejects
prohibited addresses including private/link-local/loopback except explicit local-dev
profiles, revalidates at connect time, and blocks redirects that violate policy.

### D4 — STDIO is self-hosted runner-only

STDIO execution uses named reviewed profiles containing executable, allowed arguments,
environment allowlist, working directory policy, CPU/memory/time limits, and process
isolation. The web form selects a profile; it never accepts a free-form command in a
hosted deployment.

### D5 — Centralize gateway resilience and redacted telemetry

Apply timeout, concurrency, idempotency-aware retry, circuit-breaker, rate-limit, and
tracing wrappers around the provider execution port. Emit tenant/project/connection/
tool/correlation outcome metrics and audits after a redaction boundary. Payload capture
is opt-in and schema-aware, never raw by default.

## Risks / Trade-offs

- **[Strict SSRF protection breaks local demos]** → ship an explicit development
  profile that permits approved localhost endpoints only; production defaults deny.
- **[OAuth provider variance]** → implement standards-based HTTP OAuth flow plus
  connection-specific capability flags; unsupported providers are safely
  `action_required`, not partially connected.
- **[Circuit breaking hides intermittent recovery]** → expose state and manual test/
  reset diagnostics, with controlled half-open recovery.

## Migration Plan

1. Introduce secret-store and egress-policy ports with safe no-op/migration adapters.
2. Move existing provider credential references through the secret-store boundary.
3. Add OAuth lifecycle endpoints/callbacks and UI status without exposing artifacts.
4. Enforce egress policy for all outbound MCP/OAuth flows; enable strict production
   default after documented migration checks.
5. Add resilience/telemetry wrappers and dashboards/diagnostics.
6. Roll back per connection by disabling gateway mode or the connection; credential
   references remain intact and no plaintext migration is required.
