# auth/oidc-flow Specification

## Purpose
Define the SSO login flow using the OIDC Authorization Code flow with PKCE (PRD
§79). It establishes the endpoints for initiating a provider login, handling the
provider redirect, and exchanging the authorization code for tokens and user
profile.

## Requirements

### Requirement: Initiate SSO login

The platform SHALL expose an endpoint to start an SSO login for a provider.

- `GET /api/auth/sso/:provider` SHALL generate the authorization URL and
  redirect the user to the provider (or return the URL).
- The flow SHALL generate a PKCE code verifier + challenge and a `state` value.
- The `state` and PKCE verifier SHALL be stored (e.g. in a signed cookie or
  short-lived session) for the callback to verify.

#### Scenario: Redirect to provider

- **WHEN** a user requests SSO login for a provider
- **THEN** the platform redirects to the provider's authorization URL with
  `code_challenge`, `code_challenge_method=S256`, and `state`

#### Scenario: State protects against CSRF

- **WHEN** the callback returns
- **THEN** the platform SHALL verify the returned `state` matches the stored
  value

### Requirement: OIDC callback

The platform SHALL handle the provider redirect on a callback endpoint.

- `GET /api/auth/sso/:provider/callback` SHALL receive `code` and `state`.
- The platform SHALL verify `state`, then exchange the `code` for tokens using
  PKCE (`code_verifier`).
- The platform SHALL validate the ID token (issuer, audience, expiry,
  signature).

#### Scenario: Successful callback

- **WHEN** a valid `code` and matching `state` are provided
- **THEN** the platform exchanges the code, validates the ID token, and retrieves
  the user profile

#### Scenario: Invalid state rejected

- **WHEN** `state` does not match
- **THEN** the platform rejects the callback (CSRF protection)

### Requirement: User profile extraction

The platform SHALL extract a normalized user profile from the OIDC
claims/userinfo.

- Normalized fields SHALL include: `sub` (provider subject id), `email`, `name`,
  and optionally `photo`.
- The provider + subject id SHALL form the unique external identity.

#### Scenario: Profile normalized

- **WHEN** userinfo/claims are retrieved
- **THEN** they are normalized into a canonical profile (sub, email, name, photo)

### Requirement: Token verification

The platform SHALL verify ID tokens before trusting them.

- Verification SHALL validate signature (JWKS), `iss` (issuer), `aud`
  (audience), and `exp` (expiry).
- Nonce/`azp` SHALL be validated where the provider enforces it.

#### Scenario: Forged/expired token rejected

- **WHEN** an ID token fails signature, issuer, audience, or expiry validation
- **THEN** the flow is rejected with an authentication error

### Requirement: Errors

SSO failures SHALL produce clear, safe error responses.

- Callback errors SHALL redirect to a frontend error page with a query param, or
  return a JSON error for API clients.
- No raw provider tokens or secrets SHALL be exposed in error responses.

#### Scenario: Safe error handling

- **WHEN** an SSO flow fails
- **THEN** the user receives a clear message and no provider secrets are leaked
