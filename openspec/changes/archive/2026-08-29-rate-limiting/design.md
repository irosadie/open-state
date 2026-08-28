## Context

Epic #6 Phase 3 adds rate limiting (PRD §83) to protect public auth endpoints and capability invocation from brute-force and abuse. No rate limits exist today.

## Goals / Non-Goals

**Goals:**
- `RateLimiter` domain port.
- In-memory token-bucket implementation (`golang.org/x/time/rate`).
- Scope-key helpers and per-operation config (env-driven).
- HTTP `RateLimit` middleware (fail-open, 429 + Retry-After).
- Rate limiting on login/register and capability invocation.

**Non-Goals:**
- Redis/shared store.
- Global rate limiting for every route.
- Structured logging (Phase 4).

## Decisions

### D1: Domain RateLimiter port
`internal/domain/services/rate_limiter.go` defines `RateLimiter { Allow(ctx, key) (bool, error) }`. The HTTP middleware and the capability invoker depend on the port, not a concrete implementation. The capability package keeps its own port of the same signature; `TokenBucket` satisfies both structurally.

### D2: Token-bucket implementation
`internal/infrastructure/ratelimit/token_bucket.go` uses `golang.org/x/time/rate` with one independent bucket per key, created lazily and guarded by a mutex. State is in-memory and resets on restart (documented limitation; Redis is future work).

### D3: Scope keys
`key.go` provides consistent `dimension:value` helpers: `TenantKey`, `UserKey`, `APIKey`, `RouteUserKey`, `RouteIPKey`, `TenantCapabilityKey`.

### D4: Per-operation config
`config/ratelimit.go` reads `RATE_LIMIT_LOGIN_*`, `RATE_LIMIT_REGISTER_*`, `RATE_LIMIT_CAPABILITY_*` (rate + burst) with safe defaults, exposed on `Config.RateLimit`.

### D5: HTTP RateLimit middleware
`middleware/ratelimit.go` builds a scope key per request (`LoginKey` from email, fallback IP; `RegisterKey` from IP), calls `limiter.Allow`, fails open on limiter error, and returns 429 + Retry-After on denial. `LoginKey` restores the request body so the controller can still bind it.

### D6: Capability invoke rate limiting
`CapabilityInvoker` scopes by `tenant:<id>:capability:<id>`, short-circuits before provider invocation on denial, and returns a classified `ErrorKindRateLimited` (`capability.rate_limited`). The error handler maps it to 429 + Retry-After. `CapabilityService.TestInvoke` injects the limiter.
