# auth/login-rate-limit Specification

## Purpose
Define how rate limiting is applied to the public authentication endpoints
(`/api/auth/login` and `/api/auth/register`) to protect against brute-force and
credential-stuffing abuse (PRD §83). It wires the rate-limiter port (see
`backend/rate-limit`) into the auth HTTP layer, scoped per identity.

## Requirements

### Requirement: Rate limit login

The platform SHALL rate-limit `POST /api/auth/login` (PRD §83).

- Each request SHALL be bucketed by the target account (email) and/or client
  IP.
- A configurable per-account rate (default, e.g. 10 requests/minute burst) SHALL
  apply.
- Exceeding the limit SHALL return 429 with a stable code and `Retry-After`.

#### Scenario: Repeated login attempts limited

- **WHEN** a client submits many login attempts for one account in a short window
- **THEN** the account key becomes limited and further attempts return 429

#### Scenario: Different accounts unaffected

- **WHEN** one account is rate-limited
- **THEN** other accounts SHALL still be able to attempt login

### Requirement: Rate limit register

The platform SHALL rate-limit `POST /api/auth/register` (PRD §83).

- Registration SHALL be bucketed by client IP (and/or device fingerprint) since
  there is no account identity yet.
- A configurable per-IP rate SHALL apply to prevent mass account creation.

#### Scenario: Mass registration limited

- **WHEN** a single IP attempts many registrations in a short window
- **THEN** further registrations from that IP return 429

### Requirement: Fail-open vs fail-closed

The platform SHALL define behavior when the rate limiter is unavailable.

- If the rate limiter returns an error, the platform SHALL **fail open** for
  authentication (allow the request) and log the failure, so a rate-limiter
  outage does not lock out legitimate users.
- This SHALL be documented and tested.

#### Scenario: Rate limiter error fails open

- **WHEN** the rate limiter errors during a login attempt
- **THEN** the login request proceeds (fail-open) and the error is logged

### Requirement: Configuration

Login and register rate limits SHALL be configurable via environment variables
(see `backend/rate-limit`).

- `RATE_LIMIT_LOGIN_RATE` / `RATE_LIMIT_LOGIN_BURST`
- `RATE_LIMIT_REGISTER_RATE` / `RATE_LIMIT_REGISTER_BURST`
- Sensible defaults SHALL be documented and applied when env vars are absent.

#### Scenario: Defaults applied

- **WHEN** no rate-limit env vars are set
- **THEN** documented default limits apply to login and register

### Requirement: Integration point

The rate limiter SHALL be applied as middleware on the auth route group (or
inside the auth controller) so it runs before credential validation.

- The limiter SHALL run after CORS but before the login/register handler.
- The scoping key SHALL be derived from the request (email body + IP), never
  from an unauthenticated caller-supplied trusted value alone.

#### Scenario: Limit checked before validation

- **WHEN** a login request arrives
- **THEN** the rate check runs before bcrypt/credential validation to avoid
  expensive work on limited requests
