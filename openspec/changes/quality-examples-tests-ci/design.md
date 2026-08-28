## Context

Epic #7 needs the platform to be demonstrably correct, reproducible, and
open-source-ready. The engine, persistence, and MCP layers exist but lack regression
evidence and seedable examples. See proposal.md for motivation.

## Goals / Non-Goals

**Goals:**
- Seed 3 example workflows (PADEL, Order Makanan, Order Dokter) + intent registry so
  they run end-to-end.
- Golden conversation tests (PRD 125) as AI-behavior regression.
- Deterministic runtime tests without an LLM (PRD 126).
- E2E from an LLM/MCP mock → engine → state transition.
- A baseline load/throughput test for state transitions.
- Go backend CI (build/vet/test) in GitHub Actions.
- Open-source docs: README, deployment (local/docker/k8s), contributing, license.

**Non-Goals:**
- Real LLM/RAG in tests (mock only, PRD 170).
- Production load/soak profiling or formal benchmark harness.
- Product feature work (engine, MCP, UI).
- Implementing a separate k8s deployment — docs only.

## Decisions

### D1. Seeds live under a repeatable seed command + intent registry
A seed path (`cmd/seed` or an idempotent seeding function) upserts the three example
workflows and their intent entries under a fixed demo tenant + project (e.g. a
`demo` tenant with a `padel`/`retail`/`health` project). Seeding is idempotent
(upsert by workflow slug / intent id) so it can run on every migration (PRD 40.1,
§5). The canonical workflow JSON (already at `docs/padel-booking.workflow.json`) is
the source for each example; `order-food` and `order-doctor` are added alongside it.

### D2. Golden tests are data-driven conversation fixtures
Golden conversation tests are represented as declarative fixtures (one per workflow)
replaying user utterances and asserting the resolved intent + current state after
each turn (PRD 125). A tiny harness replays each fixture through intent resolution
and the engine (LLM classification is stubbed/mocked to return the expected intent)
and compares against expected state. Failures surface as a diff of expected vs actual
state per turn.

### D3. Deterministic runtime tests run directly on the engine, no LLM
The `engine-core` domain is domain-pure (no HTTP/DB/LLM dependency). Deterministic
tests construct workflows + contexts in memory and drive `ProcessEvent` through
`event → guard → transition`, asserting exact resulting state/status for both passing
and rejected transitions (PRD 126). These mirror the existing `engine_test.go`
pattern but are expanded into a systematic suite covering every guard operator and
priority-ordering rule.

### D4. E2E uses an LLM/MCP mock driving the MCP tool surface
An E2E harness starts the MCP server (or calls the MCP tool handlers directly) with a
deterministic mock client that: resolves an intent, proposes an event, and checks the
resulting state transition. No real LLM is called; the mock stands in for the 3rd-party
client (PRD 170). Assertions cover the full path: intent → workflow → state → event →
transition.

### D5. Load test is a focused baseline
A basic Go benchmark (`go test -bench`) or a small harness measures state-transition
throughput (transitions/sec) against an in-memory engine and a Postgres-backed
repository. Output is a baseline number, not a full soak/profiling suite.

### D6. Go backend CI is added as a dedicated workflow
A new `.github/workflows/go-ci.yml` runs `go build ./...`, `go vet ./...`, and
`go test ./...` for `apps/api` and `apps/worker` on PRs and pushes to `main`. It is
kept separate from the frontend `app-ci.yml` (different toolchain: Go vs bun) but
gated in the same branch-protection flow (PRD 122, `monorepo-identity`).

### D7. Docs follow a `docs/` layout
README keeps installation/architecture/quickstart/config; a new `docs/DEPLOYMENT.md`
documents local, docker, and k8s deployment; `CONTRIBUTING.md` and `LICENSE` are
completed and cross-linked. No runtime behavior changes.

## Schema Outline

No new DB tables. Seeds reuse the existing `workflow` and `intent` entities (from
`persistence-workflow-definitions` / engine intent registry). Golden/deterministic/E2E/
load tests are test fixtures and code, not schema.

## Risks / Trade-offs

- [Seeds drift from real workflows] → Seeds are generated from the canonical
  `*.workflow.json` files kept in `docs/`, so they stay in sync by construction.
- [Golden tests depend on intent classification] → Intent resolution is mocked/stubbed
  to be deterministic; the golden test asserts the state machine outcome, not LLM
  classification quality (PRD 170 boundary).
- [Load test flakiness on shared machines] → The load test asserts a loose lower
  bound and reports the measured number rather than failing on tight timing.
- [CI time increases] → Go CI is fast and parallel with the frontend matrix; kept
  scoped to build/vet/test.

## Migration Plan

1. Branch `feature/epic7-quality-examples-tests-ci`.
2. Add `order-food` + `order-doctor` workflow JSONs to `docs/` (mirroring
   `padel-booking.workflow.json`).
3. Implement idempotent seed command + intent registry entries for the demo tenant.
4. Add golden-conversation fixtures + harness (PRD 125).
5. Expand deterministic engine tests (PRD 126).
6. Add E2E harness with an LLM/MCP mock → engine → transition.
7. Add baseline load/throughput test.
8. Add `.github/workflows/go-ci.yml`.
9. Write `docs/DEPLOYMENT.md` (local/docker/k8s); update README, CONTRIBUTING, LICENSE.
10. `go build ./...`, `go vet ./...`, `go test ./...`; `bun run check` for frontend.
11. PR → review → merge.

**Rollback**: all additions are additive (tests, fixtures, docs, a CI workflow, seed
data); removing the workflow/seed/tests restores prior behavior with no data migration
(the seeds are idempotent and can be dropped).

## Open Questions

- Exact demo tenant/project identifiers and demo owner credentials for the seeded
  environment (default to a `demo` tenant to be confirmed during implementation).
- Whether the Go CI should live in its own workflow file or extend `app-ci.yml`
  (default: separate `go-ci.yml` for toolchain isolation).
