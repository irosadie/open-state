## Why

The Runtime Engine (Epic #2) is the foundation of OpenState: it executes
workflows deterministically so conversations always follow business rules,
independent of the LLM. This first proposal — **engine-domain-core** — builds
the domain-pure heart: entities, state machine executor, guard evaluator,
lifecycle, and intent resolution. It has **no HTTP/DB/LLM dependency**, so it
can be fully unit-tested in isolation (PRD §126) and later persisted via
repository ports (Epic #3).

Without this, nothing else (MCP, LLM integration, Admin console) can be built.

## What Changes

- **New Go package** `apps/api/internal/domain/engine/` — the domain core.
- **Domain entities**: `Workflow`, `State`, `Transition`, `Event`, `Guard`,
  `WorkflowInstance`, `StateInstance`, `Policy`, `Capability`.
- **State machine executor**: `process(event)` → validate → guard eval →
  transition → snapshot (PRD §152).
- **Guard evaluator**: deterministic, JSON-based operators (`==`, `!=`, `>`,
  `>=`, `<`, `<=`, `IN`, `EXISTS`, `AND/OR/NOT`) (PRD §35).
- **Lifecycles**: workflow
  (`CREATED→RUNNING→WAITING→COMPLETED/CANCELLED/FAILED/EXPIRED`) and state
  (`ENTERING→ACTIVE→WAITING→EXITING→COMPLETED`) (PRD §10, §11).
- **Intent resolution**: conversation → intent → workflow → initial state
  (PRD §40.1).
- **Repository ports (interfaces)**: `WorkflowRepo`, `InstanceRepo`,
  `EventRepo` — defined here (domain), implemented by PostgresAdapter in Epic #3.
- **Unit tests** — deterministic, no LLM/DB (PRD §126).

## Capabilities

### New Capabilities

- `engine/domain-model`: Workflow/State/Transition/Event/Guard/Instance/Policy
  entities + enums + lifecycle states.
- `engine/state-machine`: deterministic executor — process(event) → guard →
  transition → snapshot.
- `engine/guard-eval`: safe, data-driven guard evaluator (no arbitrary code).
- `engine/intent-resolver`: conversation → intent → workflow → initial state.
- `engine/repository-ports`: domain repository interfaces (Workflow/Instance/Event).

## Impact

- **`apps/api/internal/domain/engine/`** — new package (entities, executor,
  guard, resolver, ports).
- **`packages/go-shared/`** — no change required; reused for DomainError.
- No HTTP/DB/LLM changes in this proposal.
- Existing `apps/api` auth domain untouched.
- Quality gate: `go build ./...`, `go vet ./...`, `go test ./...`.

## Non-Goals

- No PostgreSQL implementation (Epic #3).
- No MCP server / LLM integration (Epic #4).
- No suspension/resume & idempotency (separate proposal `engine-context-resolver`
  / `engine-suspension-idempotency`).
- No context resolver (separate proposal).
