## Why

Production workflows must be robust: interrupted mid-flow must suspend &
resume without losing context (PRD §42-43), duplicate events must be
deduplicated (PRD §30), and concurrent updates must be safe via optimistic
locking (PRD §31). The engine core (Proposal `engine-domain-core`) provides the
base executor; this proposal adds these reliability behaviors.

## What Changes

- **Suspension/resume**: `SuspendWorkflow(instanceId)` and
  `ResumeWorkflow(instanceId)` on the engine; instance status gains
  `SUSPENDED`; preserves state, context, history, version (PRD §43).
- **Mid-flow interruption**: `product.change_requested`-style events allowed to
  route out of a running state and back (PRD §43.1) — supported by the executor
  already; this adds suspension lifecycle hooks.
- **Optimistic concurrency**: `Version` counter on `WorkflowInstance`; every
  transition/update requires the expected version, else `CONFLICT` (PRD §31).
- **Idempotency**: `Event` carries an `idempotencyKey`; processing checks a
  dedup map (via port) before applying (PRD §30).
- **Unit tests**.

## Capabilities

### New Capabilities

- `engine/suspend-resume`: instance suspension & resumption with context
  preservation.
- `engine/concurrency`: optimistic versioning & conflict detection.
- `engine/idempotency`: event dedup by idempotency key.

## Impact

- **`apps/api/internal/domain/engine/`** — extend executor + add
  suspend/resume, version, idempotency.
- Repository ports extended with version-aware update & idempotency check
  (implementation still in Epic #3).
- Depends on `engine-domain-core` and `engine-context-resolver` (preserve
  context on suspend).

## Non-Goals

- No PostgreSQL implementation (Epic #3).
- No MCP tools (Epic #4).
- No distributed locks (Redis advisory) — optimistic concurrency only (PRD §130).
