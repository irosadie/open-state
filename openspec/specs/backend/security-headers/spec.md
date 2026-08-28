# backend/security-headers Specification

## Purpose
Define HTTP security headers for the API (and reverse-proxied web responses)
including Content Security Policy (CSP) and HTTP Strict Transport Security
(HSTS), reducing XSS, clickjacking, and MITM risk (PRD §139, §84). It
establishes a security-headers middleware with safe defaults.

## Requirements

### Requirement: Security headers middleware

The platform SHALL add a security-headers middleware that sets recommended
headers on every HTTP response.

- Headers SHALL include at minimum: `X-Content-Type-Options: nosniff`,
  `X-Frame-Options: DENY` (or `SAMEORIGIN`), `Referrer-Policy`,
  `Permissions-Policy`, and `X-XSS-Protection: 0`.
- The middleware SHALL be applied globally in `create_app.go`.

#### Scenario: Default headers set

- **WHEN** any API response is returned
- **THEN** the recommended security headers are present

### Requirement: HSTS

The platform SHALL set HTTP Strict Transport Security when served over HTTPS.

- `Strict-Transport-Security: max-age=<seconds>; includeSubDomains` SHALL be set
  for HTTPS responses.
- HSTS SHALL be configurable (max-age and enable/disable) and SHALL NOT be forced
  in plain-HTTP local dev by default.

#### Scenario: HSTS on HTTPS

- **WHEN** the service is served over HTTPS
- **THEN** the HSTS header is present with a configured max-age

### Requirement: Content Security Policy

The platform SHALL set a Content-Security-Policy header.

- A default CSP SHALL be defined (e.g. `default-src 'self'`) appropriate to the
  served content.
- For the API (JSON only), a restrictive CSP SHALL be set; for the web frontend,
  a CSP permitting the app's scripts/styles SHALL be documented.

#### Scenario: CSP present

- **WHEN** a response is served
- **THEN** a `Content-Security-Policy` header is set per the applicable policy

### Requirement: Configurability

Security headers SHALL be configurable so operators can tune them.

- Env vars SHALL control HSTS max-age/enable and CSP string.
- Safe defaults SHALL apply when env vars are absent.

#### Scenario: Configurable headers

- **WHEN** an operator sets header env vars
- **THEN** the middleware uses those values
- **WHEN** none are set
- **THEN** safe defaults apply

### Requirement: Do not break API clients

Security headers SHALL not interfere with API clients or MCP/SSE connections.

- CORS and security headers SHALL coexist (see `backend/cors-hardening`).
- SSE/WebSocket responses SHALL still work with the headers applied.

#### Scenario: API and MCP still function

- **WHEN** security headers are enabled
- **THEN** existing API and MCP/SSE endpoints continue to work
