## Context

Epic #6 Phase 6 hardens the API (PRD §139, §74). Echo's default CORS allows all origins and the Dockerfile ran as root.

## Goals / Non-Goals

**Goals:**
- Security headers (CSP, HSTS) middleware.
- Hardened CORS (allow-list).
- Non-root, multi-stage Docker image with healthcheck.
- Secrets management: env-only secrets, JWT strength, CI scan, rotation docs.

**Non-Goals:**
- TLS termination (reverse proxy).
- WAF/network segmentation.

## Decisions

### D1: Security headers middleware
`middleware.SecurityHeaders(cfg)` sets `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `Referrer-Policy`, `Permissions-Policy`, `X-XSS-Protection: 0`, `Content-Security-Policy` (default `default-src 'self'`), and HSTS when `EnableHSTS` (default off for local HTTP dev). Applied in `cmd/server/main.go`.

### D2: Hardened CORS
`middleware.CORS(cfg)` uses Echo's `CORSWithConfig` with an explicit origin allow-list, restricted methods (GET/POST/PUT/PATCH/DELETE/OPTIONS) and headers (Authorization, Content-Type, X-Tenant-ID, X-Project-ID). Credentials are allowed only when an allow-list is set. An empty allow-list denies all cross-origin requests (same-origin still works). Config via `CORS_ALLOWED_ORIGINS`.

### D3: Docker hardening
The Dockerfile uses a `golang:1.26-alpine` builder producing a `CGO_ENABLED=0` static binary, then a minimal `alpine:3.20` runtime with CA certs/tzdata and a non-root `appuser`. A `HEALTHCHECK` calls `/health`. No builder tooling or secrets are present in the runtime stage.

### D4: Secrets management
Secrets are injected at runtime via env (never baked into images or committed). `JWT_SECRET` must be ≥ 32 chars (fail-fast). `.gitignore` excludes `.env*`. CI runs `trufflehog` to scan for leaked secrets. `SECURITY.md` documents required secrets and rotation.
