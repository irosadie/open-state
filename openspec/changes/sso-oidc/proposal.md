## Why

Epic **#6 (Security & Ops)** Phase 5 adds SSO/OIDC login (PRD §79). Users currently authenticate only with email/password. This change adds an OIDC provider port, the Authorization Code + PKCE flow, and external-identity account linking with auto-provisioning.

## What Changes

- **NEW** — Domain `OIDCProvider` port + `OIDCIdentity` (`domain/services/oidc_provider.go`).
- **NEW** — `infrastructure/oidc.Provider` adapter (Google/GitHub/Entra) with discovery + JWKS, code exchange + PKCE (S256), and ID-token verification.
- **NEW** — SSO config (`config/sso.go`, `SSO_GOOGLE_*`, `SSO_GITHUB_*`, `SSO_ENTRA_*`).
- **NEW** — `user_identities` table (migration 00009) + `IUserIdentityRepository` + pgx adapter.
- **NEW** — `SSOService`: `StartAuth` (state + PKCE) and `CompleteLogin` (exchange, resolve/auto-provision, issue session).
- **NEW** — `SSOController` + routes: `GET /api/auth/sso/providers`, `/sso/:provider`, `/sso/:provider/callback`.
- **NEW** — `auth.sso_login` audit action.
- Uses **`api-feature`**, **`db-sqlc-schema`** skills.

## Capabilities

### New Capabilities

- `auth/oidc-provider`: OIDC provider port + adapter (discovery/JWKS, PKCE, ID-token verify).
- `auth/oidc-flow`: Authorization Code + PKCE endpoints (init + callback).
- `auth/account-linking`: `user_identities`, auto-provision (VIEWER default), session issuance, SSO audit.

## Impact

- **`apps/api/db/migrations/`** — add `00009_user_identities.sql`.
- **`apps/api/db/queries/`** — add `identity.sql`; regenerate sqlc.
- **`apps/api/internal/domain/services/`** — add `oidc_provider.go`.
- **`apps/api/internal/domain/repositories/`** — add `user_identity_repository.go`.
- **`apps/api/internal/domain/entities/`** — add `user_identity.go`, extend audit actions.
- **`apps/api/internal/infrastructure/oidc/`** — new package.
- **`apps/api/internal/infrastructure/database/`** — add pgx identity repo; compose in adapter.
- **`apps/api/internal/application/services/sso_service.go`** — new.
- **`apps/api/internal/interfaces/http/controllers/sso_controller.go`** — new.
- **`apps/api/internal/interfaces/http/routes/`** — register SSO routes.
- **`apps/api/go.mod`** — OIDC/OAuth2 deps.

## Non-Goals

- SSO frontend UI — backend flow + endpoints delivered here.
- Per-user tenant selection — auto-provisioned users default to the demo tenant.
- SCIM provisioning — explicitly future (PRD §79).

## Dependencies

- Phase 1 `rbac-tenant-permissions` (default VIEWER role, audit).
- Phase 2 audit (`auth.sso_login`).
