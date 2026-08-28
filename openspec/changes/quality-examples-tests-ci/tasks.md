## 1. Example workflows + seed (Skill: db-sqlc-schema, api-feature)

- [ ] 1.1 Add `order-food` and `order-doctor` canonical workflow JSONs under `docs/`
      (mirroring `docs/padel-booking.workflow.json` format: nodes, transitions,
      guards, policies, triggers) (PRD §40.1)
- [ ] 1.2 Read `.agents/guides/api-entity.md`, `.agents/guides/api-repository.md`
- [ ] 1.3 Implement an idempotent seed path (e.g. `apps/api/cmd/seed`) that upserts
      the three workflows (padel-court-booking, order-food, order-doctor) under a
      demo tenant/project
- [ ] 1.4 Register intent entries `BOOKING_PADEL`, `ORDER_FOOD`, `ORDER_DOCTOR` with
      sample `examples` phrases and workflow mapping (PRD 40.1)
- [ ] 1.5 Verify re-running the seed does not duplicate rows (idempotent upsert)
- [ ] 1.6 Verify seeds are scoped to the demo tenant/project (PRD §4)

## 2. Golden conversation tests (Skill: api-feature)

- [ ] 2.1 Read `.agents/guides/api-service.md`, `.agents/guides/api-dto.md`
- [ ] 2.2 Create golden conversation fixtures for each example workflow (User
      utterances + Expected state per turn) (PRD 125)
- [ ] 2.3 Build a replay harness: stub intent classification to the expected intent,
      replay turns through the engine, assert actual vs expected state
- [ ] 2.4 Assert failure surfaces a diff of expected vs actual state per turn
- [ ] 2.5 Ensure golden tests run in CI without a real LLM (PRD 170)

## 3. Deterministic runtime tests (Skill: api-feature, api-bugfix)

- [ ] 3.1 Expand `apps/api/internal/domain/engine` tests to cover every guard
      operator (`==`, `!=`, `>`, `>=`, `<`, `<=`, `IN`, `EXISTS`) with passing +
      failing cases (PRD 126)
- [ ] 3.2 Add AND/OR guard-grouping tests
- [ ] 3.3 Add priority-ordering tests (highest-priority passing transition wins,
      PRD 33-34)
- [ ] 3.4 Add rejection tests (disallowed event → no state change + structured
      rejection)
- [ ] 3.5 Verify all deterministic tests run without an LLM

## 4. E2E test with LLM/MCP mock (Skill: api-feature, api-code-review)

- [ ] 4.1 Create a deterministic mock client standing in for the 3rd-party LLM (PRD 170)
- [ ] 4.2 Drive the real MCP tool handlers (`resolve_intent`/`get_active_workflow`,
      `propose_event`) via the mock
- [ ] 4.3 Assert the full path: intent → workflow → state → event → transition
- [ ] 4.4 Assert persisted state via the repository layer after the transition
- [ ] 4.5 Ensure the E2E is repeatable (fixed mock responses)

## 5. Baseline load test (Skill: api-feature)

- [ ] 5.1 Add an in-memory engine throughput test (transitions/sec) via Go benchmark
- [ ] 5.2 Add an optional Postgres-backed throughput run
- [ ] 5.3 Use a loose lower bound to avoid flakiness; report measured values
- [ ] 5.4 Add a Go test/bench target wired into the repo scripts

## 6. Go backend CI (Skill: api-code-review)

- [ ] 6.1 Add `.github/workflows/go-ci.yml` running `go build ./...`, `go vet ./...`,
      and `go test ./...` for `apps/api` and `apps/worker` (PRD 122)
- [ ] 6.2 Configure scoped triggers on `apps/**`, `go.work*`, and Go module files
- [ ] 6.3 Ensure golden/deterministic/E2E tests run in the CI `go test ./...` step
- [ ] 6.4 Verify the existing frontend `app-ci.yml` still passes (no regression)

## 7. Documentation (Skill: docs-openapi, api-code-review)

- [ ] 7.1 Write `docs/DEPLOYMENT.md`: local, docker, and k8s deployment (PRD 122)
- [ ] 7.2 Update `README.md`: installation, architecture, quickstart, configuration
- [ ] 7.3 Complete/cross-link `CONTRIBUTING.md` and `LICENSE`
- [ ] 7.4 Document the new seed command and example workflows in the README

## 8. Quality gate (Skills: api-code-review, api-bugfix)

- [ ] 8.1 `go build ./...`, `go vet ./...`, `go test ./...` pass in `apps/api` and
      `apps/worker`
- [ ] 8.2 `bun run check` passes (frontend lint/typecheck/test/build unaffected)
- [ ] 8.3 Verify Go CI workflow passes in GitHub Actions on the feature branch
- [ ] 8.4 `api-code-review`: seeds idempotent + tenant-scoped, tests deterministic,
      E2E uses a mock only (PRD 170), no feature-behavior changes, files end with
      newline
