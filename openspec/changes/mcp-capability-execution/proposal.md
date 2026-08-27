## Why

Epic **#4 (MCP & Integrasi)** requires the platform to *execute* external capabilities
safely and deterministically. The PRD mandates a `CapabilityProvider` abstraction so the
core engine stays MCP-agnostic (PRD §172), a full security chain before invocation
(PRD §62), structured failure results that become events (PRD §63), idempotency for
side-effecting actions (PRD §64), and retry/timeout policies (PRD §88, §160). Without
this layer, no workflow state can invoke an external MCP tool, and the platform would
leak raw provider errors or bypass tenant/project/workflow/state authorization.

This first slice — **mcp-capability-execution** — builds the domain + application layer:
the Capability Resolver, the `CapabilityProvider` port, the execution pipeline, mock
providers, retry/timeout/idempotency, and structured result normalization. It has **no
HTTP controller and no MCP SDK dependency** (PRD §2559), so it is fully unit-testable and
sits directly on `ICapabilityRepository` from `persistence-capabilities-policies`.

## What Changes

- **NEW Go package** `apps/api/internal/domain/capability/` — the execution domain core:
  - `CapabilityProvider` **port** (interface): `Invoke(ctx, Invocation) (Result, error)`.
  - `CapabilityResolver`: logical capability → resolved provider + schema, honoring
    tenant/project/workflow/state bindings with most-restrictive-wins (PRD §60).
  - `CapabilityInvoker` / execution pipeline: authorize → validate input schema →
    rate-limit → invoke → normalize result → emit result (PRD §153).
  - `Invocation` / `InvocationResult` value objects + normalized `CapabilityError`
    classified per PRD §87 (`TIMEOUT`, `UNAUTHORIZED`, `VALIDATION`, `EXTERNAL`, …).
  - Failure-to-event mapping (`capability.timeout`, `.unauthorized`, `.validation_failed`,
    `.unavailable`, `.business_error`) per PRD §63.
  - Idempotency support via `idempotency_key = workflow_instance_id + action_id` (PRD §64).
  - Retry with exponential backoff + jitter, only for retryable errors (PRD §88); timeout
    policy (PRD §160).
- **NEW mock provider** `apps/api/internal/infrastructure/capability/mock_provider.go`
  (default sandbox/mocked mode per PRD §2064) implementing `CapabilityProvider`.
- **Repository port usage** — reads capabilities/bindings/policies from
  `ICapabilityRepository` (defined in `persistence-capabilities-policies`).
- **Unit tests** — deterministic, no HTTP/DB/LLM (PRD §126).

## Capabilities

### New Capabilities

- `mcp/capability-execution`: safe, deterministic capability execution — resolver,
  provider abstraction, security chain, normalization, retry/timeout, idempotency.

### Modified Capabilities

- None (new capability introduced by this epic).

## Impact

- **`apps/api/internal/domain/capability/`** — new package (port, resolver, invoker,
  value objects, errors).
- **`apps/api/internal/infrastructure/capability/`** — mock provider implementation.
- **`apps/api/internal/domain/repositories/`** — consumes `ICapabilityRepository`
  interface (no change; implemented in `persistence-capabilities-policies`).
- **`packages/go-shared/`** — reused for `DomainError`; no change required.
- **No** HTTP controllers/routes, OpenAPI, web, worker, docker in this proposal.
- Quality gate: `go build ./...`, `go vet ./...`, `go test ./...`; coverage > 80%.

## Non-Goals

- The MCP **client/server** integration (actual SDK connection) — separate proposal
  `mcp-server-runtime`.
- Any HTTP controller/route for capability management — separate proposal
  `capability-admin-api`.
- The `resolve_intent` / `get_active_workflow` / `get_context` MCP tools — separate
  proposal `mcp-server-runtime`.
- LLM multi-provider abstraction & structured output — out of scope for this epic slice.
- Secrets management — only consumes `credential_reference` (PRD §61); actual secret
  resolution lives in `mcp-server-runtime`.
- Capability/policy **persistence** — already covered by `persistence-capabilities-policies`.

## Dependencies

- `persistence-capabilities-policies` (provides `ICapabilityRepository` + entities).
- `engine-domain-core` (workflow/instance/state entities used in authorization context).
- Epic #4.
