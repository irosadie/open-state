## Purpose

Integrates the State Builder frontend with the Builder API: shared Zod schemas and
response types, API route/query-key constants, react-query hooks, and the replacement
of browser-local PGlite draft persistence with API-backed draft save/load.

## ADDED Requirements

### Requirement: State Builder persists drafts to the API
The system SHALL let the State Builder save and load workflow definition drafts through
the Builder API instead of the browser-local PGlite store.

- Draft save SHALL call the create/update workflow API.
- Draft load SHALL fetch the workflow by id from the API.
- The State Builder in-memory React Flow state (nodes, edges, undo/redo, validation)
  SHALL remain unchanged.

#### Scenario: Save draft to API
- **WHEN** the operator edits a workflow in the State Builder and auto-save triggers
- **THEN** the draft is persisted to the Builder API and the save status is reflected in the UI.

#### Scenario: Load draft from API
- **WHEN** the State Builder hydrates on open with a persisted workflow id
- **THEN** the workflow definition is fetched from the API and materialized into nodes/edges.

### Requirement: Shared workflow schema and types
The system SHALL define shared Zod schemas and response types for the workflow draft
contract consumed by the frontend.

- The schema SHALL validate workflow draft create/update payloads.
- Response types SHALL model the workflow and workflow-version API responses.

#### Scenario: Valid workflow payload
- **WHEN** a workflow draft payload satisfies the shared Zod schema
- **THEN** it is accepted and forwarded to the API.

### Requirement: React-query hooks for workflows
The system SHALL expose react-query hooks for workflow list, get, create/upsert draft,
and publish.

- Hooks SHALL call the API through the axios instance, never directly inside JSX.
- Mutations SHALL invalidate the workflow list query on success.

#### Scenario: Create-or-update draft hook
- **WHEN** the create-or-update draft mutation resolves successfully
- **THEN** the workflow list query is invalidated so downstream views reflect the change.

### Requirement: No direct HTTP in components
Components SHALL NOT call `axios`/`fetch` directly for workflow operations; all data
access SHALL go through the react-query hooks.

#### Scenario: Component uses hooks
- **WHEN** a State Builder component needs workflow data
- **THEN** it uses the workflow hooks, never raw axios/fetch calls.
