## Why

Epic **#3 (Data & Persistence)** requires the Capability Registry and Policy storage to be persistent (PRD 59-64, 3.13, 18). This is the fifth persistence slice: **capabilities, capability bindings, and policies** — behind `ICapabilityRepository` and a pgx adapter.

Capabilities are logical operations (PRD 3.11, 16) referenced by states; the registry maps logical capability → provider (MCP/INTERNAL/HTTP, PRD 59). Bindings scope a capability's availability at tenant/workflow/state level with most-restrictive-wins resolution (PRD 60). Policies hold runtime/security/business constraints (PRD 3.13). Secrets are never stored here — only `credential_reference` (PRD 61).

## What Changes

- **NEW** — `apps/api/db/migrations/00006_capabilities.sql` goose migration creating:
  - `capabilities` — the Capability Registry (PRD 59).
  - `capability_bindings` — tenant/workflow/state scoped availability (PRD 60).
  - `policies` — tenant/workflow/state scoped constraints (PRD 3.13, 12).
- **NEW** — `apps/api/db/queries/capability.sql` sqlc query file.
- **NEW** — Domain entities `Capability`, `CapabilityBinding`, `Policy` in `apps/api/internal/domain/entities/`.
- **NEW** — Repository interface `apps/api/internal/domain/repositories/capability_repository.go` (`ICapabilityRepository`).
- **NEW** — pgx adapter `apps/api/internal/infrastructure/database/pgx_capability_repository.go`.
- **REGENERATED** — sqlc-generated Go under `apps/api/internal/infrastructure/db/`.
- Uses **`db-sqlc-schema`** and **`api-feature`** skills. `docs-openapi` not touched (no public endpoint).

## Capabilities

### New Capabilities

- `backend/persistence/capabilities-policies`: tenant-isolated Capability Registry, scoped bindings, and policies behind `ICapabilityRepository` and a pgx adapter.

### Modified Capabilities

- None (references `workflow_instances`/`states` FK from earlier slices where applicable).

## Impact

- **`apps/api/db/migrations/`** — add `00006_capabilities.sql`.
- **`apps/api/db/queries/`** — add `capability.sql`.
- **`apps/api/internal/infrastructure/db/`** — sqlc-generated code regenerated.
- **`apps/api/internal/domain/entities/`** — add `capability.go`, `capability_binding.go`, `policy.go`.
- **`apps/api/internal/domain/repositories/`** — add `capability_repository.go`.
- **`apps/api/internal/infrastructure/database/`** — add `pgx_capability_repository.go`.
- **No** changes to web, worker, shared packages, OpenAPI, docker.

## Non-Goals

- The Capability Resolver / invocation pipeline (auth, rate-limit, invoke, PRD 62-64) — separate epic.
- MCP client/server integration — separate epic.
- Secrets management — only `credential_reference` stored (PRD 61).
- Audit persistence (capability.invoked/denied events) — separate change (`persistence-audit-adapter`).
- Any HTTP layer.

## Dependencies

- `persistence-workflow-definitions` (states FK), `persistence-runtime-instances` (optional instance bindings).
- Epic #3.
