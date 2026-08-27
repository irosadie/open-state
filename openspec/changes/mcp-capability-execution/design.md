## Context

Epic #4 (MCP & Integrasi) needs a safe, deterministic capability execution layer on top
of the capability/policy data already specified in `persistence-capabilities-policies`.
The core engine is MCP-agnostic (PRD §172, §2559); this slice provides the domain +
application layer only — no HTTP, no MCP SDK. See proposal.md for motivation.

## Goals / Non-Goals

**Goals:**
- Provide a `CapabilityProvider` port and a resolver/invoker that are fully unit-testable.
- Enforce the PRD §62 security chain and PRD §60 most-restrictive-wins binding resolution.
- Normalize results/failures and map them to deterministic capability events (PRD §63).
- Support retry/timeout (PRD §88, §160) and idempotency (PRD §64).
- Default to mock/sandbox execution (PRD §2064).

**Non-Goals:**
- Real MCP client/server (separate proposal `mcp-server-runtime`).
- HTTP controllers/routes (separate proposal `capability-admin-api`).
- LLM abstraction, secrets resolution, persistence (already covered elsewhere).

## Decisions

### D1. Package layout
```
apps/api/internal/domain/capability/
├── provider.go       // CapabilityProvider port + Invocation/Result types
├── resolver.go       // CapabilityResolver (binding-aware resolution)
├── invoker.go        // execution pipeline: security chain → invoke → normalize
├── errors.go         // CapabilityError classification (PRD §87)
├── retry.go          // backoff + jitter, retryable classification
├── idempotency.go    // idempotency key support
└── *_test.go
apps/api/internal/infrastructure/capability/
└── mock_provider.go  // sandbox/mock provider
```
The domain package has **no** infra/HTTP/MCP-SDK dependency (PRD §2559).

### D2. Provider port shape
`CapabilityProvider.Invoke(ctx, Invocation) (InvocationResult, error)`. Providers are
implementations of this port. MCP becomes one implementation later (D1 in
`mcp-server-runtime`). The resolver maps a logical capability to a provider by
`provider_type` + `provider_id` from the registry (PRD §59).

### D3. Binding resolution semantics
Resolution walks Global → Tenant → Workflow → State (PRD §60). A `DENY` at any more
specific level wins over `ALLOW`. Absence of an explicit binding falls through to the
less-specific scope. Most-restrictive-wins is evaluated by the resolver, not the repo.

### D4. Security chain order
Fixed order per PRD §62: authenticate → authorize tenant → authorize workflow →
authorize state → validate input schema → rate limit → invoke. Any step failure short-
circuits with a classified `CapabilityError` and no provider call.

### D5. Failure → event mapping
A `CapabilityError` carries a `Kind` (PRD §87) and a `Code` (PRD §63, e.g.
`capability.timeout`). The invoker returns both the normalized result and an optional
capability event so downstream transition logic (engine) can react. No raw provider
error is exposed to callers (PRD §2951).

### D6. Retry / timeout
Timeout is enforced at the invoker via `context.WithTimeout` and a per-policy deadline
(PRD §160). Retry applies only to retryable kinds (timeout, unavailable, transient
network) using exponential backoff + jitter (PRD §88); non-retryable kinds
(authorization, validation, business) short-circuit. Retry budget comes from the state
policy (`max_retry`, `retryable`) provided by the caller.

### D7. Idempotency
An invocation may carry `IdempotencyKey = workflow_instance_id + action_id` (PRD §64).
The invoker checks an idempotency store (an injected port) before invoking; a hit
returns the stored prior result. The store is a port so the real persistence can be
added later; tests use an in-memory implementation.

### D8. Mock provider default
When no real provider is bound (or in sandbox mode), the invoker uses the mock provider
(PRD §2064). Results are flagged `FromMock=true` so downstream code can distinguish and
the admin/test UI can show sandbox execution.

## Risks / Trade-offs

- [Provider timeouts still block a worker until deadline] → Use per-call context timeout
  derived from policy; keep default short.
- [Rate limiting needs a store] → Define a port; in-memory for now, Redis later.
- [Idempotency store added late could change signatures] → Design it as a port from the
  start (D7).
- [Mock provider could mask integration bugs] → Flag mock results explicitly and gate
  real execution behind explicit provider binding (PRD §2064).

## Migration Plan

- Land after `persistence-capabilities-policies` (provides `ICapabilityRepository`).
- No schema migration in this proposal; additive Go packages only.
- Rollback: remove the new packages; existing engine/auth code unaffected.

## Open Questions

None — decisions deferred to later slices (real MCP, HTTP) do not change this design.
