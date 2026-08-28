# backend/rate-limit Specification

## Purpose
Define the rate-limiter port and a token-bucket implementation used to protect
the API from brute-force and abuse (PRD §83, §62). It provides a domain port,
an in-memory token-bucket implementation, tenant/user/API-key scope keys, and
per-route configuration so rate limits can be applied at login, register, and
capability invocation.

## Requirements

### Requirement: Rate limiter port

The platform SHALL expose a rate-limiter port in the domain layer so the
application/HTTP layers depend on an interface, not a concrete implementation.

- The port SHALL be a Go interface (e.g. `RateLimiter`) with an
  `Allow(ctx, key string) (bool, error)` method.
- `key` SHALL encode the scope and dimension being limited (e.g.
  `tenant:<id>`, `user:<id>`, `apikey:<id>`, `route:login:user:<id>`).
- `Allow` returns `true` when the request is permitted and `false` when the
  limit is exceeded for the key.
- The HTTP layer SHALL map a `false` result to a `429 Too Many Requests`
  response.

#### Scenario: Request within limit is allowed

- **WHEN** a key's usage is below its configured rate
- **THEN** `Allow` returns `true`

#### Scenario: Request over limit is rejected

- **WHEN** a key's usage exceeds its configured rate
- **THEN** `Allow` returns `false`, and the HTTP layer returns 429

### Requirement: Token-bucket implementation

The platform SHALL provide a token-bucket rate limiter as the default
implementation (PRD §83).

- The implementation SHALL be built on `golang.org/x/time/rate.Limiter` (token
  bucket), keyed per scope.
- Each key SHALL have an independent bucket with a configured rate (tokens per
  second) and burst size.
- An in-memory store SHALL hold buckets and be safe for concurrent access.
- The implementation SHALL satisfy the `RateLimiter` port.

#### Scenario: Burst then sustained rate

- **WHEN** a key makes a burst of requests up to its burst size
- **THEN** they are allowed, and subsequent requests are limited to the
  sustained rate

#### Scenario: Independent buckets

- **WHEN** two different keys are limited
- **THEN** their buckets SHALL not interfere with one another

### Requirement: Rate limit configuration

The platform SHALL configure rate limits per route/operation (PRD §83).

- A `RateLimitConfig` SHALL define a `rate` (tokens/second) and `burst` for a
  named operation.
- Default limits SHALL be defined for: login, register, and capability invoke.
- Limits SHALL be configurable via environment variables (e.g.
  `RATE_LIMIT_LOGIN_RATE`, `RATE_LIMIT_LOGIN_BURST`) with sane defaults.

#### Scenario: Configurable limits

- **WHEN** an operator sets a rate-limit env var
- **THEN** the matching operation uses that rate/burst
- **WHEN** no env var is set
- **THEN** a documented default applies

### Requirement: Scope keys

The platform SHALL scope rate limits by tenant, user, and/or API key where
relevant (PRD §83).

- Login/register SHALL be scoped by user identity (email/IP) to protect against
  brute-force per account.
- Capability invocation SHALL be scoped by tenant and capability.
- The key format SHALL be documented and consistent (e.g. `dimension:value`).

#### Scenario: Brute-force protection per account

- **WHEN** repeated failed login attempts target the same account
- **THEN** that account's key is limited while other accounts remain unaffected

### Requirement: Rate limit exceeded error

The platform SHALL return a consistent, machine-readable 429 response when a
limit is exceeded.

- The error SHALL use HTTP status 429.
- The response SHALL include a stable code (e.g. `rate_limit_exceeded`).
- The `Retry-After` header SHALL indicate when the client may retry.

#### Scenario: Client receives 429 with retry hint

- **WHEN** a request exceeds a rate limit
- **THEN** the response is 429 with a `Retry-After` header

### Requirement: No bypass across restart

Rate-limit state is in-memory by default (acceptable for a single instance).

- The spec SHALL document that an in-memory store resets on process restart and
  that a shared store (Redis) is a future enhancement.
- This limitation SHALL NOT block the primary brute-force/abuse protection.

#### Scenario: Documented in-memory behavior

- **WHEN** the process restarts
- **THEN** rate-limit counters reset (documented limitation, not a bug)
