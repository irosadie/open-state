# ops/secrets-management Specification

## Purpose

Define how secrets (JWT secret, database credentials, OIDC client secrets,
provider credentials) are managed so they are never baked into images or
committed to source, and are injected at runtime (PRD §84, §139).

## ADDED Requirements

### Requirement: Secrets via environment

The platform SHALL read all secrets from environment variables at runtime.

- Secrets SHALL NOT be hardcoded in source or committed to the repository.
- Required secrets SHALL be: `DATABASE_URL`, `JWT_SECRET`, OIDC client secrets
  (`SSO_*_CLIENT_SECRET`), and capability provider credentials.

#### Scenario: Secrets sourced from env

- **WHEN** the API starts
- **THEN** it reads secrets from environment variables, failing fast if required
  secrets are missing

### Requirement: No secrets in images

The platform SHALL NOT bake secrets into container images.

- Secrets SHALL be passed via container runtime (env, secrets mount, or
  orchestration platform).
- The Dockerfile SHALL NOT `COPY`/`ENV` any secret values.

#### Scenario: Image contains no secrets

- **WHEN** a built image is inspected
- **THEN** it contains no secret values (only references via env at runtime)

### Requirement: .gitignore and scanning

The platform SHALL keep secrets out of version control.

- Secret-like files (`.env`, `.env.*`, credentials) SHALL be in `.gitignore`.
- Example/env-template files (`*.env.example`) SHALL contain placeholders, not
  real values.
- A secret-scanning step SHALL be part of CI to block accidental commits.

#### Scenario: Env files ignored

- **WHEN** a developer creates a local `.env`
- **THEN** it is ignored by git and never committed

#### Scenario: CI scans for secrets

- **WHEN** a PR is opened
- **THEN** a secret scanner runs and flags any leaked secrets

### Requirement: Capability credential handling

The platform SHALL keep capability provider credentials as references, never as
plaintext in API responses (PRD §61, §91).

- The capability registry SHALL store a `credential_reference` (a key/vault path)
  rather than the secret value.
- Resolving the actual secret SHALL happen in infrastructure, scoped and
  access-controlled.

#### Scenario: Registry stores references only

- **WHEN** a capability is created/read
- **THEN** only the credential reference is returned, never the secret value

### Requirement: JWT secret strength

The `JWT_SECRET` SHALL be required and strong.

- The server SHALL fail fast if `JWT_SECRET` is missing or too short.
- Operators SHALL be advised to use a cryptographically random secret.

#### Scenario: Weak secret rejected

- **WHEN** `JWT_SECRET` is missing or below a minimum length
- **THEN** the API fails to start with a clear error

### Requirement: Rotation support

The platform SHALL document how secrets are rotated.

- Rotation SHALL be achievable by updating env values and restarting.
- The documentation SHALL describe the procedure for JWT, DB, and OIDC secrets.

#### Scenario: Documented rotation

- **WHEN** an operator rotates a secret
- **THEN** the documented procedure is followed and the service restarts with the
  new value
