## Purpose

Provide tenant/project-scoped intent metadata and natural-language examples so an external LLM can select a canonical intent before asking OpenState to resolve and execute its workflow.

## ADDED Requirements

### Requirement: Intent discovery tool

The MCP server SHALL expose a read-only `list_intents` tool that accepts a tenant identifier and project identifier and returns every routable intent in that scope.

#### Scenario: Discover intents for a project

- **WHEN** the LLM calls `list_intents` with a valid tenant and project
- **THEN** the server returns each routable intent with its canonical `id`, name, description, natural-language examples, project identifier, and mapped workflow slug
- **AND** the returned records are sufficient for the LLM to choose an intent without guessing a workflow identifier

### Requirement: Natural-language routing examples

Each routable intent SHALL provide example user utterances that describe the language it covers.

#### Scenario: Match a court-booking request

- **WHEN** the catalog contains the `BOOKING_PADEL` intent
- **THEN** its metadata includes an example equivalent to “saya mau order lapangan” or “saya mau booking lapangan padel”
- **AND** the catalog identifies the mapped padel booking workflow

### Requirement: Tenant and project isolation

The intent discovery tool SHALL return only intents belonging to the requested tenant and project and SHALL NOT fall back to a global catalog.

#### Scenario: Isolate another project's intents

- **WHEN** the LLM calls `list_intents` for project A
- **THEN** the response excludes intents belonging to project B, even when both projects belong to the same tenant

#### Scenario: Reject incomplete scope

- **WHEN** the LLM calls `list_intents` without a tenant or project identifier
- **THEN** the server returns a validation error

### Requirement: Routable catalog contents

The intent discovery tool SHALL list only intents whose mapped workflow is available for routing, and SHALL return an empty list when the requested scope has no routable intents.

#### Scenario: Exclude unavailable workflow mappings

- **WHEN** an intent is mapped to a workflow that is not published or no longer exists
- **THEN** that intent is excluded from the discovery response

#### Scenario: No intents configured

- **WHEN** a valid tenant and project have no routable intents
- **THEN** the server returns an empty intent list
