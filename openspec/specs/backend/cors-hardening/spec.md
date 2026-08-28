# backend/cors-hardening Specification

## Purpose
Define a hardened CORS configuration for the API that allows only trusted
origins and the required headers/methods, preventing cross-origin abuse while
preserving legitimate web-frontend access (PRD §74, §84).

## Requirements

### Requirement: Allow-list CORS origins

The platform SHALL restrict CORS to an explicit allow-list rather than
`*`.

- An allow-list of origins SHALL be configured (e.g. via `CORS_ALLOWED_ORIGINS`
  env var, comma-separated).
- Only origins in the allow-list SHALL receive CORS headers.
- Requests from disallowed origins SHALL NOT be granted CORS access.

#### Scenario: Allowed origin gets CORS headers

- **WHEN** a request originates from an allowed origin
- **THEN** the response includes the appropriate `Access-Control-Allow-Origin`

#### Scenario: Disallowed origin blocked

- **WHEN** a request originates from a non-allowed origin
- **THEN** it does NOT receive CORS authorization headers

### Requirement: Restrict methods and headers

The platform SHALL restrict CORS to the methods and headers the API actually
uses.

- Allowed methods SHALL be limited to GET, POST, PATCH, PUT, DELETE, OPTIONS.
- Allowed headers SHALL include Authorization, Content-Type, `X-Tenant-ID`, and
  `X-Project-ID`.
- Credentials (cookies) SHALL only be allowed when the origin allow-list is used
  and configured.

#### Scenario: Preflight for used headers

- **WHEN** a preflight (OPTIONS) request lists used headers
- **THEN** those headers are reflected as allowed

#### Scenario: Unknown headers rejected

- **WHEN** a preflight lists headers the API does not use
- **THEN** they are not allowed in the CORS response

### Requirement: Default deny in production

In production, CORS SHALL default to deny-all unless explicitly configured.

- Without `CORS_ALLOWED_ORIGINS`, no cross-origin access SHALL be granted (except
  same-origin).
- Local dev MAY default to allowing the dev origin for convenience.

#### Scenario: Production defaults deny

- **WHEN** running in production without a configured allow-list
- **THEN** cross-origin requests are not authorized

### Requirement: Coexist with security headers

CORS SHALL work together with the security-headers middleware (see
`backend/security-headers`).

- Both middlewares SHALL be applied without conflict.
- Security headers SHALL NOT be stripped when CORS headers are added.

#### Scenario: Headers combined

- **WHEN** an allowed cross-origin request is served
- **THEN** both CORS headers and security headers are present
