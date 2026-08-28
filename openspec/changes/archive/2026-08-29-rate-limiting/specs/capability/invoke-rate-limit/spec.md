# capability/invoke-rate-limit Specification

## Purpose

Define how rate limiting is applied to capability invocation (PRD §62, §83) so a
single tenant, user, or capability cannot abuse the provider. It integrates the
rate-limiter port (see `backend/rate-limit`) into the capability security chain
(before invoke) and into the admin test-invoke path.

## ADDED Requirements

### Requirement: Rate limit capability invocation

The platform SHALL rate-limit capability invocation (PRD §62, §83).

- Invocations SHALL be scoped per tenant and per capability (e.g.
  `tenant:<id>:capability:<id>`).
- The security chain SHALL check the rate limit before invoking the provider
  (authorize state → validate input schema → **rate limit** → invoke).
- Exceeding the limit SHALL deny the invocation and be treated as
  `capability.denied` with a rate-limit reason.

#### Scenario: Invocation within limit proceeds

- **WHEN** a capability invocation is within its rate
- **THEN** the invocation proceeds to the provider

#### Scenario: Invocation over limit is denied

- **WHEN** a capability invocation exceeds its rate
- **THEN** it is denied, an audit entry (`capability.denied`) is recorded, and a
  rate-limit error is returned

### Requirement: Rate limiting in the capability invoker

The platform SHALL enforce rate limiting inside the `CapabilityInvoker` security
chain (PRD §62).

- The invoker SHALL accept a `RateLimiter` port.
- The invoker SHALL build the scope key from `TenantID` + `CapabilityID`.
- A `false` result from the limiter SHALL short-circuit before provider
  invocation.

#### Scenario: Short-circuit before invoke

- **WHEN** the rate limit is exceeded
- **THEN** the provider is NOT invoked and the denial is returned immediately

### Requirement: Rate limit test-invoke

The platform SHALL apply rate limiting to the admin sandbox test-invoke endpoint
(`POST /api/capabilities/:id/test`) (PRD §83).

- Test invocations SHALL be scoped per tenant + capability (or per user).
- A configurable limit SHALL prevent an operator from flooding the sandbox.

#### Scenario: Test invoke is limited

- **WHEN** an operator rapidly test-invokes a capability
- **THEN** further test invocations return 429 / a rate-limit denial

### Requirement: Separate limits by dimension

The platform SHALL support separate rate-limit dimensions where meaningful
(PRD §83).

- Capability invocation SHALL support a per-tenant limit and a per-capability
  limit independently.
- A request that exceeds EITHER dimension SHALL be denied.

#### Scenario: Per-tenant and per-capability limits combine

- **WHEN** a tenant exceeds its overall invoke limit OR a specific capability
  exceeds its own limit
- **THEN** the invocation is denied

### Requirement: Rate-limit error classification

A rate-limit denial SHALL be represented by a typed capability error with a
stable code (e.g. `capability.rate_limited`) and mapped to HTTP 429 where it
surfaces at the HTTP boundary.

- The error SHALL be distinct from authorization (`capability.unauthorized`) and
  validation failures.

#### Scenario: Rate-limit error is classified

- **WHEN** a capability invocation is rate-limited
- **THEN** the returned error has code `capability.rate_limited` (or equivalent)
  and maps to 429 at the HTTP boundary
