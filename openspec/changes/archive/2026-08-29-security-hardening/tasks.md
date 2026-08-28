## 1. Security headers (Skill: api-feature)

- [x] 1.1 Create `middleware/security_headers.go` (nosniff, X-Frame-Options, Referrer-Policy, Permissions-Policy, CSP, HSTS)
- [x] 1.2 Add security config (`config/security.go`, `SECURITY_CSP`, `SECURITY_HSTS_*`)
- [x] 1.3 Apply in `cmd/server/main.go`

## 2. CORS hardening (Skill: api-feature)

- [x] 2.1 Create `middleware/cors.go` (allow-list, restricted methods/headers)
- [x] 2.2 Remove permissive `echomw.CORS()` from `CreateApp`
- [x] 2.3 Apply hardened CORS in `cmd/server/main.go` (`CORS_ALLOWED_ORIGINS`)

## 3. Docker hardening (Skill: ops-docker)

- [x] 3.1 Multi-stage Dockerfile (builder + alpine runner)
- [x] 3.2 CGO-disabled static build
- [x] 3.3 Non-root user (`appuser`)
- [x] 3.4 Healthcheck on `/health`

## 4. Secrets management (Skill: ops-docker)

- [x] 4.1 Enforce `JWT_SECRET` ≥ 32 chars (fail-fast in config)
- [x] 4.2 Confirm `.gitignore` excludes `.env*`
- [x] 4.3 Add trufflehog secret-scan step to `go-ci.yml`
- [x] 4.4 Document secrets + rotation in `SECURITY.md`

## 5. Verify

- [x] 5.1 `go build ./...` + `go vet ./...` pass
- [x] 5.2 Tests pass (security headers middleware)
- [x] 5.3 `gofmt` clean
