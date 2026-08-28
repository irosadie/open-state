## Context

Epic #6 Phase 5 adds SSO via OIDC (PRD §79). The platform authenticates with email/password today; SSO adds external identity login with account linking and auto-provisioning.

## Goals / Non-Goals

**Goals:**
- OIDC provider port + adapter (Google/GitHub/Entra).
- Authorization Code + PKCE flow.
- `user_identities` store + auto-provision + account linking.
- SSO session issuance (same JWT as local login) + audit.

**Non-Goals:**
- SSO frontend UI.
- Per-user tenant selection (defaults to demo tenant).
- SCIM.

## Decisions

### D1: OIDC provider port
`domain/services/oidc_provider.go` defines `OIDCProvider` (`AuthURL`, `Exchange`, `VerifyIDToken`) and `OIDCIdentity` (Provider, Subject, Email, Name, Photo). `infrastructure/oidc.Provider` implements it using `coreos/go-oidc` (discovery + JWKS) and `golang.org/x/oauth2`.

### D2: PKCE S256
`AuthURL` computes `code_challenge = base64url(SHA256(code_verifier))` (RFC 7636). The controller stores `state` + `code_verifier` in HttpOnly cookies; the callback verifies `state` (CSRF) and passes the verifier to `Exchange`.

### D3: user_identities
Migration `00009` creates `user_identities(user_id, provider, subject_id, auto_provisioned)` with a unique `(provider, subject_id)`. `IUserIdentityRepository` exposes find/create/list. `user_id` references `users(id)`; a user may hold one identity per provider.

### D4: Auto-provision + linking
`SSOService.resolveOrProvision` looks up `(provider, subject)`; if found, returns the linked user; else creates a local user (email, name, VIEWER role via `role_assignments`), links the identity (`auto_provisioned=true`), and issues the same JWT as local login. SSO events are audited (`auth.sso_login`).

### D5: Flow endpoints
`SSOController`:
- `GET /api/auth/sso/providers` → enabled providers.
- `GET /api/auth/sso/:provider` → build auth URL (state+verifier cookies) and redirect.
- `GET /api/auth/sso/:provider/callback` → verify state, exchange code, redirect frontend with `?sso=success&token=...` or `?sso=error&reason=...`.
