## Why

Epic **#6 (Security & Ops)** Phase 3 adds rate limiting (PRD §83) to protect against brute-force and abuse. Public auth endpoints (login/register) and capability invocation currently have no rate limits. This change introduces a rate-limiter port, a token-bucket implementation, per-route config, and enforcement at login/register and capability invocation.

## What Changes

- **NEW** — Domain `RateLimiter` port (`internal/domain/services/rate_limiter.go`).
- **NEW** — Token-bucket implementation `internal/infrastructure/ratelimit/token_bucket.go` (built on `golang.org/x/time/rate`), safe for concurrent use.
- **NEW** — Scope-key helpers `internal/infrastructure/ratelimit/key.go` (tenant/user/API key/route).
- **NEW** — Rate-limit config `internal/infrastructure/config/ratelimit.go` (env-driven with defaults) wired into `Config`.
- **NEW** — Echo `RateLimit` middleware (`internal/interfaces/http/middleware/ratelimit.go`) with fail-open and 429 + Retry-After.
- **MODIFIED** — `RegisterAuthRoutes` applies rate limiting to `POST /login` (per email/IP) and `POST /register` (per IP).
- **MODIFIED** — `CapabilityInvoker` rate-limits invocation with a tenant+capability scope key and a classified `capability.rate_limited` error.
- **MODIFIED** — `CapabilityService.TestInvoke` injects the rate limiter into the invoker.
- **MODIFIED** — Error handler maps `RATE_LIMITED` → 429 + Retry-After.
- **MODIFIED** — OpenAPI documents 429 on `/capabilities/{id}/test`.
- Uses **`api-feature`** skill.

## Capabilities

### New Capabilities

- `backend/rate-limit`: `RateLimiter` port + token-bucket implementation + scope keys + config.
- `auth/login-rate-limit`: rate limiting on login/register (brute-force protection, fail-open).
- `capability/invoke-rate-limit`: rate limiting in the capability security chain + classified error.

## Impact

- **`apps/api/internal/domain/services/`** — add `rate_limiter.go`.
- **`apps/api/internal/infrastructure/ratelimit/`** — new package (`token_bucket.go`, `key.go`).
- **`apps/api/internal/infrastructure/config/`** — add `ratelimit.go`, modify `env.go`.
- **`apps/api/internal/interfaces/http/middleware/`** — add `ratelimit.go`, modify `error_handler.go`.
- **`apps/api/internal/interfaces/http/routes/`** — apply rate limits to auth routes.
- **`apps/api/internal/domain/capability/`** — modify `invoker.go`, `errors.go`.
- **`apps/api/internal/application/services/capability_service.go`** — inject limiter.
- **`apps/api/cmd/server/main.go`** — build and inject limiters.
- **`apps/api/go.mod`** — `golang.org/x/time` promoted to direct dependency.
- **No** DB migration, worker, or shared-package changes.

## Non-Goals

- Redis/shared rate-limit store — in-memory only (documented limitation, resets on restart).
- Global per-route rate limiting for every endpoint — focused on auth + capability.
- Structured logging for limiter failures — deferred to Phase 4.

## Dependencies

- Phase 1 `rbac-tenant-permissions` (RequirePermission middleware coexist).
- Existing auth and capability services.
