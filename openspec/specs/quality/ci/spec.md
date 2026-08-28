# quality/ci Specification

## Purpose
Define a Go backend CI gate (build, vet, test) in GitHub Actions so the Go backend is
validated on every PR and push to main, complementing the existing frontend CI
(PRD 122, `monorepo-identity`).

## Requirements

### Requirement: Go backend build

CI SHALL build the Go backend packages on every relevant change.

#### Scenario: Build on PR

- **WHEN** a pull request modifies `apps/**`, `packages/**`, or the Go module files
- **THEN** `go build ./...` passes for `apps/api` and `apps/worker`

### Requirement: Go vet

CI SHALL run `go vet` on the Go backend.

#### Scenario: Vet on PR

- **WHEN** the Go build runs in CI
- **THEN** `go vet ./...` passes for `apps/api` and `apps/worker` (PRD §185)

### Requirement: Go test

CI SHALL run the Go backend test suite, including the new quality tests.

#### Scenario: Test on PR

- **WHEN** the Go CI job runs
- **THEN** `go test ./...` passes for `apps/api` and `apps/worker`
- **AND** the golden, deterministic, and E2E tests execute as part of the suite

### Requirement: Trigger on relevant paths

The Go CI workflow SHALL trigger on changes to the Go backend and its workspace.

#### Scenario: Scoped trigger

- **GIVEN** the Go module files, `apps/api`, `apps/worker`, or `go.work*` change
- **THEN** the Go CI workflow triggers
- **AND** it does not redundantly run on frontend-only changes unless the workspace
  is affected
