## 1. Rate limiter port + implementation (Skill: api-feature)

- [x] 1.1 Create `internal/domain/services/rate_limiter.go` (`RateLimiter` port)
- [x] 1.2 Create `internal/infrastructure/ratelimit/token_bucket.go` (token bucket, `golang.org/x/time/rate`, mutex)
- [x] 1.3 Create `internal/infrastructure/ratelimit/key.go` (scope-key helpers)
- [x] 1.4 Promote `golang.org/x/time` to direct dependency in `go.mod`

## 2. Rate-limit config (Skill: api-feature)

- [x] 2.1 Create `internal/infrastructure/config/ratelimit.go` (env-driven with defaults)
- [x] 2.2 Add `RateLimit` to `Config` and load in `env.go`

## 3. HTTP middleware + auth wiring (Skill: api-feature)

- [x] 3.1 Create `internal/interfaces/http/middleware/ratelimit.go` (`RateLimit`, `LoginKey`, `RegisterKey`)
- [x] 3.2 Fail-open on limiter error; 429 + Retry-After on denial
- [x] 3.3 Apply to `POST /login` and `POST /register` in `routes.go`
- [x] 3.4 Update `CreateApp` + `main.go` to build/inject limiters

## 4. Capability invoke rate limiting (Skill: api-feature)

- [x] 4.1 `CapabilityInvoker` scope key `tenant:capability` + `ErrorKindRateLimited` (`capability.rate_limited`)
- [x] 4.2 Add `ErrorKindRateLimited` to `domain/capability/errors.go`
- [x] 4.3 `CapabilityService.TestInvoke` injects limiter into invoker
- [x] 4.4 Error handler maps `RATE_LIMITED` → 429 + Retry-After

## 5. Docs (Skill: docs-openapi)

- [x] 5.1 Add 429 response to `/capabilities/{id}/test` in `docs/openapi.json`

## 6. Verify

- [x] 6.1 `go build ./...` + `go vet ./...` pass
- [x] 6.2 Tests pass (token bucket, middleware, invoker rate limit, error handler)
- [x] 6.3 `gofmt` clean on changed files
