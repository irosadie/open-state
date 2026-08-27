## Context

The Capability Registry is owned by the Orchestrator (PRD 59), separate from the State Builder (PRD 16-17). Capabilities are logical references resolved to providers at runtime; bindings scope availability (PRD 60) with most-restrictive-wins; policies hold constraints (PRD 3.13). This change adds the capability/policy persistence slice of epic #3.

## Goals / Non-Goals

**Goals:**
- Schema for `capabilities`, `capability_bindings`, `policies`.
- Domain entities + `ICapabilityRepository`.
- pgx adapter backed by sqlc.
- Tenant scoping + scoped binding resolution.

**Non-Goals:**
- Capability Resolver / invocation / authorization enforcement (separate epic).
- MCP client integration (separate epic).
- Secrets storage — only `credential_reference` (PRD 61).
- Audit logging of capability calls (separate change).
- Any HTTP layer.

## Decisions

### D1: Capability registry as core table
`capabilities` stores the PRD 59 shape: `name` (unique per tenant), `description`, `provider_type` (`MCP`/`INTERNAL`/`HTTP`/`FUTURE`), `provider_id`, `input_schema`/`output_schema` (JSONB), `status`, `version`. VARCHAR + Go constants for `provider_type` and `status`.

### D2: Scoped bindings with most-restrictive-wins
`capability_bindings` associates a capability with a `scope_type` (TENANT/WORKFLOW/STATE) + `scope_id`, and a `permission` (`ALLOW`/`DENY`). Resolution (PRD 60) is implemented by the resolver later; the schema stores all scopes so the most restrictive can be computed. Global capabilities are rows with tenant scope.

### D3: Policies keyed by scope
`policies` stores `type` (JSONB or named) with `content` JSONB (PRD 3.13, 12: timeout, retry, human handoff, max_retries, interruptible). Scoped to TENANT/WORKFLOW/STATE via `scope_type`+`scope_id`. Policy evaluation is a later engine concern.

### D4: No secrets
`capabilities` carries only `credential_reference` (optional string), never credentials/tokens (PRD 61).

### D5: Tenant isolation everywhere
All three tables carry `tenant_id`; all queries filter by it (PRD 4, 96). Names unique per tenant.

## Schema Outline

```
capabilities
  id            UUID PK
  tenant_id     UUID NOT NULL
  name          VARCHAR NOT NULL        -- logical capability, e.g. payment.create
  description   TEXT
  provider_type VARCHAR NOT NULL        -- MCP/INTERNAL/HTTP/FUTURE (PRD 59)
  provider_id   VARCHAR
  input_schema  JSONB NOT NULL DEFAULT '{}'
  output_schema JSONB NOT NULL DEFAULT '{}'
  status        VARCHAR NOT NULL DEFAULT 'ACTIVE'   -- ACTIVE/INACTIVE/DISABLED
  version       INT NOT NULL DEFAULT 1
  credential_reference VARCHAR             -- PRD 61, never secrets
  created_at / updated_at
  UNIQUE(tenant_id, name)

capability_bindings
  id            UUID PK
  tenant_id     UUID NOT NULL
  capability_id UUID NOT NULL REFERENCES capabilities(id) ON DELETE CASCADE
  scope_type    VARCHAR NOT NULL        -- TENANT/WORKFLOW/STATE (PRD 60)
  scope_id      VARCHAR NOT NULL        -- tenant/workflow/state id
  permission    VARCHAR NOT NULL DEFAULT 'ALLOW'   -- ALLOW/DENY
  created_at / updated_at
  UNIQUE(tenant_id, capability_id, scope_type, scope_id)

policies
  id            UUID PK
  tenant_id     UUID NOT NULL
  scope_type    VARCHAR NOT NULL        -- TENANT/WORKFLOW/STATE
  scope_id      VARCHAR NOT NULL
  type          VARCHAR NOT NULL        -- e.g. timeout / retry / human_handoff / workflow (PRD 3.13, 12)
  content       JSONB NOT NULL DEFAULT '{}'
  created_at / updated_at
  UNIQUE(tenant_id, scope_type, scope_id, type)
```

Indexes: `capability_bindings(tenant_id, capability_id)`, `policies(tenant_id, scope_type, scope_id)`.

## Risks / Trade-offs

- **Risk: binding scope_id is VARCHAR but FK targets are UUIDs** → Mitigation: bindings can reference workflows/states (UUID) or tenants; keep VARCHAR for flexibility and index it. Hard FK to a polymorphic scope is not feasible.
- **Risk: most-restrictive-wins not enforced in schema** → Mitigation: enforcement is a resolver concern (later). Schema stores permission per scope; the resolver computes the effective permission.
- **Risk: policy content is free-form JSONB** → Mitigation: intentional for extensibility; the Policy Engine validates known types later.

## Migration Plan

1. Branch `feature/epic3-persistence-capabilities-policies`.
2. Add migration `00006_capabilities.sql` (Up + Down).
3. Add `db/queries/capability.sql`.
4. Run `sqlc generate`.
5. Add entities + `ICapabilityRepository`.
6. Implement `pgx_capability_repository.go`.
7. `go build ./...`, `go vet ./...`, `go test ./...`.
8. Optional smoke: `goose up`; create capability + binding + policy; resolve bindings per scope.
9. PR → review → merge.

**Rollback**: migration `Down` drops the three tables; feature branch isolates changes.

## Open Questions

- Whether `policies` should be merged into `workflow_versions.definition.policy` JSONB only. Decision: keep a relational `policies` table for tenant-level and state-level policies that live outside the immutable workflow definition (PRD 3.13, 12); the definition's own `policy` remains JSONB in `workflow_versions`.
