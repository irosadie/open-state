## Why

Epic **#7 (Quality)** exists to prove the platform is correct, reproducible, and
ready for open-source contribution. The runtime engine (`engine-core`), persistence
(`persistence-*`), and MCP surface (`mcp-server-runtime`, `mcp-orchestrator-tools`)
are implemented, but there is no evidence the engine behaves deterministically at
scale, no reproducible example workflows seeded into the database, no golden
conversation regression tests (PRD 125), no deterministic runtime tests without an
LLM (PRD 126), no end-to-end test from an LLM/MCP mock through the engine to a state
transition, no baseline load/throughput test, and CI does not yet gate the Go
backend. This slice closes those gaps and completes the open-source documentation
(README, deployment, contributing, license) so the repo can be used and
contributed to.

## What Changes

- **NEW** seeded example workflows + intent registry: `padel-court-booking`,
  `order-food`, and `order-doctor` are registered for a seeded tenant/project with
  matching intent entries (`BOOKING_PADEL`, `ORDER_FOOD`, `ORDER_DOCTOR`) so they can
  be resolved and executed end-to-end (PRD §40.1).
- **NEW** golden conversation tests for each example workflow: a per-conversation
  test harness that replays user utterances and asserts the resolved intent /
  current state after each turn (PRD 125), used as AI-behavior regression tests.
- **NEW** deterministic runtime tests without an LLM: exercise `event → guard →
  transition` on the engine directly with pure inputs and assert deterministic
  outcomes (PRD 126).
- **NEW** end-to-end test: an LLM/MCP **mock** drives the MCP tools → engine → state
  transition, proving the full path works without a real LLM.
- **NEW** baseline load test: measure state-transition throughput under load to give
  operators an initial performance signal.
- **NEW** Go backend CI: `go build`, `go vet`, and `go test` for `apps/api` and
  `apps/worker` in GitHub Actions, complementing the existing frontend `app-ci.yml`.
- **UPDATED** open-source documentation: README (installation, architecture,
  quickstart, configuration), deployment guide (local, docker, k8s), contribution
  guide, and license.

## Capabilities

### New Capabilities

- `quality/examples`: seedable, executable example workflows + intent registry
  entries (PADEL, Order Makanan, Order Dokter).
- `quality/golden-tests`: golden conversation test harness + per-workflow test cases.
- `quality/deterministic-tests`: LLM-free deterministic runtime engine test suite.
- `quality/e2e-tests`: LLM/MCP mock → engine → state-transition end-to-end test.
- `quality/load-tests`: baseline state-transition throughput load test.
- `quality/ci`: Go backend build/vet/test gates in GitHub Actions.

### Modified Capabilities

- `monorepo-identity`: extend CI to cover the Go backend (currently only the
  frontend matrix is gated).

## Impact

- **`apps/api/`** — seed command/data, golden-test fixtures, deterministic engine
  tests, E2E mock harness, load-test tooling.
- **`.github/workflows/`** — add a Go backend CI workflow (or extend `app-ci.yml`).
- **`docs/`** — deployment guide (local, docker, k8s); README/architecture updates.
- **`CONTRIBUTING.md`**, **`LICENSE`** — documented contribution flow and license
  notice.
- **No** changes to the runtime engine, persistence schema, MCP tool contract, or
  frontend UI logic (test surface only, plus docs + CI).

## Non-Goals

- A real (non-mock) LLM/RAG integration in tests — E2E uses a deterministic mock
  (PRD 170 keeps the LLM external).
- Production-grade load/soak profiling or a formal benchmark harness — only a basic
  baseline throughput test.
- Product feature work (engine, MCP, UI) — testing/docs/CI only.
- A separate k8s deployment implementation — only the deployment documentation.

## Dependencies

- `engine-core` (state machine + guard evaluation) — deterministic tests exercise it.
- `persistence-*` (workflow/instance/event/context repositories) — seeds + E2E use
  them.
- `mcp-server-runtime` + `mcp-orchestrator-tools` — E2E drives these tools via a mock.
- `monorepo-identity` — CI structure baseline.
- Epic #7.

## Notes

- PRD 123 (Testing Strategy) lists the test types (unit, integration, simulation,
  contract, end-to-end, load, security); this slice covers simulation/golden,
  deterministic unit, E2E, and a baseline load test.
- PRD 122 (CI Validation) motivates `validate/simulate/lint/export/import` tooling;
  the CI spec here gates the existing toolchain, with the Go gates added first.
- PRD 40.1 (Intent Registry) fixes the example intents and their workflow mapping.
