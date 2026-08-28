# auth/account-linking Specification

## Purpose

Define how an external OIDC identity is associated with a platform account:
either linking to an existing local account or auto-provisioning a new one (PRD
§79). It establishes the identity store, first-login behavior, and session
issuance after SSO.

## ADDED Requirements

### Requirement: External identity store

The platform SHALL persist the association between an external provider identity
and a local user.

- A `user_identities` table SHALL store `provider`, `subject_id`, `user_id`, and
  timestamps.
- A unique constraint SHALL prevent the same provider subject from mapping to
  more than one user.
- A user MAY have multiple external identities (one per provider).

#### Scenario: Identity recorded on first login

- **WHEN** a user authenticates via a provider
- **THEN** a `user_identities` row linking `(provider, subject)` to the local
  user is created if absent

#### Scenario: Duplicate identity prevented

- **WHEN** a provider subject already maps to a user
- **THEN** no duplicate identity is created

### Requirement: Auto-provision new users

The platform SHALL create a local user automatically when an external identity
is seen for the first time (PRD §79).

- On first login, a new user SHALL be created from the normalized profile with
  the least-privilege role (VIEWER, see Phase 1 RBAC).
- The user SHALL be linked to the external identity.
- A default tenant role assignment SHALL be created so the new user is
  authorized (default VIEWER).

#### Scenario: First SSO login provisions a user

- **WHEN** a user with no existing account logs in via SSO
- **THEN** a local user is created, linked to the identity, and assigned the
  default tenant role

### Requirement: Link to existing account

The platform SHALL support linking an external identity to an existing local
account.

- When a user is already authenticated and chooses to link a provider identity,
  the platform SHALL associate the identity with the current user.
- Linking SHALL require the user to be authenticated.

#### Scenario: Link provider identity to current user

- **WHEN** an authenticated user links a new provider identity
- **THEN** the identity is associated with that user

### Requirement: Resolve identity to user

The platform SHALL resolve an external identity to a local user on login.

- On SSO callback, the platform SHALL look up the user by `(provider, subject)`.
- If found, login proceeds for that user; if not, auto-provision (or prompt to
  link).

#### Scenario: Existing identity logs in

- **WHEN** an external identity already maps to a user
- **THEN** the platform authenticates as that user

### Requirement: Session and token issuance

After successful SSO, the platform SHALL issue the same JWT/session used for
local login.

- The existing `LoginUserUseCase`/token issuance SHALL be reused so SSO and local
  logins share the same session semantics.
- The returned user SHALL carry role and permissions for the tenant (Phase 1).

#### Scenario: SSO login returns a session

- **WHEN** SSO authentication succeeds
- **THEN** an access token and user (with tenant role/permissions) are returned
  as with a normal login

### Requirement: Audit SSO events

SSO login and account-linking SHALL be auditable (PRD §50).

- Successful SSO login SHALL append an audit action (e.g. `auth.sso_login`).
- Account provisioning/linking SHALL be auditable.

#### Scenario: SSO login audited

- **WHEN** a user logs in via SSO
- **THEN** an audit entry records the event with the user actor
