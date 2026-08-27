# monorepo-identity Specification

## Purpose

Define the repository identity (name, package scope, module paths) and the
naming conventions used across the monorepo.

## MODIFIED Requirements

### Requirement: Consistent product name

The product and repository are branded **OpenState** end-to-end instead of `vibecoding-starter`.

#### Scenario: Display identity

- GIVEN a user views the repository or runs the application
- THEN the product name is "OpenState"
- AND no `vibecoding-starter` branding remains in display metadata

### Requirement: Go module paths

Go modules use the `github.com/irosadie/open-state/*` path instead of `github.com/vibecoding-starter/*`.

#### Scenario: Imports

- GIVEN a Go file imports an internal package
- THEN it uses `github.com/irosadie/open-state/...`
- AND `go build ./...` succeeds without `vibecoding-starter` references

### Requirement: Frontend package scope

Frontend packages use the `@openstate/*` scope instead of `@vibecoding-starter/*`.

#### Scenario: TypeScript imports

- GIVEN a TS file imports a shared package
- THEN it uses `@openstate/...`
- AND `tsc --noEmit` succeeds

### Requirement: Infrastructure naming

Containers and the database use the `openstate` prefix instead of `vibecoding-starter`.

#### Scenario: Docker compose

- GIVEN the compose stack is started
- THEN containers are named `openstate-*`
- AND the database is named `openstate`
