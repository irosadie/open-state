# monorepo-identity Specification

## Purpose
Define the repository identity (name, package scope, module paths) and the
naming conventions used across the monorepo.

## Requirements

### Requirement: Consistent product name

The product and repository SHALL be branded **OpenState** end-to-end.

#### Scenario: Display identity

- GIVEN a user views the repository or runs the application
- THEN the product name is "OpenState"
- AND no legacy starter branding remains in display metadata

### Requirement: Go module paths

The repository SHALL use the canonical `github.com/irosadie/open-state/*` path for Go modules.

#### Scenario: Imports

- GIVEN a Go file imports an internal package
- THEN it uses `github.com/irosadie/open-state/...`
- AND `go build ./...` succeeds without legacy module references

### Requirement: Frontend package scope

The repository SHALL use the `@openstate/*` scope for frontend packages.

#### Scenario: TypeScript imports

- GIVEN a TS file imports a shared package
- THEN it uses `@openstate/...`
- AND `tsc --noEmit` succeeds

### Requirement: Infrastructure naming

Containers and the database SHALL use the `openstate` prefix.

#### Scenario: Docker compose

- GIVEN the compose stack is started
- THEN containers are named `openstate-*`
- AND the database is named `openstate`
