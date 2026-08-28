## 1. OIDC provider port + adapter (Skill: api-feature)

- [x] 1.1 Add OIDC/OAuth2 deps (`coreos/go-oidc`, `golang.org/x/oauth2`)
- [x] 1.2 Create `domain/services/oidc_provider.go` (`OIDCProvider` port + `OIDCIdentity`)
- [x] 1.3 Create `infrastructure/oidc/provider.go` (discovery, JWKS, PKCE S256, ID-token verify)
- [x] 1.4 Add SSO config (`config/sso.go`, `SSO_*` env)
- [x] 1.5 Add PKCE challenge test

## 2. user_identities store (Skill: db-sqlc-schema)

- [x] 2.1 Add migration `00009_user_identities.sql`
- [x] 2.2 Add `db/queries/identity.sql`; run `sqlc generate`
- [x] 2.3 Create entity `UserIdentity` + `IUserIdentityRepository`
- [x] 2.4 Create pgx adapter + compose in `PostgresAdapter`

## 3. SSO flow + account linking (Skill: api-feature)

- [x] 3.1 Create `application/services/sso_service.go` (`StartAuth`, `CompleteLogin`, auto-provision)
- [x] 3.2 Create `interfaces/http/controllers/sso_controller.go` (cookies, redirect, callback)
- [x] 3.3 Register SSO routes (`/sso/providers`, `/sso/:provider`, `/sso/:provider/callback`)
- [x] 3.4 Wire providers + service + controller in `cmd/server/main.go`
- [x] 3.5 Add `auth.sso_login` audit action

## 4. Verify

- [x] 4.1 `go build ./...` + `go vet ./...` pass
- [x] 4.2 `go test ./...` passes
- [x] 4.3 `go mod tidy` clean
- [x] 4.4 `gofmt` clean
