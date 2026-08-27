## Overview

Reliability behaviors on top of the engine core: suspension/resume,
optimistic concurrency, and idempotency. Domain-pure; ports extended for later
Postgres implementation (Epic #3).

## Decisions

### D1. Suspension is a status, not a new state machine
`SUSPENDED` is a `WorkflowInstanceStatus`. The underlying state/context/history
are preserved; resume re-enters the saved current state (PRD §43).

### D2. Optimistic concurrency
A monotonic `Version` integer on the instance. Every state-changing operation
must present the expected version; repository port enforces
`UPDATE ... WHERE version = expected` → affected rows 0 = CONFLICT (PRD §31,
§130). No distributed locks.

### D3. Idempotency key on events
Every external event carries an `idempotencyKey`. The engine checks the event
repository before applying; a duplicate is a no-op (PRD §30).

### D4. Port extensions
`InstanceRepository` gains version-aware update; `EventRepository` gains
idempotency check/mark. Concrete SQL comes in Epic #3.

## Risks / Notes
- **Stale version**: clients must handle CONFLICT (retry or surface).
- **Idempotency scope**: key uniqueness is per (tenant, key).
