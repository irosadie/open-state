## Overview

Domain-pure runtime engine core. No HTTP/DB/LLM dependencies — the engine
speaks to repository **ports** (interfaces), implemented later by PostgreSQL
(Epic #3). Follows Clean Architecture (PRD §72) and modular monolith (PRD §175).

## Decisions

### D0. Hierarchy

The domain follows the hierarchy:

```
Perusahaan (Tenant) → Project → Intent → Workflow → State
```

- **Tenant** owns many **Projects** (business areas: resto, padel, dokter).
- **Project** owns many **Intents**.
- **Intent** resolves to a **Workflow** (state machine) within the same project.
- **Workflow** contains many **States** (nodes + transitions).

This replaces the earlier flat `Tenant → Workflow → State` model. Scoping is
enforced at every repository port (`tenantID`, `projectID`).

### D1. Package layout
`apps/api/internal/domain/engine/`:
```
engine/
├── model.go            // entities + enums
├── guard.go            // guard evaluator
├── state_machine.go    // executor
├── intent.go           // intent registry + resolver
├── ports.go            // repository interfaces
└── *_test.go
```
No `infrastructure`, `interfaces`, or `application` coupling inside this package.

### D2. Transition selection
Transitions from a state are evaluated by **priority (ascending)**; the first
whose guard passes wins (PRD §34). Ambiguity (equal priority, both pass) is a
validation error surfaced by the engine.

### D3. Guard model
Guards are pure data (`GuardGroup{logic: AND|OR, conditions[]}`) evaluated by a
safe function. No arbitrary code (PRD §35). Context is a flat map keyed by
dot-paths; `EXISTS` checks key presence.

### D4. Repository ports
Interfaces defined in `engine/ports.go` return/accept domain entities only.
Optimistic concurrency is exposed via a `Version` field on `WorkflowInstance`.
In-memory fakes in tests prove the engine is DB-agnostic.

### D5. Idempotency & suspension
Deferred to later proposals (`engine-suspension-idempotency`,
`engine-context-resolver`) to keep this proposal focused and reviewable.
The ports include hooks (idempotency key on events; status on instances) so the
later proposals don't require breaking changes.

## Risks / Notes
- **Scope creep**: only domain core here; resist adding HTTP/DB.
- **Transition ambiguity**: enforce priority + validation.
- **Guard type coercion**: define strict comparison semantics to avoid surprises.
