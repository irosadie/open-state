# auth/oidc-provider Specification

## Purpose

Define the OIDC provider abstraction so the platform can authenticate users via
external identity providers (Google, GitHub, Entra/Microsoft) (PRD §79). It
establishes a provider port, provider configuration, and the discovery of
provider metadata and signing keys.

## ADDED Requirements

### Requirement: OIDC provider port

The platform SHALL expose an OIDC provider port in the domain layer so
authentication depends on an interface, not a concrete provider.

- The port SHALL abstract: discovery, authorization-code exchange, token
  verification, and user-info retrieval.
- A concrete adapter SHALL exist for at least one provider (Google), with the
  interface designed to support GitHub and Entra/Microsoft.
- The port SHALL be implemented by an infrastructure adapter (e.g.
  `infrastructure/oidc`).

#### Scenario: Provider abstraction

- **WHEN** the auth flow needs to authenticate via an external provider
- **THEN** it SHALL use the provider port, independent of the concrete provider

### Requirement: Provider configuration

The platform SHALL configure OIDC providers via environment variables.

- Per-provider config SHALL include: `client_id`, `client_secret`,
  `issuer_url`/`discovery_url`, `redirect_uri`, and `scopes`.
- Env vars SHALL follow a consistent pattern (e.g. `SSO_GOOGLE_CLIENT_ID`,
  `SSO_GOOGLE_CLIENT_SECRET`, `SSO_GOOGLE_REDIRECT_URI`).
- Multiple providers SHALL be supported simultaneously; enabled providers SHALL
  be discoverable.

#### Scenario: Google provider configured

- **WHEN** `SSO_GOOGLE_*` env vars are set
- **THEN** the Google provider is enabled and available

#### Scenario: Provider disabled when unconfigured

- **WHEN** a provider's env vars are absent
- **THEN** that provider is not exposed as an option

### Requirement: Provider discovery and keys

The platform SHALL discover OIDC metadata and signing keys from the provider.

- The adapter SHALL fetch the OpenID configuration (authorization endpoint,
  token endpoint, JWKS URI) from the issuer.
- The adapter SHALL fetch and cache the JWKS keyset to verify ID tokens.

#### Scenario: Metadata discovered

- **WHEN** the provider is initialized
- **THEN** it resolves the authorization/token endpoints and JWKS URI from
  discovery

#### Scenario: Signing keys verified

- **WHEN** an ID token is verified
- **THEN** its signature is validated against the provider's JWKS keys

### Requirement: Supported providers

The platform SHALL support Google, GitHub, and Entra/Microsoft as first-class
providers (PRD §79).

- At minimum Google SHALL be fully implemented; GitHub and Entra SHALL be
  supported by the same port with provider-specific configuration.
- Provider selection SHALL be by a provider identifier (e.g. `google`, `github`,
  `entra`).

#### Scenario: Multiple providers coexist

- **WHEN** several providers are configured
- **THEN** the login UI lists each as a login option and the flow selects the
  correct adapter
