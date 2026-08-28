## Why

Epic **#6 (Security & Ops)** Phase 6 hardens the platform (PRD §139, §74): security headers, hardened CORS, a non-root hardened Docker image, and secrets management. The API used Echo's permissive default CORS and the Dockerfile ran as root.

## What Changes

- **NEW** — `middleware.SecurityHeaders` (nosniff, X-Frame-Options, Referrer-Policy, Permissions-Policy, CSP, configurable HSTS).
- **NEW** — `middleware.CORS` hardened middleware (explicit origin allow-list, restricted methods/headers, credentials only with allow-list).
- **MODIFIED** — `CreateApp` removed the permissive default CORS; `cmd/server/main.go` applies hardened CORS + security headers.
- **NEW** — Security config (`config/security.go`, `SECURITY_CSP`, `SECURITY_HSTS_*`, `CORS_ALLOWED_ORIGINS`).
- **MODIFIED** — `apps/api/Dockerfile`: multi-stage, CGO-disabled static binary, non-root user, healthcheck, no build tooling in runtime.
- **MODIFIED** — `JWT_SECRET` minimum length enforced (fail-fast) at startup.
- **NEW** — CI secret scan (trufflehog) in `go-ci.yml`.
- **MODIFIED** — `SECURITY.md` secrets management + rotation docs.
- Uses **`ops-docker`**, **`api-feature`** skills.

## Capabilities

### New Capabilities

- `backend/security-headers`: security-headers middleware (CSP, HSTS).
- `backend/cors-hardening`: allow-list CORS middleware.
- `ops/docker-hardening`: non-root, multi-stage, healthcheck Dockerfile.
- `ops/secrets-management`: env secrets, JWT strength, CI scan, rotation docs.

## Impact

- **`apps/api/internal/interfaces/http/middleware/`** — add `security_headers.go`, `cors.go`.
- **`apps/api/internal/interfaces/http/create_app.go`** — remove permissive CORS.
- **`apps/api/internal/infrastructure/config/`** — add `security.go`.
- **`apps/api/cmd/server/main.go`** — apply hardened CORS + security headers.
- **`apps/api/Dockerfile`** — hardened.
- **`.github/workflows/go-ci.yml`** — secret scan step.
- **`SECURITY.md`** — secrets + rotation docs.
- **No** DB/worker/shared changes.

## Non-Goals

- TLS termination — handled by reverse proxy/ingress (headers assume HTTPS for HSTS).
- Full network segmentation / WAF — external.

## Dependencies

- None beyond the existing API server.
